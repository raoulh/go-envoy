package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/raoulh/go-envoy"
)

func main() {
	var command string

	flag.Parse()
	args := flag.Args()
	argc := len(args)
	if argc == 0 {
		command = "unset"
	}
	if argc >= 1 {
		command = args[0]
	}

	e := envoy.New("192.168.0.134", "raoul-pubs@calaos.fr", "xxxxxx", "122233103807")

	if e.JWTToken == "" {
		err := e.Login()
		if err != nil {
			panic(err)
		}

		err = e.GetToken()
		if err != nil {
			panic(err)
		}
	}

	err := e.GetLocalSessionCookie()
	if err != nil {
		err := e.Login()
		if err != nil {
			panic(err)
		}

		err = e.GetToken()
		if err != nil {
			panic(err)
		}

		err = e.GetLocalSessionCookie()
		if err != nil {
			panic(err)
		}
	}

	switch command {
	case "prod":
		s, err := e.Production()
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
	case "now":
		p, c, net, err := e.Now()
		if err != nil {
			panic(err)
		}
		max, err := e.SystemMax()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Production: %2.2fW / %dW\tConsumption: %2.2fW\tNet: %2.2fW\n", p, max, c, net)
	case "today":
		p, c, net, err := e.Today()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Production: %2.2fkWh\tConsumption: %2.2fkWh\tNet: %2.2fkWh\n", p/1000, c/1000, net/1000)
	case "home":
		s, err := e.Home()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", s)
	case "inventory":
		s, err := e.Inventory()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%+v\n", s)
	case "info":
		s, err := e.Info()
		if err != nil {
			panic(err)
		}
		// fmt.Printf("%+v\n", s)
		fmt.Println("Serial Number: ", s.Device.Sn)
		fmt.Println("Part Number: ", s.Device.Pn)
		fmt.Println("Software Version: ", s.Device.Software)
	case "stream":
		/* s, err := e.Home()
		if err != nil {
			panic(err)
		} */
		fmt.Printf("working on it...\n")
	default:
		fmt.Println("usage: envoy <command> <IQ IP address/hostname>")
		fmt.Println("Valid commands: prod, home, inventory, stream, now, today, info")
	}

	e.Close()
}
