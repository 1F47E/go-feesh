package core

import (
	"context"
	"time"

	mblock "github.com/1F47E/go-feesh/entity/models/block"
	"github.com/1F47E/go-feesh/logger"
)

func (c *Core) workerParserBlocks(ctx context.Context, period time.Duration) {
	log := logger.Log.WithField("context", "[workerParserBlocks]")
	log.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		log.Infof(" stopped\n")
		ticker.Stop()
	}()

	// WARN: debug reset
	c.height = 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// get the block height
			info, err := c.cli.GetInfo()
			if err != nil {
				log.Errorf("error on getinfo: %v\n", err)
				continue
			}
			// skip if initial blocks already parsed and no new blocks
			if c.height == info.Blocks && len(c.blocks) > 0 {
				continue
			}
			c.height = info.Blocks
			log.Debugf("new block height: %d\n", info.Blocks)

			// get best block
			best, err := c.cli.GetBestBlock()
			if err != nil {
				log.Errorf("error on getbestblock: %v\n", err)
				continue
			}

			if len(c.blocks) > c.blockDepth {
				log.Warnf("blocks buffer is full. dropping oldest block. have %d blocks", len(c.blocks))
				c.blocks = c.blocks[:len(c.blocks)-1]
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
					log.Errorf("error on getblockheader: %v\n", err)
					continue
				}
				// l.Debugf("best block hash: %s\n", best.Hash)
				// l.Debugf("prev block hash: %s\n", header.Previousblockhash)
				currentHash = header.Previousblockhash
			}
			log.Debugf("got %d last blocks\n", len(blocks))
			for _, hash := range blocks {
				log.Debugf("block hash: %s\n", hash)
			}

			// parse N blocks
			now := time.Now()
			for i, hash := range blocks {
				// get full block data (tx list)
				exists, _ := c.storage.BlockExists(hash)
				if !exists {
					log.Debugf("%d/%d block parsing: %s\n", i+1, len(blocks), hash)
					b, err := c.cli.GetBlock(hash)
					if err != nil {
						log.Errorf("error on getblock: %v\n", err)
					}
					// TODO: store raw block info also
					_ = c.storage.BlockAdd(b.Hash, b.Transactions)
					// add to in mem blocks index
					c.mu.Lock()
					c.blocksIndex = append(c.blocksIndex, b.Hash)
					c.mu.Unlock()
					// send block txs parser
					txs, _ := c.storage.BlockGet(b.Hash)
					for _, txid := range txs {
						c.parserJobCh <- txid
					}
				}
			}
			log.Debugf("blocks %d processed in %s\n", len(blocks), time.Since(now))
		}
	}
}

func (c *Core) workerBlocksProcessor(ctx context.Context, period time.Duration) {
	log := logger.Log.WithField("context", "[workerBlocksProcessor]")
	log.Info("started")
	ticker := time.NewTicker(period)
	defer func() {
		log.Infof(" stopped\n")
		ticker.Stop()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// check blocks and what tx are parsed
			txCnt := 0
			if len(c.blocksIndex) == len(c.blocks) {
				continue
			}
			log.Info("processing blocks")
			for _, hash := range c.blocksIndex {
				var bWeight, bSize, bFee, bAmount uint64
				// log.Log.Debugf("checking block %s\n", hash)
				txs, _ := c.storage.BlockGet(hash)
				// log.Log.Debugf("block has %s txs: %d\n", hash, len(txs))
				cnt := 0
				for _, txid := range txs {
					// check if tx is parsed
					tx, _ := c.storage.TxGet(txid)
					if tx != nil {
						// skip first one. TODO: detect coinbase by param
						cnt++
						if cnt == 1 {
							continue
						}
						bWeight += uint64(tx.Weight)
						bSize += uint64(tx.Size)
						bFee += tx.Fee
						bAmount += tx.AmountOut
						log.Debugf("block %s tx %s fee %d amount %d\n", hash, txid, tx.Fee, tx.AmountOut)
					}
				}
				log.Debugf("block %s has tx %d parsed. total fee: %d amount: %d\n", hash, cnt, bFee, bAmount)
				txCnt += cnt
				// save block stats
				b := mblock.Block{
					Hash:   hash,
					Txs:    uint64(len(txs)),
					Weight: bWeight,
					Size:   bSize,
					Fee:    bFee,
					Value:  bAmount,
				}
				c.blocks = append(c.blocks, b)
				log.Infof("block %s added to blocks list. cnt: %d\n", hash, cnt)
				// TODO: add to storage
				// l.Debugf("block %s has %d/%d txs parsed. Weight: %d, Amount: %d", hash, cnt, len(txs), bWeight, bAmount)
			}
			if txCnt > 0 {
				log.Debugf("total parsed txs: %d\n", txCnt)
			}
		}
	}
}
