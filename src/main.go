package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
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
	Jsonrpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   interface{}     `json:"error"`
}

/*
{
  "version": 230300,
  "protocolversion": 70002,
  "blocks": 2441415,
  "timeoffset": 0,
  "connections": 8,
  "proxy": "",
  "difficulty": 1,
  "testnet": true,
  "relayfee": 0.00001,
  "errors": ""
}
*/

type ResponseGetinfo struct {
	Version         int     `json:"version"`
	ProtocolVersion int     `json:"protocolversion"`
	Blocks          int     `json:"blocks"`
	Timeoffset      int     `json:"timeoffset"`
	Connections     int     `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	Testnet         bool    `json:"testnet"`
	Relayfee        float64 `json:"relayfee"`
	Errors          string  `json:"errors"`
}

func NewRPCRequest(method string, params interface{}) *RPCRequest {
	return &RPCRequest{
		Jsonrpc: "1.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}
}

// RAW MEMPOOL RESPONSE
/*
{
  "e335aa60ffb1462f8c15c6b3322d97294f9b65c8701c658b45325003b94a55e2": {
    "size": 222,
    "vsize": 141,
    "weight": 561,
    "fee": 0.00000144,
    "time": 1689116517,
    "height": 2441417,
    "startingpriority": 2788259.8453038675,
    "currentpriority": 5576519.690607735,
    "depends": []
  },
{
  "e335aa60ffb1462f8c15c6b3322d97294f9b65c8701c658b45325003b94a55e2": {
    "size": 222,
    "vsize": 141,
    "weight": 561,
    "fee": 0.00000144,
    "time": 1689116517,
    "height": 2441417,
    "startingpriority": 2788259.8453038675,
    "currentpriority": 5576519.690607735,
    "depends": []
  }
}
}
*/

type MemPoolTx struct {
	Size             int     `json:"size"`
	Vsize            int     `json:"vsize"`
	Weight           int     `json:"weight"`
	Fee              float64 `json:"fee"`
	Time             int     `json:"time"`
	Height           int     `json:"height"`
	Startingpriority float64 `json:"startingpriority"`
	Currentpriority  float64 `json:"currentpriority"`
	// Depends          []int   `json:"depends"`
}

// ===== CLIENT
type Client struct {
	client   *http.Client
	url      string
	user     string
	password string
}

func NewClient(url, user, password string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		url:      url,
		user:     user,
		password: password,
	}
}

func (c *Client) doRequest(r *RPCRequest) (*RPCResponse, error) {
	jr, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(jr)

	req, err := http.NewRequest(http.MethodPost, c.url, bodyReader)
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
	var info ResponseGetinfo
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
	var resp map[string]MemPoolTx
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

func main() {
	nodeUrl := "http://localhost:18334"
	username := "rpcuser"
	password := "rpcpass"
	c := NewClient(nodeUrl, username, password)

	err := c.getInfo()
	if err != nil {
		log.Fatalln("error on getinfo:", err)
	}
	err = c.rawMempool()
	if err != nil {
		log.Fatalln("error on rawmempool:", err)
	}
}
