package pool

import (
	"go-btc-scan/src/pkg/entity/models/tx"
)

// mutex should be locked outside
// because this struct is used in a loop a lot
type Pool struct {
	txs         []*tx.Tx
	cache       map[string]struct{}
	BlockHeight int
}

func NewPool() *Pool {
	return &Pool{
		txs:   make([]*tx.Tx, 0),
		cache: make(map[string]struct{}, 0),
	}
}

func (p *Pool) Len() int {
	return len(p.txs)
}

// check txid in a cache
func (p *Pool) HasTx(txid string) bool {
	_, ok := p.cache[txid]
	return ok
}

func (p *Pool) AddTx(tx *tx.Tx) {
	p.txs = append(p.txs, tx)
	p.cache[tx.Hash] = struct{}{}
}

// on new block - reset the pool
func (p *Pool) Reset(height int) {
	p.BlockHeight = height
	p.txs = make([]*tx.Tx, 0)
	p.cache = make(map[string]struct{}, 0)
}

// get all txs
func (p *Pool) GetTxs() []*tx.Tx {
	return p.txs
}

func (p *Pool) Size() int {
	return len(p.txs)
}

// get recent N txs from pool sorted by time DESC
func (p *Pool) GetTxsRecent(limit int) []*tx.Tx {
	if len(p.txs) == 0 {
		return p.txs
	}
	if limit > len(p.txs) {
		limit = len(p.txs)
	}
	recent := p.txs[len(p.txs)-limit:]
	recentOrdered := make([]*tx.Tx, len(recent))
	for i, tx := range recent {
		recentOrdered[len(recent)-i-1] = tx
	}
	return recentOrdered
}
