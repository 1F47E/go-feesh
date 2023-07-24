package api

import (
	mtx "go-btc-scan/src/pkg/entity/models/tx"

	fiber "github.com/gofiber/fiber/v2"
)

type PoolResponse struct {
	Height int       `json:"height"`
	Size   uint      `json:"size"`
	Amount uint64    `json:"amount"`
	Txs    []*mtx.Tx `json:"txs"`
}

func (a *Api) Pool(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	txs, err := a.core.GetPoolTxsRecent(limit)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	amount, err := a.core.GetTotalAmount()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	ret := PoolResponse{
		Height: a.core.GetPoolHeight(),
		Size:   a.core.GetPoolSize(),
		Amount: amount,
		Txs:    txs,
	}
	return c.JSON(ret)
}
