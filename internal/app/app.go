package app

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberLog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html"

	"github.com/raoulh/go-envoy/internal/config"
	logger "github.com/raoulh/go-envoy/internal/log"
	"github.com/sirupsen/logrus"
)

const (
	maxFileSize  = 1 * 1024 * 1024 * 1024
	dataWaitTime = time.Second * 1
)

type AppServer struct {
	quitHeartbeat chan interface{}
	wgDone        sync.WaitGroup

	appFiber *fiber.App
}

var logging *logrus.Entry

func init() {
	logging = logger.NewLogger("app")
}

// Init the app
func NewApp() (a *AppServer, err error) {
	logging.Infoln("Init server")

	engine := html.New(config.Config.String("general.static")+"/templates", ".html")

	a = &AppServer{
		quitHeartbeat: make(chan interface{}),
		appFiber: fiber.New(fiber.Config{
			ServerHeader:          "Envoy (Linux)",
			ReadTimeout:           time.Second * 20,
			AppName:               "Envoy",
			DisableStartupMessage: true,
			EnablePrintRoutes:     false,
			BodyLimit:             maxFileSize,
			Views:                 engine,
		}),
	}

	a.appFiber.
		Use(fiberLog.New(fiberLog.Config{}))

	a.appFiber.Static("/js", config.Config.String("general.static")+"/js")
	a.appFiber.Static("/css", config.Config.String("general.static")+"/css")

	a.appFiber.Hooks().OnShutdown(func() error {
		a.wgDone.Done()
		return nil
	})

	a.appFiber.Get("/", func(c *fiber.Ctx) error {
		return a.homePage(c)
	})

	//API
	api := a.appFiber.Group("/api")
	api.Get("/production", func(c *fiber.Ctx) error {
		return a.apiProduction(c)
	})
	api.Get("/inventory", func(c *fiber.Ctx) error {
		return a.apiInventory(c)
	})
	api.Get("/inverters", func(c *fiber.Ctx) error {
		return a.apiInverters(c)
	})

	return
}

// Run the app
func (a *AppServer) Start() {
	addr := config.Config.String("general.address") + ":" + strconv.Itoa(config.Config.Int("general.port"))

	logging.Infoln("\u21D2 Server listening on", addr)

	go func() {
		if err := a.appFiber.Listen(addr); err != nil {
			logging.Fatalf("Failed to listen http server: %v", err)
		}
	}()
	a.wgDone.Add(1)

	go a.getDataFromGateway()
	a.wgDone.Add(1)
}

// Stop the app
func (a *AppServer) Shutdown() {
	close(a.quitHeartbeat)
	a.appFiber.Shutdown()
	a.wgDone.Wait()
}

func (a *AppServer) getDataFromGateway() {
	defer a.wgDone.Done()

	for {
		select {
		case <-a.quitHeartbeat:
			logging.Debugln("exiting data gather routine")
			return
		case <-time.After(dataWaitTime):
			a.doDataRead()
		}
	}
}
