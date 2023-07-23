package core

import (
	"context"
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/core/pool"

	// "go-btc-scan/src/pkg/entity/btc/block"
	mblock "go-btc-scan/src/pkg/entity/models/block"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	"log"
	"sync"
	"time"
)

type Core struct {
	ctx context.Context
	mu  *sync.Mutex
	cli *client.Client
	// poolCache map[string]struct{}
	pool   *pool.Pool
	blocks []*mblock.Block
}

func NewCore(ctx context.Context, cli *client.Client) *Core {
	return &Core{
		ctx:    ctx,
		mu:     &sync.Mutex{},
		cli:    cli,
		pool:   pool.NewPool(),
		blocks: make([]*mblock.Block, 0),
	}
}

func (c *Core) Run() {
	go c.workerPool()

}

// parse last N blocks
func (c *Core) bootstrap() {
	// TODO: bootstap
	// get best block
	// download header, get prev block, repeat N times
	// parse every block
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
				if !c.pool.HasTx(txid) {
					txsNew = append(txsNew, txid)
				}
			}
			if len(txsNew) > 0 {
				go c.parsePoolTxs(txsNew)
			}
			c.mu.Unlock()
		}
	}
}

func (c *Core) parsePoolTxs(txs []string) {
	for _, txid := range txs {
		// tx := &tx.Tx{Hash: txid}
		rpcTx, err := c.cli.TransactionGet(txid)
		if err != nil {
			log.Printf("error on get tx %s: %s\n", txid, err)
			continue
		}
		// calc vin via parsing vin txid
		amountIn, err := c.cli.TransactionGetVin(rpcTx.Hash)
		if err != nil {
			log.Printf("error on get vin tx %s: %s\n", txid, err)
			continue
		}

		// construct tx model
		tx := &mtx.Tx{
			Hash:      txid,
			Time:      time.Unix(int64(rpcTx.Time), 0),
			AmountIn:  amountIn,
			AmountOut: rpcTx.GetTotalOut(),
		}
		if tx.AmountIn == 0 && tx.AmountOut == 0 {
			tx.Fee = tx.AmountIn - tx.AmountOut
		}

		c.pool.AddTx(tx)
	}
	// TODO: push new pool tx list to web socket
}
