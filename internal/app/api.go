package app

import "github.com/gofiber/fiber/v2"

func (a *AppServer) apiProduction(c *fiber.Ctx) error {
	return c.JSON(production)
}

func (a *AppServer) apiInventory(c *fiber.Ctx) error {
	return c.JSON(inventory)
}

func (a *AppServer) apiInverters(c *fiber.Ctx) error {
	return c.JSON(inverters)
}
