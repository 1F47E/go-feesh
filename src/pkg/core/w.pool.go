package core

import (
	"go-btc-scan/src/pkg/config"
	"go-btc-scan/src/pkg/entity/btc/txpool"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	log "go-btc-scan/src/pkg/logger"
	"sort"
	"time"
)

func (c *Core) workerPoolPuller(period time.Duration) {
	l := log.Log.WithField("context", "[workerPoolPuller]")
	l.Infof("started\n")
	ticker := time.NewTicker(period)
	defer func() {
		l.Infof(" stopped\n")
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
				l.Errorf("error on getinfo: %v\n", err)
				continue
			}

			if c.height != info.Blocks {
				c.height = info.Blocks
				l.Debugf("new block height: %d\n", info.Blocks)
			}

			// get ordered list of pool tsx. new first
			poolTxs, err := c.cli.RawMempool()
			if err != nil {
				l.Errorf("error on rawmempool: %v\n", err)
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
			l.Debugf("got some new txs\n")

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
					l.Errorf("error on txget: %v\n", err)
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

func (c *Core) workerPoolSorter(period time.Duration) {
	l := log.Log.WithField("context", "[workerPoolSorter]")
	l.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		l.Info("stopped")
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
			buckets := []uint{2, 3, 4, 5, 6, 8, 10, 15, 25, 35, 50, 70, 85, 100, 125, 150, 200, 250, 300, 350, 400, 450, 499}
			feeBuckets := make([]uint, len(buckets)+1)

			for _, tx := range c.poolCopy {
				// get parsed tx
				parsedTx, err := c.storage.TxGet(tx.Txid)
				if err != nil {
					l.Errorf("error on txget: %v\n", err)
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
				bucket := 0
				for i, b := range buckets {
					if feeB <= b {
						bucket = i
						break
					}
				}
				// fee is too big
				if feeB > buckets[len(buckets)-1] {
					bucket = len(buckets)
				}
				feeBuckets[bucket]++
			}

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

			bucketsMap := make(map[uint]uint)
			for i, b := range buckets {
				bucketsMap[b] = feeBuckets[i]
			}
			c.feeBuckets = bucketsMap

			c.mu.Unlock()
			if prevPoolCnt != len(res) {
				l.Debugf("pool sorted, took: %v\n", time.Since(now))
				l.Debugf("total txs: %d\n", len(res))
			}
		}
	}
}
