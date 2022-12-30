package app

import (
	"github.com/raoulh/go-envoy/internal/envoy"
)

var (
	production envoy.Production
	inventory  []envoy.Inventory
	inverters  []envoy.Inverter
)

func (a *AppServer) doDataRead() {
	e, err := tryLogin()
	if err != nil {
		logging.Error("Failed to login")
		return
	}
	defer e.Close()

	prod, err := e.Production()
	if err != nil {
		logging.Error("Failed to get prod info")
	} else {
		production = *prod
	}

	inv, err := e.Inventory()
	if err != nil {
		logging.Error("Failed to get inventory info")
	} else {
		inventory = *inv
	}

	inver, err := e.Inverters()
	if err != nil {
		logging.Error("Failed to get inverters info")
	} else {
		inverters = *inver
	}
}

func tryLogin() (e *envoy.Envoy, err error) {
	e = envoy.New()

	//No token yet, try to login
	if e.JWTToken == "" {
		err = e.Login()
		if err != nil {
			return
		}

		err = e.GetToken()
		if err != nil {
			return
		}
	}

	//Try to get the local cookie
	err = e.GetLocalSessionCookie()
	if err != nil {
		//The cookie is not valid anymore, login again
		err = e.Login()
		if err != nil {
			return
		}

		err = e.GetToken()
		if err != nil {
			return
		}

		err = e.GetLocalSessionCookie()
		if err != nil {
			return
		}
	}

	return
}
