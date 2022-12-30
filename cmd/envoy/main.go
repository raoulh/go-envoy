package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/raoulh/go-envoy/internal/envoy"
	logger "github.com/raoulh/go-envoy/internal/log"
	"github.com/sirupsen/logrus"

	"github.com/fatih/color"
	cli "github.com/jawher/mow.cli"
)

const (
	CharStar     = "\u2737"
	CharAbort    = "\u2718"
	CharCheck    = "\u2714"
	CharWarning  = "\u26A0"
	CharArrow    = "\u2012\u25b6"
	CharVertLine = "\u2502"
	CharElec     = "🔌"
)

var (
	blue       = color.New(color.FgBlue).SprintFunc()
	errorRed   = color.New(color.FgRed).SprintFunc()
	errorBgRed = color.New(color.BgRed, color.FgBlack).SprintFunc()
	green      = color.New(color.FgGreen).SprintFunc()
	cyan       = color.New(color.FgCyan).SprintFunc()
	bgCyan     = color.New(color.FgWhite).SprintFunc()

	verbose *bool
)

func exit(err error, exit int) {
	log.Fatalln(errorRed(CharAbort), err)
	cli.Exit(exit)
}

func main() {
	app := cli.App("envoy", "Envoy CLI App")

	app.Spec = "[-v]"

	verbose = app.BoolOpt("v verbose", false, "Verbose debug mode")

	app.Before = func() {
		if *verbose {
			logger.SetFilterFormater(logger.NewCustomFormatter(false, logrus.TraceLevel))
		} else {
			logger.SetFilterFormater(logger.NewCustomFormatter(false, logrus.InfoLevel))
		}
	}

	app.Command("config", "manage account", func(config *cli.Cmd) {
		config.Command("set", "set account settings", func(setCmd *cli.Cmd) {
			setCmd.Spec = "[-h=<host>] [-u=<username>] [-p=<password>] [-s=<gateway_serial>]"

			var (
				host     = setCmd.StringOpt("h host", "", "Envoy Gateway IP hostname")
				username = setCmd.StringOpt("u username", "", "Envoy username")
				password = setCmd.StringOpt("p password", "", "Envoy password")
				serial   = setCmd.StringOpt("s serial", "", "Envoy Gateway serial")
			)

			setCmd.Action = func() {
				envoy.SetConfig(*host, *username, *password, *serial)
			}
		})
	})

	app.Command("now", "display current production", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			p, c, net, err := e.Now()
			if err != nil {
				fmt.Println("Failed to get current readings")
				exit(err, 1)
			}
			max, err := e.SystemMax()
			if err != nil {
				fmt.Printf("Failed to get system max prod: %v\n", err)
			}

			netimport := fmt.Sprintf("%2.2fW", net)
			if net > 0 {
				netimport = errorRed(netimport)
			} else {
				netimport = green(netimport)
			}

			fmt.Printf(CharElec+cyan("Production:")+" %2.2fW / %dW\t"+cyan("Consumption:")+" %2.2fW\tNet import: %s\n", p, max, c, netimport)
		}
	})

	app.Command("today", "display stats for today production", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			p, c, net, err := e.Today()
			if err != nil {
				fmt.Println("Failed to get today readings")
				exit(err, 1)
			}
			fmt.Printf(CharElec+cyan("Production:")+" %2.2fkWh\t"+cyan("Consumption:")+" %2.2fkWh\tNet: "+green("%2.2fkWh")+"\n", p/1000, c/1000, net/1000)
		}
	})

	app.Command("info", "display info about gateway", func(cmd *cli.Cmd) {
		cmd.Spec = "[-j]"

		var j = cmd.BoolOpt("j json", false, "JSON output mode")

		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			s, err := e.Info()
			if err != nil {
				fmt.Println("Failed to get gateway info")
				exit(err, 1)
			}

			if *j {
				i, _ := json.MarshalIndent(s, "", " ")
				fmt.Printf("%s\n", i)
			} else {
				fmt.Println("Serial Number: ", s.Device.Sn)
				fmt.Println("Part Number: ", s.Device.Pn)
				fmt.Println("Software Version: ", s.Device.Software)
			}
		}
	})

	app.Command("production", "display raw json production", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			p, err := e.Production()
			if err != nil {
				fmt.Println("Failed to get prod info")
				exit(err, 1)
			}

			i, _ := json.MarshalIndent(p, "", " ")
			fmt.Printf("%s\n", i)
		}
	})

	app.Command("inventory", "display raw json inventory", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			p, err := e.Inventory()
			if err != nil {
				fmt.Println("Failed to get prod info")
				exit(err, 1)
			}

			i, _ := json.MarshalIndent(p, "", " ")
			fmt.Printf("%s\n", i)
		}
	})

	app.Command("inverters", "display raw json inverters", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			p, err := e.Inverters()
			if err != nil {
				fmt.Println("Failed to get prod info")
				exit(err, 1)
			}

			i, _ := json.MarshalIndent(p, "", " ")
			fmt.Printf("%s\n", i)
		}
	})

	app.Command("home", "display raw json /home.json", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			e, err := tryLogin()
			if err != nil {
				fmt.Println("Failed to login")
				exit(err, 1)
			}
			defer e.Close()

			p, err := e.Home()
			if err != nil {
				fmt.Println("Failed to get prod info")
				exit(err, 1)
			}

			i, _ := json.MarshalIndent(p, "", " ")
			fmt.Printf("%s\n", i)
		}
	})

	if err := app.Run(os.Args); err != nil {
		exit(err, 1)
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
