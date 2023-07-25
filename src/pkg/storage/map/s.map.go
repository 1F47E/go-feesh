package storage_map

import (
	"go-btc-scan/src/pkg/entity/models/tx"
	"sync"
)

type MapStorage struct {
	mu  *sync.Mutex
	txs map[string]*tx.Tx
}

func New() *MapStorage {
	return &MapStorage{
		mu:  &sync.Mutex{},
		txs: make(map[string]*tx.Tx),
	}
}

func (m *MapStorage) TxGet(txid string) (*tx.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.txs[txid], nil
}

func (m *MapStorage) TxAdd(tx tx.Tx) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs[tx.Hash] = &tx
	return nil
}
