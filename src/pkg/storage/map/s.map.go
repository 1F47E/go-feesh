package storage_map

import (
	"go-btc-scan/src/pkg/entity/models/tx"
	"sync"
)

type MapStorage struct {
	mu       *sync.Mutex
	txs      map[string]*tx.Tx
	pool     map[string]*tx.Tx
	poolList []string
}

func New() *MapStorage {
	return &MapStorage{
		mu:       &sync.Mutex{},
		txs:      make(map[string]*tx.Tx),
		pool:     make(map[string]*tx.Tx),
		poolList: make([]string, 0),
	}
}

func (m *MapStorage) PoolGetTx(txid string) (*tx.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool[txid], nil
}

func (m *MapStorage) PoolAddTx(tx *tx.Tx) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pool[tx.Hash] = tx
	return nil
}

func (m *MapStorage) GetPool() (map[string]*tx.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool, nil
}

func (m *MapStorage) PoolGetCache() (map[string]struct{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cache := make(map[string]struct{}, len(m.pool))
	for _, tx := range m.pool {
		cache[tx.Hash] = struct{}{}
	}
	return cache, nil
}

func (m *MapStorage) PoolSize() uint {
	return uint(len(m.pool))
}

func (m *MapStorage) PoolReset() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pool = make(map[string]*tx.Tx)
	return nil
}

func (m *MapStorage) PoolListUpdate(txs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.poolList = txs
	return nil
}

func (m *MapStorage) PoolListGet() ([]*tx.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := make([]*tx.Tx, len(m.poolList))
	for i, txid := range m.poolList {
		list[i] = m.txs[txid]
	}
	return list, nil
}

func (m *MapStorage) TxGet(txid string) (*tx.Tx, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.txs[txid], nil
}

func (m *MapStorage) TxAdd(tx *tx.Tx) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs[tx.Hash] = tx
	return nil
}

func (m *MapStorage) Size() uint {
	return uint(len(m.txs))
}
