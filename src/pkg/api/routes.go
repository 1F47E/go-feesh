package api

import (
	fiber "github.com/gofiber/fiber/v2"
)

func (a *Api) Ping(c *fiber.Ctx) error {
	return c.SendString("pong")
}

func (a *Api) NodeInfo(c *fiber.Ctx) error {
	// txs := a.core.GetPoolTxs()
	info, err := a.core.GetNodeInfo()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	return c.JSON(info)
}
