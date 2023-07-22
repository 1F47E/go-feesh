package main

import (
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/utils"
	"log"
	"os"
)

func main() {
	var err error
	// get vars from env
	// TODO: move to config
	nodeUrl := os.Getenv("RPC_HOST")
	username := os.Getenv("RPC_USER")
	password := os.Getenv("RPC_PASSWORD")

	cli, err := client.NewClient(nodeUrl, username, password)
	if err != nil {
		log.Fatalln("error creating client:", err)
	}

	// get node info
	err = cli.GetInfo()
	if err != nil {
		log.Fatalln("error on getinfo:", err)
	}

	// get raw mempool
	txs, err := cli.RawMempool()
	if err != nil {
		log.Fatalln("error on rawmempool:", err)
	}
	log.Println("Raw mempool:", len(txs))

	// get block
	blockHash := "00000000000000048e1b327dd79f72fab6395cc09a049e54fe2c0b90aa837914"
	err = cli.GetBlock(blockHash)
	if err != nil {
		log.Fatalln("error on getblock:", err)
	}

	// get raw tx
	txHash := "6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"
	txData, err := cli.TransactionGet(txHash)
	if err != nil {
		log.Fatalln("error on gettransaction:", err)
	}
	// decode tx
	tx, err := cli.TransactionDecode(txData)
	if err != nil {
		log.Fatalln("error on decoderawtransaction:", err)
	}
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
}
