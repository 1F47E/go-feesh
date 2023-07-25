package core

import (
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/entity/btc/tx"
	"go-btc-scan/src/pkg/entity/btc/txpool"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	log "go-btc-scan/src/pkg/logger"
	"math/rand"
	"sort"
	"time"
)

func (c *Core) workerPoolPuller(period time.Duration) {
	name := "[wPoolPuller]"
	log.Log.Infof("[%s] started\n", name)
	ticker := time.NewTicker(period)
	defer func() {
		log.Log.Infof("[%s] stopped\n", name)
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
				log.Log.Errorf("error on getinfo: %v\n", err)
				continue
			}

			if c.height != info.Blocks {
				c.height = info.Blocks
				log.Log.Debugf("new block height: %d\n", info.Blocks)
			}

			// get ordered list of pool tsx. new first
			poolTxs, err := c.cli.RawMempool()
			if err != nil {
				log.Log.Errorf("%s error on rawmempool: %v\n", name, err)
				continue
			}
			if len(poolTxs) == 0 {
				continue
			}
			// check if we have new txs
			new := 0
			for _, tx := range poolTxs {
				if _, ok := c.poolCopyMap[tx.Txid]; !ok {
					new++
				}
			}
			if new == 0 {
				continue
			}
			log.Log.Debugf("%s got %d new txs\n", name, new)

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
					log.Log.Errorf("%s error on txget: %v\n", name, err)
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
		case txid := <-c.parserJobCh:
			var err error

			// parse tx with retry
			// log.Log.Debugf("%s parsing tx: %s\n", name, txid)
			maxRetry := 10
			btx := new(tx.Transaction)
			for i := 0; i <= maxRetry; i++ {
				// sleep randomly to not overload the node
				time.Sleep(time.Duration(100*i+rand.Intn(300)) * time.Millisecond)
				btx, err = c.cli.TransactionGet(txid)
				if err != nil {
					log.Log.Errorf("%s error on getrawtransaction: %v\n", name, err)
				}
				if err != nil {
					if err.Error() == client.ERR_5xx {
						log.Log.Errorf("%s error 5xx, retrying %d/%d\n", i+1, name, maxRetry)
						continue
					}
					log.Log.Errorf("%s error on parsePoolTx: %v\n", name, err)
					continue
				}
				break
			}
			// log.Log.Debugf("%s parsed tx txid: %s\n", name, txid)
			// remap raw tx to model
			tx := mtx.Tx{
				Hash:   txid,
				Amount: btx.GetTotalOut(),
			}
			// combine with data from pool copy (fees, time, etc)
			if poolTx, ok := c.poolCopyMap[txid]; ok {
				tx.Time = time.Unix(poolTx.Time, 0)
				tx.Size = poolTx.Size
				tx.Vsize = poolTx.Vsize
				tx.Weight = poolTx.Weight
				tx.Fee = poolTx.Fee
				// save tx
				err = c.storage.TxAdd(tx)
				if err != nil {
					log.Log.Errorf("%s error on storage.TxAdd: %v\n", name, err)
				}
			} else {
				log.Log.Errorf("%s error on poolCopyMap: tx not found: %s\n", name, txid)
			}
		}
	}
}

func (c *Core) workerPoolSorter(period time.Duration) {
	name := "workerPoolSorter"
	log.Log.Infof("[%s] started\n", name)
	ticker := time.NewTicker(period)
	defer func() {
		log.Log.Infof("[%s] stopped\n", name)
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
			buckets := []uint{2, 5, 10, 20, 50, 100, 200, 499}
			feeBuckets := make([]uint, len(buckets)+1)
			for _, tx := range c.poolCopy {
				// get parsed tx
				parsedTx, err := c.storage.TxGet(tx.Txid)
				if err != nil {
					log.Log.Errorf("%s error on txget: %v\n", name, err)
					continue
				}
				if parsedTx == nil {
					continue
				}
				res = append(res, *parsedTx)

				// totals
				amount += parsedTx.Amount
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
			totalWeight := uint64(0)
			for i, tx := range res {
				totalWeight += uint64(tx.Weight)
				res[i].Fits = true
				if totalWeight > BLOCK_SIZE {
					break
				}
			}

			// sort by time
			sort.Slice(res, func(i, j int) bool {
				if !res[i].Time.Equal(res[j].Time) {
					return res[i].Time.After(res[j].Time)
				}
				// sometimes time can be equal, sort by Hash
				return res[i].Hash < res[j].Hash
			})
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
			log.Log.Debugf("%s pool sorted, took: %v\n", name, time.Since(now))
			log.Log.Debugf("%s total txs: %d\n", name, len(res))
		}
	}
}
