package main

import (
	"context"
	"go-btc-scan/src/pkg/api"
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/core"
	mblock "go-btc-scan/src/pkg/entity/models/block"
	mtx "go-btc-scan/src/pkg/entity/models/tx"
	"go-btc-scan/src/pkg/utils"
	"log"
	"os"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcutil"
)

var cli *client.Client

func init() {
	var err error
	nodeUrl := os.Getenv("RPC_HOST")
	username := os.Getenv("RPC_USER")
	password := os.Getenv("RPC_PASSWORD")
	cli, err = client.NewClient(nodeUrl, username, password)
	if err != nil {
		log.Fatalln("error creating client:", err)
	}
}

func main() {
	// get node info
	info, err := cli.GetInfo()
	if err != nil {
		log.Fatalln("error on getinfo:", err)
	}
	log.Printf("node info: %+v\n", info)

	// get last block hash
	bestBlock, err := cli.GetBestBlock()
	if err != nil {
		log.Fatalln("error on getbestblock:", err)
	}
	log.Println("last block hash:", bestBlock.Hash)

	b, err := cli.GetBlock(bestBlock.Hash)
	if err != nil {
		log.Fatalln("error on getblock:", err)
	}
	log.Println("block tx cnt:", len(b.Transactions))

	// parse block tx, calc value and fee
	// txs := make([]*txpool.TxPool, len(b.Transactions))
	var totalValue, totalFee uint64
	// for _, txid := range b.Transactions {
	// in, out := getTxAmounts(txid)
	// tx := &mtx.Tx{
	// 	Hash:      txid,
	// 	AmountIn:  in,
	// 	AmountOut: out,
	// }
	// totalValue += in
	// totalFee += tx.Fee()
	// txs[i] = tx
	// }
	wBlock := &mblock.Block{
		Hash:   b.Hash,
		Height: b.Height,
		Value:  totalValue,
		Fee:    totalFee,
	}
	log.Printf("block %d, value: %d, fee: %d\n", wBlock.Height, wBlock.Value, wBlock.Fee)

	// get block header
	// header, err := cli.GetBlockHeader(bestBlock.Hash)
	// if err != nil {
	// 	log.Fatalln("error on getblockheader:", err)
	// }
	// log.Println("prev block hash:", header.Previousblockhash)
	// // get full block data (tx list)
	// b, err := cli.GetBlock(bestBlock.Hash)
	// if err != nil {
	// 	log.Fatalln("error on getblock:", err)
	// }
	// log.Println("block tx cnt:", len(b.Transactions))

	// TODO: graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := core.NewCore(ctx, cli)
	c.Start()
	a := api.NewApi(c)
	err = a.Listen()
	if err != nil {
		log.Fatalf("error on listen: %v", err)
	}

}

func debugParseMempool() {

	// debug calc mempool fee demo
	mu := &sync.Mutex{}
	pool := make(map[string]*mtx.Tx)
	cnt := 0

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ticker.C:
				txs, err := cli.RawMempool()
				if err != nil {
					log.Fatalln("error on rawmempool:", err)
				}
				if len(txs) != cnt {
					log.Println("Raw mempool cnt:", len(txs))
				}
				cnt = len(txs)
				// add new tx to the pool
				mu.Lock()
				for _, tx := range txs {
					if _, ok := pool[tx]; !ok {
						pool[tx] = &mtx.Tx{
							Hash: tx,
						}
					}
				}
				mu.Unlock()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				for _, tx := range pool {
					// check tx data
					if tx.AmountIn == 0 || tx.AmountOut == 0 {
						in, out := getTxAmounts(tx.Hash)
						tx.AmountIn = in
						tx.AmountOut = out
						// report
						inBtc := btcutil.Amount(in)
						outBtc := btcutil.Amount(out)
						log.Printf("tx %s in: %s, out: %s\n", tx.Hash, inBtc.String(), outBtc.String())
					}
				}
				mu.Unlock()
			}
		}
	}()

	cnt = 0
	for {
		time.Sleep(1 * time.Second)
		mu.Lock()
		var totalIn, totalOut, fee uint64
		for _, tx := range pool {
			totalIn += tx.AmountIn
			totalOut += tx.AmountOut
			fee += tx.Fee
		}
		poolGoodCnt := 0
		for _, tx := range pool {
			if tx.AmountIn != 0 && tx.AmountOut != 0 {
				poolGoodCnt++
			}
		}
		if cnt != poolGoodCnt {
			totalInBtc := btcutil.Amount(totalIn)
			totalOutBtc := btcutil.Amount(totalOut)
			feeBtc := btcutil.Amount(fee)
			log.Printf("total in: %s, out: %s, fee: %s\n", totalInBtc.String(), totalOutBtc.String(), feeBtc.String())
		}
		cnt = len(pool)
		mu.Unlock()
	}

}

func getTxAmounts(txid string) (uint64, uint64) {
	if len(txid) != 64 {
		log.Fatalln("getTxAmounts invalid txid:", txid)
	}
	// txidparam := string(t)
	tx, err := cli.TransactionGet(txid)
	if err != nil {
		log.Fatalln("error on gettransaction:", err)
	}
	// ===== find out amount from vin tx matching by vout index
	var in uint64
	for _, vin := range tx.Vin {
		// mined
		if vin.Coinbase != "" {
			continue
		}
		txIn, err := cli.TransactionGet(vin.Txid)
		if err != nil {
			log.Fatalln("error on gettransaction:", err)
		}
		for _, vout := range txIn.Vout {
			if vout.N != vin.Vout {
				continue
			}
			in += uint64(vout.Value * 1_0000_0000)
		}
	}
	out := tx.GetTotalOut()
	return in, out
	// log.Println("in amount:", in)
	// fee := in - out
	// log.Printf("fee sat: %d\n", fee)
	// fee per byte
	// feePerByte := float64(fee) / float64(tx.Size)
	// log.Printf("fee per byte: %.1f\n", feePerByte)

}

func debug() {
	var err error

	// get node info
	into, err := cli.GetInfo()
	if err != nil {
		log.Fatalln("error on getinfo:", err)
	}
	log.Printf("node info: %+v\n", into)

	// get block
	blockHash := "00000000000000048e1b327dd79f72fab6395cc09a049e54fe2c0b90aa837914"
	b, err := cli.GetBlock(blockHash)
	if err != nil {
		log.Fatalln("error on getblock:", err)
	}
	log.Printf("block %s tx cnt: %d\n", blockHash, len(b.Transactions))

	// get raw tx
	txHash := "6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"
	tx, err := cli.TransactionGet(txHash)
	if err != nil {
		log.Fatalln("error on gettransaction:", err)
	}
	// decode tx
	// tx, err := cli.TransactionDecode(txData)
	// if err != nil {
	// 	log.Fatalln("error on decoderawtransaction:", err)
	// }
	utils.PrintStruct(tx)

	// get peers
	peers, err := cli.GetPeers()
	if err != nil {
		log.Fatalln("error on getpeerinfo:", err)
	}
	log.Println("Peers:")
	for _, p := range peers {
		log.Println(p.Addr)
	}

	// get raw mempool
	txs, err := cli.RawMempool()
	if err != nil {
		log.Fatalln("error on rawmempool:", err)
	}
	log.Println("Raw mempool:", len(txs))

	// get extended mempool
	txs2, err := cli.RawMempoolExtended()
	if err != nil {
		log.Fatalln("error on rawmempool:", err)
	}
	log.Println("Raw mempool extended:", len(txs2))
}
