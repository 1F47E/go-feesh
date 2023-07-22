package main

import (
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/utils"
	"log"
	"os"
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
	// ticker := time.NewTicker(1 * time.Second)
	// cnt := 0
	// for {
	// 	select {
	// 	case <-ticker.C:
	// 		txs, err := cli.RawMempool()
	// 		if err != nil {
	// 			log.Fatalln("error on rawmempool:", err)
	// 		}
	// 		if len(txs) != cnt {
	// 			log.Println("Raw mempool cnt:", len(txs))
	// 		}
	// 		cnt = len(txs)
	// 	}
	// }
	txid := "8e4cd481812a6fefc86ce7f3fcd7a668fa840465dd8ffaa09761026185e5e7a9"
	tx, err := cli.TransactionGet(txid)
	if err != nil {
		log.Fatalln("error on gettransaction:", err)
	}
	// print struct
	utils.PrintStruct(tx)
	// out
	out := tx.GetTotalOut()
	log.Println("out amount:", out)
	// get in tx out
	var in uint64
	for _, vin := range tx.Vin {
		txIn, err := cli.TransactionGet(vin.Txid)
		if err != nil {
			log.Fatalln("error on gettransaction:", err)
		}
		in += txIn.GetTotalOut()
	}
	log.Println("in amount:", in)
	fee := in - out
	log.Printf("fee sat: %d\n", fee)
	// fee per byte
	feePerByte := float64(fee) / float64(tx.Size)
	log.Printf("fee per byte: %f\n", feePerByte)

	// utils.PrintStruct(tx)

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
