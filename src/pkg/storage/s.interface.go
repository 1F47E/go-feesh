package storage

import (
	mtx "go-btc-scan/src/pkg/entity/models/tx"
)

type PoolRepository interface {
	TxGet(txid string) (*mtx.Tx, error)
	TxAdd(tx mtx.Tx) error
	BlockExists(hash string) (bool, error)
	BlockGet(hash string) ([]string, error)
	BlockAdd(hash string, txs []string) error
}
