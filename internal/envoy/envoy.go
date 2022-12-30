package envoy

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"

	"time"

	logger "github.com/raoulh/go-envoy/internal/log"
	"github.com/sirupsen/logrus"
)

var (
	logging *logrus.Entry
)

func init() {
	logging = logger.NewLogger("envoy")
}

func SetLoggerLevel(l logrus.Level) {
	logging.Logger.SetLevel(l)
}

type Envoy struct {
	Host             string `json:"host"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	EnvoySerial      string `json:"serial"`
	JWTToken         string `json:"jwt_token"`
	ManagerSessionId string `json:"-"`
	LocalSessionId   string `json:"-"`

	client *http.Client `json:"-"`
}

const (
	kEnlightenLoginUrl  = "https://enlighten.enphaseenergy.com/login/login.json"
	kEnlightenTokenUrl  = "https://enlighten.enphaseenergy.com/entrez-auth-token?serial_num=%s"
	kEnvoyCheckTokenUrl = "https://%s/auth/check_jwt"
	kEnvoyProductionUrl = "https://%s/production.json?details=1"
)

func New() *Envoy {
	e := Envoy{}
	e.loadFromCache()

	e.client = newClient()

	return &e
}

func SetConfig(host, user, pass, serial string) {
	e := Envoy{}

	if host == "" {
		host, _ = Discover()
		logging.Debugln("Found envoy host:", host)
	}

	e.Host = host
	e.Username = user
	e.Password = pass
	e.EnvoySerial = serial

	e.saveToCache()
}

func (e *Envoy) Rediscover() error {
	var err error
	e.Host, err = Discover()
	return err
}

func (e *Envoy) Close() {
	e.saveToCache()
}

func (e *Envoy) loadFromCache() {
	path := getCachePath()

	b, err := os.ReadFile(fmt.Sprintf("%s/envoy.cache", path))
	if err != nil {
		return
	}

	err = json.Unmarshal(b, e)
	if err != nil {
		logging.Debugf("unmarshal cache file failed: %s", err)
	}
}

func (e *Envoy) saveToCache() {
	path := getCachePath()

	b, err := json.Marshal(e)
	if err != nil {
		logging.Debugf("marshal cache file failed: %s", err)
	}

	err = os.WriteFile(fmt.Sprintf("%s/envoy.cache", path), b, os.ModePerm)
	if err != nil {
		return
	}
}

func getCachePath() string {
	path := os.Getenv("ENVOY_CACHE_PATH")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		if home != "" {
			path = fmt.Sprintf("%s/.cache/envoy", home)
		}
	}

	if path == "" {
		path = "./"
	}

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		logging.Errorln(err)
	}

	return path
}

func (e *Envoy) Login() (err error) {
	logging.Debug("Login")

	u := kEnlightenLoginUrl

	v := url.Values{}
	v.Add("user[email]", e.Username)
	v.Add("user[password]", e.Password)
	encodedData := v.Encode()

	//encodedData := fmt.Sprintf("user[email]=%s&user[password]=%s", e.Username, e.Password)

	req, err := http.NewRequest("POST", u, strings.NewReader(encodedData))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedData)))

	resp, err := e.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var d loginManagerToken
	err = json.Unmarshal(body, &d)
	if err != nil {
		logging.Debugf("login failure:\n%s", body)
		return
	}

	if d.Message == "success" {
		e.ManagerSessionId = d.SessionId
	} else {
		return fmt.Errorf("login on enlighten failed")
	}

	return
}

func (e *Envoy) GetToken() (err error) {
	logging.Debugf("GetToken")

	u := fmt.Sprintf(kEnlightenTokenUrl, e.EnvoySerial)
	uri, err := url.Parse(u)
	if err != nil {
		return
	}

	e.client.Jar.SetCookies(uri, []*http.Cookie{
		{
			Name:  "_enlighten_4_session",
			Value: e.ManagerSessionId,
		},
	})

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var d loginToken
	err = json.Unmarshal(body, &d)
	if err != nil {
		return
	}

	e.JWTToken = d.Token

	return
}

func (e *Envoy) GetLocalSessionCookie() (err error) {
	logging.Debugf("GetLocalSessionCookie")

	u := fmt.Sprintf(kEnvoyCheckTokenUrl, e.Host)
	uri, err := url.Parse(u)
	if err != nil {
		return
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+e.JWTToken)

	resp, err := e.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode > 400 {
		return fmt.Errorf("auth required")
	}

	if strings.Contains(string(body), "Valid token") {
		for _, c := range e.client.Jar.Cookies(uri) {
			if c.Name == "sessionId" {
				e.LocalSessionId = c.Value
			}
		}
	}

	return
}

func (e *Envoy) Production() (*production, error) {
	u := fmt.Sprintf(kEnvoyProductionUrl, e.Host)

	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	e.client.Jar.SetCookies(uri, []*http.Cookie{
		{
			Name:  "sessionId",
			Value: e.LocalSessionId,
		},
	})

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d production
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (e *Envoy) Home() (*home, error) {
	u := fmt.Sprintf("http://%s/home.json", e.Host)

	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	e.client.Jar.SetCookies(uri, []*http.Cookie{
		{
			Name:  "sessionId",
			Value: e.LocalSessionId,
		},
	})

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d home
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// http://envoy.local/inventory.json?deleted=1
func (e *Envoy) Inventory() (*[]inventory, error) {
	u := fmt.Sprintf("http://%s/inventory.json", e.Host)

	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	e.client.Jar.SetCookies(uri, []*http.Cookie{
		{
			Name:  "sessionId",
			Value: e.LocalSessionId,
		},
	})

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d []inventory
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (e *Envoy) Info() (*EnvoyInfo, error) {
	u := fmt.Sprintf("http://%s/info.xml", e.Host)

	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	e.client.Jar.SetCookies(uri, []*http.Cookie{
		{
			Name:  "sessionId",
			Value: e.LocalSessionId,
		},
	})

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var i EnvoyInfo
	err = xml.Unmarshal(body, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (e *Envoy) Now() (float64, float64, float64, error) {
	s, err := e.Production()
	if err != nil {
		return 0.0, 0.0, 0.0, err
	}
	tp := 0.0
	for _, v := range s.Production {
		if v.MeasurementType == "production" {
			tp = v.WNow
		}
	}
	tc := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "total-consumption" {
			tc = v.WNow
		}
	}
	net := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "net-consumption" {
			net = v.WNow
		}
	}
	return tp, tc, net, nil
}

func (e *Envoy) Today() (float64, float64, float64, error) {
	s, err := e.Production()
	if err != nil {
		return 0.0, 0.0, 0.0, err
	}
	tp := 0.0
	for _, v := range s.Production {
		if v.MeasurementType == "production" {
			tp = v.WhToday
		}
	}
	tc := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "total-consumption" {
			tc = v.WhToday
		}
	}
	tnp := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "net-consumption" {
			tnp = v.WhToday
		}
	}
	return tp, tc, tnp, nil
}

func (e *Envoy) Inverters() (*[]Inverter, error) {
	u := fmt.Sprintf("http://%s/api/v1/production/inverters", e.Host)

	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	e.client.Jar.SetCookies(uri, []*http.Cookie{
		{
			Name:  "sessionId",
			Value: e.LocalSessionId,
		},
	})

	resp, err := e.client.Get(u)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var i []Inverter
	err = json.Unmarshal(body, &i)
	if err != nil {
		logging.Debugf(string(body))
		return nil, err
	}
	return &i, nil
}

func (e *Envoy) SystemMax() (uint64, error) {
	inverters, err := e.Inverters()
	if err != nil {
		return 0, err
	}
	var max uint64
	for _, v := range *inverters {
		max += uint64(v.MaxReportWatts)
	}
	return max, nil
}

func newClient() *http.Client {
	tr := &http.Transport{
		ResponseHeaderTimeout: 3 * time.Second,
		DisableKeepAlives:     true,
		MaxIdleConns:          5,
		IdleConnTimeout:       20 * time.Second,
		DisableCompression:    true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		logging.Debug(err)
	}

	client := &http.Client{
		Transport: tr,
		Jar:       jar,
	}
	return client
}
