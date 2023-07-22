package block

import "github.com/btcsuite/btcd/btcutil"

type Block struct {
	Hash   string `json:"hash"`
	Height int    `json:"height"`
	Value  uint64 `json:"value"`
	Fee    uint64 `json:"fee"`
}

func (b *Block) ValueString() string {
	return btcutil.Amount(b.Value).String()
}

func (b *Block) FeeString() string {
	return btcutil.Amount(b.Fee).String()
}

func (b *Block) IsComplete() bool {
	return b.Value != 0 && b.Fee != 0
}
