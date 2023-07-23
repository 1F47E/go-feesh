package tx

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
)

type Tx struct {
	Hash      string `json:"hash"`
	Block     string `json:"block"`
	Size      int    `json:"size"`
	AmountIn  uint64 `json:"amount_in"`
	AmountOut uint64 `json:"amount_out"`
}

func (t *Tx) IsParsed() bool {
	return t.AmountIn != 0 && t.AmountOut != 0
}

func (t *Tx) Fee() uint64 {
	// if mined, no inputs
	if t.AmountIn == 0 {
		return 0
	}
	return t.AmountIn - t.AmountOut
}

// inBtc := btcutil.Amount(in)
// outBtc := btcutil.Amount(out)
func (t *Tx) InString() string {
	return btcutil.Amount(t.AmountIn).String()
}
func (t *Tx) OutString() string {
	return btcutil.Amount(t.AmountOut).String()
}
func (t *Tx) FeeString() string {
	return btcutil.Amount(t.Fee()).String()
}

func (t *Tx) FeePerByte() string {
	feeF := float64(t.Fee()) / float64(t.Size)
	return fmt.Sprintf("%.1f", feeF)
}
