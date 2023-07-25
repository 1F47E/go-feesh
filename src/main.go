package main

import (
	"context"
	"go-btc-scan/src/pkg/api"
	"go-btc-scan/src/pkg/client"
	"go-btc-scan/src/pkg/config"
	"go-btc-scan/src/pkg/core"
	mblock "go-btc-scan/src/pkg/entity/models/block"
	smap "go-btc-scan/src/pkg/storage/map"
	"log"
)

var cli *client.Client

func main() {
	var err error
	cfg := config.NewConfig()

	// create RPC client
	cli, err = client.NewClient(cfg.RpcHost, cfg.RpcUser, cfg.RpcPass)
	if err != nil {
		log.Fatalln("error creating client:", err)
	}

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

	// create storage

	// create in mem storage (debug only)
	s := smap.New()
	// create redis storage
	// s, err := sredis.New(ctx)
	// if err != nil {
	// 	log.Fatalln("error on redis storage:", err)
	// }

	// create core with RPC client and storage
	c := core.NewCore(ctx, cfg, cli, s)
	c.Start()
	a := api.NewApi(c)
	err = a.Listen()
	if err != nil {
		log.Fatalf("error on listen: %v", err)
	}

}
