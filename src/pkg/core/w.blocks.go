package core

import (
	log "go-btc-scan/src/pkg/logger"
	"time"
)

func (c *Core) workerParserBlocks(period time.Duration) {
	l := log.Log.WithField("context", "[workerParserBlocks]")
	l.Infof("started\n")
	ticker := time.NewTicker(period)
	defer func() {
		l.Infof(" stopped\n")
		ticker.Stop()
	}()

	// WARN: debug
	c.height = 0

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
			// skip if initial blocks already parsed and no new blocks
			if c.height == info.Blocks && len(c.blocks) > 0 {
				continue
			}
			c.height = info.Blocks
			l.Debugf("new block height: %d\n", info.Blocks)

			// get best block
			best, err := c.cli.GetBestBlock()
			if err != nil {
				l.Errorf("error on getbestblock: %v\n", err)
				continue
			}

			// collect N block hashes
			// around 3k txs in a block and around 1.5Meg for txs data
			blocks := make([]string, 0)
			currentHash := best.Hash
			for i := 0; i < c.blockDepth; i++ {
				blocks = append(blocks, currentHash)
				header, err := c.cli.GetBlockHeader(currentHash)
				if err != nil {
					l.Errorf("error on getblockheader: %v\n", err)
					continue
				}
				// l.Debugf("best block hash: %s\n", best.Hash)
				// l.Debugf("prev block hash: %s\n", header.Previousblockhash)
				currentHash = header.Previousblockhash
			}
			l.Debugf("got %d last blocks\n", len(blocks))
			for _, hash := range blocks {
				l.Debugf("block hash: %s\n", hash)
			}

			// parse N blocks
			now := time.Now()
			for i, hash := range blocks {
				// get full block data (tx list)
				exists, _ := c.storage.BlockExists(hash)
				if !exists {
					l.Debugf("%d/%d block parsing: %s\n", i+1, len(blocks), hash)
					b, err := c.cli.GetBlock(hash)
					if err != nil {
						l.Errorf("error on getblock: %v\n", err)
					}
					_ = c.storage.BlockAdd(b.Hash, b.Transactions)
					// add to in mem blocks index
					c.mu.Lock()
					c.blocks = append(c.blocks, b.Hash)
					c.mu.Unlock()
					// send block txs parser
					txs, _ := c.storage.BlockGet(b.Hash)
					for _, txid := range txs {
						c.parserJobCh <- txid
					}
				}
			}
			l.Debugf("blocks %d processed in %s\n", len(blocks), time.Since(now))
		}
	}
}

func (c *Core) workerBlocksProcessor(period time.Duration) {
	l := log.Log.WithField("context", "[workerBlocksProcessor]")
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
			// check blocks and what tx are parsed
			txCnt := 0
			for _, hash := range c.blocks {
				var bWeight, bFee, bAmount uint64
				// log.Log.Debugf("checking block %s\n", hash)
				txs, _ := c.storage.BlockGet(hash)
				// log.Log.Debugf("block has %s txs: %d\n", hash, len(txs))
				cnt := 0
				for _, txid := range txs {
					// check if tx is parsed
					tx, _ := c.storage.TxGet(txid)
					if tx != nil {
						cnt++
						bWeight += uint64(tx.Weight)
						bFee += tx.Fee
						bAmount += tx.AmountOut
					}
				}
				txCnt += cnt
				// l.Debugf("block %s has %d/%d txs parsed. Weight: %d, Amount: %d", hash, cnt, len(txs), bWeight, bAmount)
			}
			if txCnt > 0 {
				l.Debugf("total parsed txs: %d\n", txCnt)
			}
		}
	}
}
