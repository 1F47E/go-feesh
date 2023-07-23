package core

import (
	"context"
	"go-btc-scan/src/pkg/client"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	"log"
	"sync"
	"time"
)

type Core struct {
	ctx  context.Context
	mu   *sync.Mutex
	cli  *client.Client
	pool map[string]*mtx.Tx
}

func NewCore(ctx context.Context, cli *client.Client) *Core {
	return &Core{
		ctx:  ctx,
		mu:   &sync.Mutex{},
		cli:  cli,
		pool: make(map[string]*mtx.Tx, 0),
	}
}

func (c *Core) GetParsedPool() []*mtx.Tx {
	c.mu.Lock()
	defer c.mu.Unlock()
	ret := make([]*mtx.Tx, 0)
	for _, tx := range c.pool {
		if tx.IsParsed() {
			ret = append(ret, tx)
		}
	}
	return ret
}

func (c *Core) Run() {
	go c.workerPool()

}

func (c *Core) workerPool() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// parse pool
			txs, err := c.cli.RawMempool()
			if err != nil {
				log.Fatalln("error on rawmempool:", err)
			}
			// add new tx to the pool
			c.mu.Lock()
			txsNew := make([]string, 0)
			for _, txid := range txs {
				if _, ok := c.pool[txid]; !ok {
					txsNew = append(txsNew, txid)
				}
			}
			if len(txsNew) > 0 {
				go c.parsePoolTx(txsNew)
			}
			c.mu.Unlock()

		}
	}
}

func (c *Core) parsePoolTx(txs []string) {
	c.mu.Lock()
	for _, txid := range txs {
		tx := &mtx.Tx{Hash: txid}

		if tx.AmountIn == 0 || tx.AmountOut == 0 {
			// request
			in, out := c.cli.TransactionCalculate(tx.Hash)
			tx.AmountIn = in
			tx.AmountOut = out
		}
		c.pool[txid] = tx
	}
	c.mu.Unlock()
	// TODO: push new pool tx list to web socket
}
