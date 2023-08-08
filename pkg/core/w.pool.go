package core

import (
	"math/rand"
	"sort"
	"time"

	"github.com/1F47E/go-feesh/pkg/config"
	"github.com/1F47E/go-feesh/pkg/entity/btc/txpool"
	mtx "github.com/1F47E/go-feesh/pkg/entity/models/tx"
	"github.com/1F47E/go-feesh/pkg/logger"
	"github.com/1F47E/go-feesh/pkg/notificator"
)

// last bucket is 500+
var buckets = []uint{2, 3, 4, 5, 6, 8, 10, 15, 25, 35, 50, 70, 85, 100, 125, 150, 200, 250, 300, 350, 400, 450, 499, 500}

// var poolSizeHistoryTimeFrame = 1 * time.Minute
var poolSizeHistoryTimeFrame = 5 * time.Second
var poolSizeHistoryLimit = 40

func (c *Core) workerPoolPuller(period time.Duration) {
	log := logger.Log.WithField("context", "[workerPoolPuller]")
	log.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		log.Infof(" stopped\n")
		ticker.Stop()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// get the block height
			info, err := c.cli.GetInfo()
			if err != nil {
				log.Errorf("error on getinfo: %v\n", err)
				continue
			}

			if c.height != info.Blocks {
				c.height = info.Blocks
				log.Debugf("new block height: %d\n", info.Blocks)
			}

			// get ordered list of pool tsx. new first
			poolTxs, err := c.cli.RawMempool()
			if err != nil {
				log.Errorf("error on rawmempool: %v\n", err)
				continue
			}
			if len(poolTxs) == 0 {
				continue
			}

			// check if we have new txs
			hasNew := false
			for _, tx := range poolTxs {
				if _, ok := c.poolCopyMap[tx.Txid]; !ok {
					hasNew = true
					break
				}
			}
			if !hasNew {
				continue
			}
			log.Debugf("got some new txs\n")
			log.Warnf("new pool size: %d\n", len(poolTxs))

			// copy pool txs mem for later reference what pool have
			c.mu.Lock()
			c.poolCopy = make([]txpool.TxPool, len(poolTxs))
			c.poolCopyMap = make(map[string]txpool.TxPool)
			for i, tx := range poolTxs {
				c.poolCopy[i] = tx
				c.poolCopyMap[tx.Txid] = tx
			}
			c.mu.Unlock()

			// send new txs to parser
			for _, tx := range poolTxs {
				// skip if already parsed
				exists, err := c.storage.TxGet(tx.Txid)
				if err != nil {
					log.Errorf("error on txget: %v\n", err)
					continue
				}
				if exists != nil {
					continue
				}
				// log.Log.Debugf("%s new tx, sending to parser: %s\n", name, tx.Txid)
				c.parserJobCh <- tx.Txid
			}
		}
	}
}

func (c *Core) workerPoolSizeHistory(period time.Duration) {
	log := logger.Log.WithField("context", "[workerPoolSizeHistory]")
	log.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		log.Info("stopped")
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// add history if time passed
			size := uint(len(c.poolSorted))
			c.poolSizeHistory = append(c.poolSizeHistory, size)

			// cleanup old records
			if len(c.poolSizeHistory) > poolSizeHistoryLimit {
				// log.Warnf("pool history is too big, cleaning up. old len: %d\n", len(c.poolSizeHistory))
				c.poolSizeHistory = c.poolSizeHistory[len(c.poolSizeHistory)-poolSizeHistoryLimit:]
				// log.Warnf("new len: %d\n", len(c.poolSizeHistory))
			}
		}
	}
}

