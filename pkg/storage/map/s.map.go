package storage_map

import (
	"sync"

	"github.com/1F47E/go-feesh/pkg/entity/models/tx"
)

type MapStorage struct {
	mu     *sync.Mutex
	txs    map[string]*tx.Tx
	blocks map[string][]string
}

func New() *MapStorage {
	return &MapStorage{
		mu:     &sync.Mutex{},
		txs:    make(map[string]*tx.Tx),
		blocks: make(map[string][]string),
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

func (m *MapStorage) BlockExists(hash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.blocks[hash]
	return ok, nil
}

func (m *MapStorage) BlockGet(hash string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.blocks[hash], nil
}

func (m *MapStorage) BlockAdd(hash string, txs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks[hash] = txs
	return nil
}
