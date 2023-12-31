package tx

import (
	"time"

	"github.com/btcsuite/btcd/btcutil"
)

type Tx struct {
	Hash   string    `json:"hash"`
	Time   time.Time `json:"time"`
	Size   uint32    `json:"size"`
	Weight uint32    `json:"weight"`
	Fee    uint64    `json:"fee"`
	// FeeKb     uint64    `json:"fee_kb"`
	// FeeByte   uint64    `json:"fee_b"`
	AmountOut uint64 `json:"amount_out"`
	AmountIn  uint64 `json:"amount_in"`
	Fits      bool   `json:"fits"`
}

func (t *Tx) FeePerKb() uint {
	return uint(float64(t.Fee) / float64(t.Size) * 1000)
}

func (t *Tx) FeePerByte() uint {
	return uint(float64(t.Fee) / float64(t.Size))
}

func (t *Tx) FeeString() string {
	return btcutil.Amount(t.Fee).String()
}
