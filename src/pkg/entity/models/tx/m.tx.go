package tx

import (
	"time"

	"github.com/btcsuite/btcd/btcutil"
)

type Tx struct {
	Hash      string    `json:"hash"`
	Time      time.Time `json:"time"`
	Size      uint64    `json:"size"`
	Vsize     uint64    `json:"vsize"`
	Weight    uint64    `json:"weight"`
	Fee       uint64    `json:"fee"`
	FeeKb     uint64    `json:"fee_kb"`
	AmountIn  uint64    `json:"amount_in"`
	AmountOut uint64    `json:"amount_out"`
}

func (t *Tx) FeeString() string {
	return btcutil.Amount(t.Fee).String()
}
