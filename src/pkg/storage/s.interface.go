package storage

import "go-btc-scan/src/pkg/entity/models/tx"

type TxRepository interface {
	TxGet(txid string) (*tx.Tx, error)
	TxAdd(tx *tx.Tx) error
	Size() uint
}
