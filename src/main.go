package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-btc-scan/src/entity"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// ===== Data

type RPCRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
}

/*
	{
	  "jsonrpc": "1.0",
	  "result": {
	    "version": 230300,
	    "protocolversion": 70002,
	    "blocks": 2441417,
	    "timeoffset": 0,
	    "connections": 8,
	    "proxy": "",
	    "difficulty": 117392538.8721802,
	    "testnet": true,
	    "relayfee": 1e-05,
	    "errors": ""
	  },
	  "error": null,
	  "id": 1
	}
*/
type RPCResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   interface{}     `json:"error"`
}

func NewRPCRequest(method string, params interface{}) *RPCRequest {
	return &RPCRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}
}

// ===== CLIENT
type Client struct {
	client   *http.Client
	host     string
	user     string
	password string
}

func NewClient(host, user, password string) (*Client, error) {
	if host == "" || user == "" || password == "" {
		return nil, fmt.Errorf("RPC_HOST, RPC_USER and RPC_PASSWORD env vars must be set")
	}
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		host:     host,
		user:     user,
		password: password,
	}, nil
}

func (c *Client) doRequest(r *RPCRequest) (*RPCResponse, error) {
	jr, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(jr)

	req, err := http.NewRequest(http.MethodPost, c.host, bodyReader)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// debug
	// read response to bytes
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// log.Printf("HTTP CLIENT response: %s\n", data)

	var ret RPCResponse
	err = json.Unmarshal(data, &ret)

	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}

	log.Println("HTTP CLIENT response RPCResponse OK")
	log.Printf("HTTP CLIENT response RPCResponse result json: %s\n", ret.Result)
	return &ret, nil
}

// getinfo request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) getInfo() error {
	r := NewRPCRequest("getinfo", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	var info entity.ResponseGetinfo
	err = json.Unmarshal(data.Result, &info)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}
	log.Printf("getinfo response: %+v\n", info)
	return nil
}

// rawmempool request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawmempool","params":[true],"id":1}' http://localhost:18334
func (c *Client) rawMempool() error {
	r := NewRPCRequest("getrawmempool", []interface{}{true})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	// resp := make(map[string]MemPoolTx)
	var resp map[string]entity.MemPoolTx
	err = json.Unmarshal(data.Result, &resp)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}
	log.Printf("raw mempool transactions found %d\n", len(resp))
	for k, v := range resp {
		log.Printf("txid: %s, fee: %f\n", k, v.Fee)
	}
	return nil
}

func (c *Client) getBlock(blockHash string) error {
	r := NewRPCRequest("getblock", []interface{}{blockHash})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	var resp entity.Block
	err = json.Unmarshal(data.Result, &resp)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}
	log.Printf("block: %+v\n", resp)
	return nil
}

func main() {
	var err error
	// get vars from env
	// TODO: move to config
	nodeUrl := os.Getenv("RPC_HOST")
	username := os.Getenv("RPC_USER")
	password := os.Getenv("RPC_PASSWORD")

	cli, err := NewClient(nodeUrl, username, password)
	if err != nil {
		log.Fatalln("error creating client:", err)
	}

	err = cli.getInfo()
	if err != nil {
		log.Fatalln("error on getinfo:", err)
	}
	err = cli.rawMempool()
	if err != nil {
		log.Fatalln("error on rawmempool:", err)
	}

	blockHash := "00000000000000048e1b327dd79f72fab6395cc09a049e54fe2c0b90aa837914"
	err = cli.getBlock(blockHash)
	if err != nil {
		log.Fatalln("error on getblock:", err)
	}
}
