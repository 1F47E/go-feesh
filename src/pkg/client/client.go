package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-btc-scan/src/pkg/entity"
	"go-btc-scan/src/pkg/entity/peer"
	"go-btc-scan/src/pkg/entity/tx"
	"go-btc-scan/src/pkg/utils"
	"io"
	"log"
	"net/http"
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
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
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
	msg := fmt.Sprintf("====== HTTP CLIENT cmd %s : %s", r.Method, r.Params)
	log.Println("\n\n", msg)
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
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ret RPCResponse
	err = json.Unmarshal(data, &ret)

	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}

	msg += " OK"
	log.Println(msg)
	return &ret, nil
}

// getinfo request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) GetInfo() error {
	r := NewRPCRequest("getinfo", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		return err
	}

	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Fatalf("Error marshalling back to raw JSON: %v", err)
	}

	// parse into struct
	var info entity.ResponseGetinfo
	err = json.Unmarshal(rawJson, &info)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
		return err
	}
	utils.PrintStruct(info)
	return nil
}

// rawmempool request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawmempool","params":[true],"id":1}' http://localhost:18334
func (c *Client) RawMempool() error {
	r := NewRPCRequest("getrawmempool", []interface{}{true})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Fatalf("Error marshalling back to raw JSON: %v", err)
	}

	var resp map[string]entity.MemPoolTx
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}

	log.Printf("raw mempool transactions found %d\n", len(resp))
	for k, v := range resp {
		log.Printf("txid: %s, fee: %f\n", k, v.Fee)
	}
	return nil
}

func (c *Client) GetBlock(blockHash string) error {
	r := NewRPCRequest("getblock", []interface{}{blockHash})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Fatalf("Error marshalling back to raw JSON: %v", err)
	}

	// parse into struct
	var resp entity.Block
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}
	utils.PrintStruct(resp)
	return nil
}

// get transaction
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawtransaction","params":["6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"],"id":1}' http://localhost:18334
func (c *Client) TransactionGet(txid string) (string, error) {
	fmt.Println("=== transactionGet")
	r := NewRPCRequest("getrawtransaction", []interface{}{txid})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	// check type of result is string
	if _, ok := data.Result.(string); !ok {
		return "", fmt.Errorf("unexpected type for result")
	}
	return data.Result.(string), nil
}

// decode raw transaction
func (c *Client) TransactionDecode(txdata string) (*tx.Transaction, error) {
	fmt.Println("=== transactionDecode")
	r := NewRPCRequest("decoderawtransaction", []interface{}{txdata})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Fatalf("Error marshalling back to raw JSON: %v", err)
	}

	// parse into struct
	var resp tx.Transaction
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}
	return &resp, nil
}

// get peer info
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getpeerinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) GetPeers() ([]*peer.Peer, error) {
	fmt.Println("=== getPeers")
	r := NewRPCRequest("getpeerinfo", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %v", err)
	}
	// check type of Result
	if _, ok := data.Result.([]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling back to raw JSON: %v", err)
	}
	// parse into struct
	var resp []*peer.Peer
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	return resp, nil
}
