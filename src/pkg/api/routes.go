package api

import (
	fiber "github.com/gofiber/fiber/v2"
)

func (a *Api) Ping(c *fiber.Ctx) error {
	return c.SendString("pong")
}
