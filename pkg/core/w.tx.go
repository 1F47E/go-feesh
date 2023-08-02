package core

import (
	"fmt"
	"time"

	mtx "github.com/1F47E/go-feesh/pkg/entity/models/tx"
	log "github.com/1F47E/go-feesh/pkg/logger"
)

// log carefull, there can be a lot of workers
func (c *Core) workerTxParser(n int) {
	l := log.Log.WithField("context", fmt.Sprintf("[workerTxParser] #%d", n))
	l.Trace("started\n")
	defer func() {
		l.Debug("stopped\n")
	}()
	for {
		select {
		case <-c.ctx.Done():
			return
		case txid := <-c.parserJobCh:
			var err error

			// parse tx
			// log.Log.Debugf("%s parsing tx: %s\n", name, txid)
			btx, err := c.cli.TransactionGet(txid)
			if err != nil {
				l.Errorf("error on getrawtransaction %s: %v\n", txid, err)
				continue
			}
			// log.Log.Debugf("%s parsed tx txid: %s\n", name, txid)

			// Vin
			// in order to calc fee we need input amounts.
			// to get them we have to parse Vin tx amounts
			// find out amount from vin tx matching by vout index
			var in int64
			for _, vin := range btx.Vin {
				// mined
				if vin.Coinbase != "" {
					in = -1
					continue
				}
				txIn, err := c.cli.TransactionGet(vin.Txid)
				if err != nil {
					log.Log.Errorf("error getting vin tx: %v\n", err)
					break
				}
				for _, vout := range txIn.Vout {
					if vout.N != vin.Vout {
						continue
					}
					in += int64(vout.Value * 1_0000_0000)
				}

				// remap raw tx to model and save
				// TODO: make constructor
				mtxIn := mtx.Tx{
					Hash:      vin.Txid,
					Time:      time.Unix(int64(btx.Time), 0),
					Size:      uint32(btx.Size),
					Weight:    uint32(btx.Weight),
					AmountOut: btx.GetTotalOut(),
					AmountIn:  0,
				}
				_ = c.storage.TxAdd(mtxIn)
			}
			if in <= 0 {
				if in == 0 {
					l.Errorf("no input amount, skipping tx: %s\n", txid)
				}
				// -1 is coinbase, no need to log error
			}

			// remap raw tx to model
			out := btx.GetTotalOut()
			fee := uint64(in) - out
			tx := mtx.Tx{
				Hash: txid,
				// NOTE: mempool tx dont have time in rawtransaction
				// only in custom ramempool tx we have pool time
				Time:      time.Unix(int64(btx.Time), 0),
				Size:      uint32(btx.Size),
				Weight:    uint32(btx.Weight),
				AmountOut: btx.GetTotalOut(),
				AmountIn:  uint64(in),
				Fee:       fee,
			}

			_ = c.storage.TxAdd(tx)
		}
	}
}