func (c *Core) workerPoolSorter(period time.Duration) {
	log := logger.Log.WithField("context", "[workerPoolSorter]")
	log.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		log.Info("stopped")
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// construct pool slice for API access
			// get pool copy, merge it with parsed tx
			// order by time

			// TODO: create another copy sorted by fee, calc fee buckets
			now := time.Now()

			res := make([]mtx.Tx, 0)
			c.mu.Lock()
			// collect parsed txs based on pool copy
			// also count totals
			var amount, fee, weight uint64
			feeBuckets := make([]uint, len(buckets))

			for _, tx := range c.poolCopy {
				// get parsed tx
				parsedTx, err := c.storage.TxGet(tx.Txid)
				if err != nil {
					log.Errorf("error on txget: %v\n", err)
					continue
				}
				if parsedTx == nil {
					continue
				}
				// fix time
				parsedTx.Time = time.Unix(int64(tx.Time), 0)

				res = append(res, *parsedTx)

				// totals
				amount += parsedTx.AmountOut
				fee += parsedTx.Fee
				weight += uint64(parsedTx.Weight)

				// count fee buckets
				feeB := parsedTx.FeePerByte()
				// fix max fee
				if feeB > 500 {
					feeB = 500
				}
				bucket := 0
				for i, b := range buckets {
					if feeB <= b {
						bucket = i
						break
					}
				}

				feeBuckets[bucket]++
			}
			// log.Warnf("fee buckets: (%d) %v\n", len(feeBuckets), feeBuckets)

			// sort by fee - check if tx will fit in the next block
			sort.Slice(res, func(i, j int) bool {
				return res[i].Fee > res[j].Fee
			})
			var totalWeight uint32
			for i := range res {
				if res[i].Fee == 0 {
					continue
				}
				if totalWeight+res[i].Weight > config.BLOCK_SIZE {
					break
				}
				totalWeight += res[i].Weight
				res[i].Fits = true
			}

			// sort by time
			sort.Slice(res, func(i, j int) bool {
				if !res[i].Time.Equal(res[j].Time) {
					return res[i].Time.After(res[j].Time)
				}
				// sometimes time can be equal, sort by Hash
				return res[i].Hash < res[j].Hash
			})
			prevPoolCnt := len(c.poolSorted)
			c.poolSorted = res
			c.totalAmount = amount
			c.totalFee = fee
			c.totalWeight = weight

			// TODO: fee estimator

			// Calc fee buckets
			bucketsMap := make(map[uint]uint)
			for i, b := range buckets {
				bucketsMap[b] = feeBuckets[i]
			}
			c.feeBucketsMap = bucketsMap
			c.feeBuckets = feeBuckets

			c.mu.Unlock()
			if prevPoolCnt != len(res) {
				log.Debugf("pool sorted, took: %v\n", time.Since(now))
				log.Debugf("total txs: %d\n", len(res))
			}

			feeAvg := 0
			if len(res) > 0 {
				feeAvg = int(fee / uint64(len(res)))
			}
			// fee butkets
			// TODO: move size to const
			var feeBucketsArr [24]uint
			copy(feeBucketsArr[:], feeBuckets)

			var poolSizeHistory [20]uint
			copy(poolSizeHistory[:], c.poolSizeHistory)
			// send websocket update
			msg := notificator.Msg{
				Height:          c.height,
				PoolSize:        len(res),
				PoolSizeHistory: poolSizeHistory,
				TotalFee:        int(fee),
				AvgFee:          feeAvg,
				Amount:          int(amount),
				Weight:          int(weight),
				FeeBuckets:      feeBucketsArr,
			}
			go c.nofity(msg)
		}
	}
}

func (c *Core) workerPoolDebug(period time.Duration) {
	log := logger.Log.WithField("context", "[workerPoolDebug]")
	log.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		log.Info("stopped")
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// send random msg
			history := [20]uint{}
			for i := range history {
				history[i] = uint(rand.Intn(1000))
			}
			buckets := [24]uint{}
			for i := range buckets {
				buckets[i] = uint(rand.Intn(1000))
			}
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			msg := notificator.Msg{
				Height:          r.Intn(1000),
				PoolSize:        r.Intn(1000),
				PoolSizeHistory: history,
				TotalFee:        r.Intn(1000),
				AvgFee:          r.Intn(1000),
				Amount:          r.Intn(1000),
				Weight:          r.Intn(1000),
				FeeBuckets:      buckets,
			}
			go c.nofity(msg)
		}
	}
}

// send websocket update
// with timeout, protection from blocking
func (c *Core) nofity(msg notificator.Msg) {
	select {
	case c.broadcastCh <- msg:
		// Message sent successfully
	case <-time.After(time.Second * 5):
		logger.Log.Error("timeout on sending websocket message\n")
	}
}
