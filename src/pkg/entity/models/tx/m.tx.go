package tx

import "fmt"

type Tx struct {
	Hash      string `json:"hash"`
	Block     string `json:"block"`
	Size      int    `json:"size"`
	AmountIn  uint64 `json:"amount_in"`
	AmountOut uint64 `json:"amount_out"`
}

func (t *Tx) Fee() uint64 {
	// if mined, no inputs
	if t.AmountIn == 0 {
		return 0
	}
	return t.AmountIn - t.AmountOut
}

func (t *Tx) FeePerByte() string {
	feeF := float64(t.Fee()) / float64(t.Size)
	return fmt.Sprintf("%.1f", feeF)
}
