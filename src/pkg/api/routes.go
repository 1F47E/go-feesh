package api

import (
	mtx "go-btc-scan/src/pkg/entity/models/tx"

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

type PoolResponse struct {
	Height int       `json:"height"`
	Size   int       `json:"size"`
	Txs    []*mtx.Tx `json:"txs"`
}

func (a *Api) Pool(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	txs := a.core.GetPoolTxsRecent(limit)
	ret := PoolResponse{
		Height: a.core.GetPoolHeight(),
		Size:   a.core.GetPoolSize(),
		Txs:    txs,
	}
	return c.JSON(ret)
}
