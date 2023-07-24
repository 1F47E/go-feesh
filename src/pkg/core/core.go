package core

import (
	"context"
	"go-btc-scan/src/pkg/client"
	log "go-btc-scan/src/pkg/logger"
	"go-btc-scan/src/pkg/storage"
	"math/rand"

	"go-btc-scan/src/pkg/entity/btc/info"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	"sync"
	"time"
)

type Core struct {
	ctx    context.Context
	mu     *sync.Mutex
	cli    *client.Client
	height int
	// pool *pool.Pool
	storage storage.PoolRepository
	// blocks      []*mblock.Block
	poolTxCh    chan *mtx.Tx
	poolTxResCh chan *mtx.Tx
}

func NewCore(ctx context.Context, cli *client.Client, s storage.PoolRepository) *Core {
	return &Core{
		ctx: ctx,
		mu:  &sync.Mutex{},
		cli: cli,
		// pool: pool.NewPool(s),
		storage: s,
		// blocks:      make([]*mblock.Block, 0),
		poolTxCh:    make(chan *mtx.Tx),
		poolTxResCh: make(chan *mtx.Tx),
	}
}

func (c *Core) Start() {
	// set the pool block height
	info, err := c.cli.GetInfo()
	if err != nil {
		log.Log.Errorf("error on getinfo: %v\n", err)
	} else {
		// even if its fails - having block 0 will update pool txs list every time
		// its just for performance reasons
		c.height = info.Blocks
	}
	go c.workerGetMemPool()
	// make a batch of parsers
	for i := 0; i < 420; i++ {
		go c.workerTxParser()
	}

	go c.workerTxAdder()
}

func (c *Core) GetNodeInfo() (*info.Info, error) {
	return c.cli.GetInfo()
}

// parse last N blocks
func (c *Core) bootstrap() {
	// TODO: bootstap blocks
	// get best block
	// download header, get prev block, repeat N times
	// parse every block
}

func (c *Core) workerTxParser() {
	name := "[workerTxParser]"
	log.Log.Tracef("%s started\n", name)
	defer func() {
		log.Log.Debugf("%s stopped\n", name)
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case tx := <-c.poolTxCh:
			var err error

			// get tx from cache
			stx, err := c.storage.TxGet(tx.Hash)
			if err != nil {
				log.Log.Errorf("%s error on storage.TxGet: %v\n", name, err)
				continue
			}
			if stx != nil {
				c.poolTxResCh <- stx
				continue
			}

			// log.Log.Debugf("[%s] got tx: %s\n", name, tx.Hash)

			// parse tx with retry
			max := 10
			txFull := new(mtx.Tx)
			for i := 0; i <= max; i++ {
				// sleep randomly
				time.Sleep(time.Duration(300*i+rand.Intn(300)) * time.Millisecond)
				txFull, err = c.parsePoolTx(tx)
				if err != nil {
					if err.Error() == client.ERR_5xx {
						log.Log.Errorf("%s error 5xx, retrying %d/%d\n", i+1, name, max)
						continue
					}
					log.Log.Errorf("%s error on parsePoolTx: %v\n", name, err)
					continue
				}
				break
			}
			// log.Log.Debugf("[%s] parsed tx: %s\n", name, tx.Hash)

			// save tx
			err = c.storage.TxAdd(txFull)
			if err != nil {
				log.Log.Errorf("%s error on storage.TxAdd: %v\n", name, err)
			}
			c.poolTxResCh <- txFull
			// log.Log.Debugf("[%s] sent tx: %s\n", name, tx.Hash)
		}
	}
}

func (c *Core) workerTxAdder() {
	name := "workerTxAdder"
	log.Log.Infof("[%s] started\n", name)
	defer func() {
		log.Log.Infof("[%s] stopped\n", name)
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case tx := <-c.poolTxResCh:
			// log.Log.Debugf("[%s] got tx: %s\n", name, tx.Hash)
			// _ = c.storage.PoolAdd(tx)
			// c.mu.Lock()
			c.storage.PoolAddTx(tx)
			// c.mu.Unlock()
			// log.Log.Debugf("[%s] added tx: %s\npool size: %d\n", name, tx.Hash, c.pool.Size())
		}
	}
}

func (c *Core) workerGetMemPool() {
	name := "workerGetMemPool"
	log.Log.Infof("[%s] started\n", name)
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		log.Log.Infof("[%s] stopped\n", name)
		ticker.Stop()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// check the block height
			info, err := c.cli.GetInfo()
			if err != nil {
				log.Log.Errorf("error on getinfo: %v\n", err)
				continue
			}
			log.Log.Debugf("block height: %d\n", info.Blocks)

			// get ordered list of pool tsx. new first
			poolTxs, err := c.cli.RawMempool()
			if err != nil {
				log.Log.Errorf("error on rawmempool: %v\n", err)
				continue
			}
			log.Log.Debugf("pool txs: %d\n", len(poolTxs))

			// update pool list []string
			poolList := make([]string, len(poolTxs))
			for i, tx := range poolTxs {
				poolList[i] = tx.Hash
			}
			c.storage.PoolListUpdate(poolList)

			// reset pool if new block
			if c.height != info.Blocks {
				log.Log.Debugf("reset pool block height: %d\n", info.Blocks)
				_ = c.storage.PoolReset()
				c.height = info.Blocks
			}

			// ramap btc tx to model tx and send them to additional parsing
			// copy cache to avoid locks
			cache, err := c.storage.PoolGetCache()
			if err != nil {
				log.Log.Errorf("error on storage.PoolGet: %v\n", err)
				cache = make(map[string]struct{})
			}

			// TODO: optimize later - do not parse txs that are already in pool
			for _, tx := range poolTxs {
				// skip if already in pool
				if _, ok := cache[tx.Hash]; ok {
					continue
				}

				// remap to tx model
				mt := &mtx.Tx{
					Hash:   tx.Hash,
					Time:   time.Unix(tx.Time, 0),
					Size:   tx.Size,
					Vsize:  tx.Vsize,
					Weight: tx.Weight,
					Fee:    tx.Fee,
					FeeKb:  tx.FeePerKB,
				}
				c.poolTxCh <- mt
			}
			// log storage size
			log.Log.Debugf("storage size: %d\n", c.storage.PoolSize())
		}
	}
}

func (c *Core) parsePoolTx(tx *mtx.Tx) (*mtx.Tx, error) {
	btx, err := c.cli.TransactionGet(tx.Hash)
	if err != nil {
		return nil, err
	}
	// update amounts
	tx.AmountOut = btx.GetTotalOut()

	return tx, nil
}

// pool access from API

func (c *Core) GetPoolTxs() ([]*mtx.Tx, error) {
	return c.storage.PoolListGet()
}

func (c *Core) GetPoolSize() uint {
	return c.storage.PoolSize()
}

func (c *Core) GetPoolTxsRecent(limit int) ([]*mtx.Tx, error) {
	allTxs, err := c.storage.PoolListGet()
	if err != nil {
		return nil, err
	}
	if len(allTxs) <= limit {
		return allTxs, nil
	}
	return allTxs[:limit], nil
}

func (c *Core) GetPoolHeight() int {
	return c.height
}

// TODO: cache this
func (c *Core) GetTotalAmount() (uint64, error) {
	var total uint64
	poolTxs, err := c.storage.PoolListGet()
	if err != nil {
		return 0, err
	}
	for _, tx := range poolTxs {
		total += tx.AmountOut
	}
	return total, nil
}
