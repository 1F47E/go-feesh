package storage

import "go-btc-scan/src/pkg/entity/models/tx"

type PoolRepository interface {
	// pool stuff
	PoolAddTx(tx *tx.Tx) error
	PoolGetTx(txid string) (*tx.Tx, error)
	// PoolGet() (map[string]*tx.Tx, error)
	PoolGetCache() (map[string]struct{}, error)
	PoolSize() uint
	PoolReset() error

	// tx ordered list
	PoolListUpdate(txs []string) error
	PoolListGet() ([]*tx.Tx, error)

	// tx cache
	TxGet(txid string) (*tx.Tx, error)
	TxAdd(tx *tx.Tx) error
}
