package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (a *AppServer) homePage(c *fiber.Ctx) error {
	type Data struct {
		ProdNow  string
		ConsoNow string
		NetNow   string

		ProdToday  string
		ConsoToday string
		NetToday   string
	}

	d := Data{}

	for _, v := range production.Production {
		if v.MeasurementType == "production" {
			d.ProdNow = fmt.Sprintf("%2.0f", v.WNow)
		}
	}

	for _, v := range production.Consumption {
		if v.MeasurementType == "total-consumption" {
			d.ConsoNow = fmt.Sprintf("%2.0f", v.WNow)
		}
	}

	for _, v := range production.Consumption {
		if v.MeasurementType == "net-consumption" {
			d.NetNow = fmt.Sprintf("%2.0f", v.WNow)
		}
	}

	for _, v := range production.Production {
		if v.MeasurementType == "production" {
			d.ProdToday = fmt.Sprintf("%2.0f", v.WhToday/1000)
		}
	}

	for _, v := range production.Consumption {
		if v.MeasurementType == "total-consumption" {
			d.ConsoToday = fmt.Sprintf("%2.0f", v.WhToday/1000)
		}
	}

	for _, v := range production.Consumption {
		if v.MeasurementType == "net-consumption" {
			d.NetToday = fmt.Sprintf("%2.0f", v.WhToday/1000)
		}
	}

	return c.Render("index", &d)
}
