package api

import (
	mtx "go-btc-scan/src/pkg/entity/models/tx"

	fiber "github.com/gofiber/fiber/v2"
)

type PoolResponse struct {
	Height     int           `json:"height"`
	Size       int           `json:"size"`
	Amount     uint64        `json:"amount"`
	Weight     uint64        `json:"weight"`
	Fee        uint64        `json:"fee"`
	FeeBuckets map[uint]uint `json:"fee_buckets"`
	Txs        []mtx.Tx      `json:"txs"`
	Blocks     []string      `json:"blocks"`
}

func (a *Api) Pool(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	txs, err := a.core.GetPool(limit)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	ret := PoolResponse{
		Height:     a.core.GetHeight(),
		Size:       a.core.GetPoolSize(),
		Amount:     a.core.GetTotalAmount(),
		Weight:     a.core.GetTotalWeight(),
		Fee:        a.core.GetTotalFee(),
		FeeBuckets: a.core.GetFeeBuckets(),
		Txs:        txs,
		Blocks:     a.core.GetBlocks(),
	}
	return c.JSON(ret)
}
