package core

import (
	"context"
	"os"
	"time"

	"github.com/1F47E/go-feesh/pkg/client"
	"github.com/1F47E/go-feesh/pkg/config"
	"github.com/1F47E/go-feesh/pkg/logger"
	"github.com/1F47E/go-feesh/pkg/notificator"
	"github.com/1F47E/go-feesh/pkg/storage"

	"sync"

	"github.com/1F47E/go-feesh/pkg/entity/btc/info"
	"github.com/1F47E/go-feesh/pkg/entity/btc/txpool"
	mtx "github.com/1F47E/go-feesh/pkg/entity/models/tx"
)

type Core struct {
	ctx     context.Context
	mu      *sync.Mutex
	Cfg     *config.Config
	cli     *client.Client
	storage storage.PoolRepository
	// ws
	broadcastCh chan notificator.Msg

	height      int
	totalFee    uint64
	totalAmount uint64
	totalWeight uint64

	feeBuckets map[uint]uint

	poolCopy    []txpool.TxPool
	poolCopyMap map[string]txpool.TxPool
	poolSorted  []mtx.Tx

	blockDepth int      // how deep to scan the blocks from the top
	blocks     []string // keep track of parsed blocks

	// blocks      []*mblock.Block
	parserJobCh chan string
}

func NewCore(ctx context.Context, cfg *config.Config, cli *client.Client, s storage.PoolRepository, broadcastCh chan notificator.Msg) *Core {
	return &Core{
		ctx:         ctx,
		mu:          &sync.Mutex{},
		Cfg:         cfg,
		cli:         cli,
		storage:     s,
		broadcastCh: broadcastCh,

		poolCopy:    make([]txpool.TxPool, 0),
		poolCopyMap: make(map[string]txpool.TxPool),
		poolSorted:  make([]mtx.Tx, 0),
		// blocks:      make([]*mblock.Block, 0),
		blockDepth: cfg.BlocksParsingDepth,
		blocks:     make([]string, 0),
		// block:       make(map[string]string),
		parserJobCh: make(chan string),
	}
}

func (c *Core) Start() {
	log := logger.Log.WithField("context", "[core]")
	if os.Getenv("DRY") == "1" {
		return
	}
	// TODO: move best block to worker
	// set the pool block height
	info, err := c.cli.GetInfo()
	if err != nil {
		log.Errorf("error on getinfo: %v\n", err)
	} else {
		// even if its fails - having block 0 will update pool txs list every time
		// its just for performance reasons
		c.height = info.Blocks
	}

	go c.workerParserBlocks(3 * time.Second)
	go c.workerBlocksProcessor(1 * time.Second)

	// make a batch of parsers
	// each parse makes a new RPC connection on every job
	for i := 0; i < c.Cfg.RpcLimit; i++ {
		go c.workerTxParser(i + 1)
	}

	go c.workerPoolPuller(1 * time.Second)
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

func (c *Core) GetFeeBuckets() map[uint]uint {
	return c.feeBuckets
}

func (c *Core) GetTotalWeight() uint64 {
	return c.totalWeight
}

func (c *Core) GetBlocks() []string {
	return c.blocks
}
