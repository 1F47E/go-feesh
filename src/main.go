package main

import (
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/entity/txpool"
	"go-btc-scan/src/pkg/utils"
	"log"
	"os"
	"sync"
	"time"
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

	// debug calc mempool fee demo
	mu := &sync.Mutex{}
	pool := make(map[string]*txpool.TxPool)
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
						pool[tx] = &txpool.TxPool{
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
						log.Printf("tx %s in: %d, out: %d\n", tx.Hash, tx.AmountIn, tx.AmountOut)
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
			fee += tx.Fee()
		}
		poolGoodCnt := 0
		for _, tx := range pool {
			if tx.AmountIn != 0 && tx.AmountOut != 0 {
				poolGoodCnt++
			}
		}
		if cnt != poolGoodCnt {
			log.Printf("total in: %d, total out: %d, fee: %d\n", totalIn, totalOut, fee)
		}
		cnt = len(pool)
		mu.Unlock()
	}

}

func getTxAmounts(txid string) (uint64, uint64) {
	tx, err := cli.TransactionGet(txid)
	if err != nil {
		log.Fatalln("error on gettransaction:", err)
	}
	// ===== find out amount from vin tx matching by vout index
	var in uint64
	for _, vin := range tx.Vin {
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
	err = cli.GetInfo()
	if err != nil {
		log.Fatalln("error on getinfo:", err)
	}

	// get block
	blockHash := "00000000000000048e1b327dd79f72fab6395cc09a049e54fe2c0b90aa837914"
	err = cli.GetBlock(blockHash)
	if err != nil {
		log.Fatalln("error on getblock:", err)
	}

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
