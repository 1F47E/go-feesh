package core

import (
	"context"
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/config"
	log "go-btc-scan/src/pkg/logger"
	"go-btc-scan/src/pkg/storage"
	"time"

	"go-btc-scan/src/pkg/entity/btc/info"
	"go-btc-scan/src/pkg/entity/btc/txpool"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	"sync"
)

const BLOCK_SIZE = 4_000_000

var cfg = config.NewConfig()

type Core struct {
	ctx context.Context
	mu  *sync.Mutex
	cli *client.Client

	height      int
	totalFee    uint64
	totalAmount uint64
	totalWeight uint64
	poolCopy    []txpool.TxPool
	poolCopyMap map[string]txpool.TxPool
	poolSorted  []mtx.Tx

	storage storage.PoolRepository
	// blocks      []*mblock.Block
	parserJobCh chan string
}

func NewCore(ctx context.Context, cli *client.Client, s storage.PoolRepository) *Core {
	return &Core{
		ctx:         ctx,
		mu:          &sync.Mutex{},
		cli:         cli,
		storage:     s,
		poolCopy:    make([]txpool.TxPool, 0),
		poolCopyMap: make(map[string]txpool.TxPool),
		poolSorted:  make([]mtx.Tx, 0),
		// blocks:      make([]*mblock.Block, 0),
		parserJobCh: make(chan string),
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

	go c.workerPoolPuller(1 * time.Second)

	// make a batch of parsers
	// each parse makes a new RPC connection on every job
	for i := 0; i < cfg.RpcLimit; i++ {
		go c.workerTxParser()
	}

	go c.workerPoolSorter(1 * time.Second)
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

func (c *Core) GetPool(limit int) ([]mtx.Tx, error) {
	if len(c.poolSorted) <= limit {
		return c.poolSorted, nil
	}
	return c.poolSorted[:limit], nil
}

func (c *Core) GetHeight() int {
	return c.height
}

func (c *Core) GetPoolSize() int {
	return len(c.poolSorted)
}

func (c *Core) GetTotalAmount() uint64 {
	return c.totalAmount
}

func (c *Core) GetTotalFee() uint64 {
	return c.totalFee
}

func (c *Core) GetTotalWeight() uint64 {
	return c.totalWeight
}
