package api

import (
	"net/http"

	mtx "github.com/1F47E/go-feesh/pkg/entity/models/tx"
	"github.com/1F47E/go-feesh/pkg/logger"

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

// @Summary Get pool information
// @Description Get information about the current state of the pool
// @Tags pool
// @Accept  json
// @Produce  json
// @Param limit query int false "Limit the number of transactions returned"
// @Success 200 {object} PoolResponse
// @Failure 500 {object} APIError
// @Router /pool [get]
func (a *Api) Pool(c *fiber.Ctx) error {
	log := c.Locals("logger").(logger.LoggerEntry)

	limit := c.QueryInt("limit", 100)
	txs, err := a.core.GetPool(limit)
	if err != nil {
		log.Errorf("error on getpool: %v\n", err)
		return apiError(c, http.StatusInternalServerError, "Something went wrong", err.Error())
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
	log.Infof("pool size: %d\n", ret.Size)
	return apiSuccess(c, ret)
}