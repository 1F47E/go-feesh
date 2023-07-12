package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-btc-scan/src/entity"
	entity_tx "go-btc-scan/src/entity/tx"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
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
	Jsonrpc string `json:"jsonrpc"`
	// Result  json.RawMessage `json:"result"`
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

// type RPCStringResponse struct {
// 	Jsonrpc string      `json:"jsonrpc"`
// 	Result  string      `json:"result"`
// 	Error   interface{} `json:"error"`
// }

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

	msg += " OK"
	log.Println(msg)
	// log.Printf("HTTP CLIENT response RPCResponse result json: %s\n", ret.Result)
	return &ret, nil
}

// getinfo request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) getInfo() error {
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
	printStruct(info)
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

func (c *Client) getBlock(blockHash string) error {
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
	printStruct(resp)
	return nil
}

// get transaction
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawtransaction","params":["6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"],"id":1}' http://localhost:18334
func (c *Client) transactionGet(txid string) (string, error) {
	fmt.Println("=== transactionGet")
	r := NewRPCRequest("getrawtransaction", []interface{}{txid})
	data, err := c.doRequest(r)
	if err != nil {
		log.Fatalln("error doing request:", err)
	}
	// fmt.Printf("=== transactionGet Result: %v, type: %T\n", data.Result, data.Result)
	// check type of result is string
	if _, ok := data.Result.(string); !ok {
		return "", fmt.Errorf("unexpected type for result")
	}
	return data.Result.(string), nil
}

// decode raw transaction
func (c *Client) transactionDecode(txdata string) (*entity_tx.Transaction, error) {
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
	var resp entity_tx.Transaction
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		log.Fatalln("error unmarshalling response:", err)
	}
	// log.Printf("transaction: %+v\n", resp)
	return &resp, nil
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

	txHash := "6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"
	// get raw tx
	txData, err := cli.transactionGet(txHash)
	if err != nil {
		log.Fatalln("error on gettransaction:", err)
	}
	// decode tx
	tx, err := cli.transactionDecode(txData)
	if err != nil {
		log.Fatalln("error on decoderawtransaction:", err)
	}
	printStruct(tx)

}

// ==== UTILS
func printStruct(s interface{}) {
	utilsPrintStruct(s, "")
}

// iterate over tx and print all fields
func utilsPrintStruct(s interface{}, indent string) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		fmt.Println("printStruct only accepts structs; got", v.Kind())
		return
	}

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("\n%s%s\n", indent, v.Type().Field(i).Name)
		field := v.Field(i)

		// If this is a nested struct, call the function recursively
		if field.Kind() == reflect.Struct {
			utilsPrintStruct(field.Interface(), indent+"  ")
		} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			utilsPrintStruct(field.Interface(), indent+"  ")
		} else {
			// This is not a nested struct, so just print the value
			fmt.Println(field.Interface())
		}
	}
}
