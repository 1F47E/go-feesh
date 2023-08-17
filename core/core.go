package core

import (
	"context"
	"os"
	"time"

	"github.com/1F47E/go-feesh/client"
	"github.com/1F47E/go-feesh/config"
	"github.com/1F47E/go-feesh/logger"
	"github.com/1F47E/go-feesh/notificator"
	"github.com/1F47E/go-feesh/storage"

	"sync"

	"github.com/1F47E/go-feesh/entity/btc/info"
	"github.com/1F47E/go-feesh/entity/btc/txpool"
	mblock "github.com/1F47E/go-feesh/entity/models/block"
	mtx "github.com/1F47E/go-feesh/entity/models/tx"
)

type Core struct {
	ctx     context.Context
	mu      *sync.Mutex
	Cfg     *config.Config
	cli     *client.Client
	storage storage.PoolRepository
	// ws
	broadcastCh chan notificator.Msg

	height int

	// because total fee in sat will overflow uint64, sat in 1000 sats
	poolFeeTotal uint64
	poolFeeAvg   uint64

	totalAmount uint64
	// totalWeight uint64
	totalSize uint64

	feeBucketsMap map[uint]uint
	feeBuckets    []uint

	poolCopy        []txpool.TxPool
	poolCopyMap     map[string]txpool.TxPool
	poolSorted      []mtx.Tx
	poolSizeHistory []uint

	blockDepth  int      // how deep to scan the blocks from the top
	blocksIndex []string // keep track of parsed blocks
	blocks      []mblock.Block

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

		poolCopy:        make([]txpool.TxPool, 0),
		poolCopyMap:     make(map[string]txpool.TxPool),
		poolSorted:      make([]mtx.Tx, 0),
		poolSizeHistory: make([]uint, 0),
		// blocks:      make([]*mblock.Block, 0),
		blockDepth:  cfg.BlocksParsingDepth,
		blocksIndex: make([]string, 0),
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

	if os.Getenv("DEBUG") == "WS" {
		go c.workerPoolDebug(1 * time.Second)
		return
	}
	go c.workerPoolPuller(1 * time.Second)
	go c.workerPoolSorter(1 * time.Second)
	go c.workerPoolSizeHistory(5 * time.Minute)
}

func (c *Core) GetNodeInfo() (*info.Info, error) {
	return c.cli.GetInfo()
}

// parse last N blocks
//
//nolint:unused
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

func (c *Core) GetPoolSizeHistory() []uint {
	return c.poolSizeHistory
}

func (c *Core) GetTotalAmount() uint64 {
	return c.totalAmount
}

func (c *Core) GetFeeTotal() uint64 {
	return c.poolFeeTotal
}

func (c *Core) GetFeeAvg() uint64 {
	return c.poolFeeAvg
}

func (c *Core) GetFeeBucketsMap() map[uint]uint {
	return c.feeBucketsMap
}

func (c *Core) GetFeeBuckets() []uint {
	return c.feeBuckets
}

func (c *Core) GetTotalSize() uint64 {
	sizeBytes := c.totalSize
	sizeKb := sizeBytes / 1024
	// sizeMb := sizeKb / 1024
	return sizeKb
}

func (c *Core) GetBlocks() []mblock.Block {
	return c.blocks
}
