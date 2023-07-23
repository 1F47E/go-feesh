package pool

import (
	"go-btc-scan/src/pkg/entity/btc/block"
	"go-btc-scan/src/pkg/entity/models/tx"
	"sync"
)

type Pool struct {
	mu     *sync.Mutex
	txs    []*tx.Tx
	cache  map[string]struct{}
	blocks []*block.Block
}

func NewPool() *Pool {
	return &Pool{
		mu:     &sync.Mutex{},
		txs:    make([]*tx.Tx, 0),
		cache:  make(map[string]struct{}, 0),
		blocks: make([]*block.Block, 0),
	}
}

// check txid in a cache
func (p *Pool) HasTx(txid string) bool {
	// mutex should be locked outside
	// because this method is used in a loop
	_, ok := p.cache[txid]
	return ok
}

// add new tx
func (p *Pool) AddTx(tx *tx.Tx) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// txs should always be in order by time, new on top
	p.txs = append(p.txs, tx)
	p.cache[tx.Hash] = struct{}{}
}
