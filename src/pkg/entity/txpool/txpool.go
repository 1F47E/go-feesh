package txpool

import "fmt"

type TxPool struct {
	Hash      string `json:"hash"`
	Block     string `json:"block"`
	Size      int    `json:"size"`
	AmountIn  uint64 `json:"amount_in"`
	AmountOut uint64 `json:"amount_out"`
}

func (t *TxPool) Fee() uint64 {
	return t.AmountIn - t.AmountOut
}

func (t *TxPool) FeePerByte() string {
	feeF := float64(t.Fee()) / float64(t.Size)
	return fmt.Sprintf("%.1f", feeF)
}
