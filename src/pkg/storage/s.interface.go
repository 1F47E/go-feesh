package storage

import (
	mtx "go-btc-scan/src/pkg/entity/models/tx"
)

type PoolRepository interface {
	TxGet(txid string) (*mtx.Tx, error)
	TxAdd(tx mtx.Tx) error
}
