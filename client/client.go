package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/1F47E/go-feesh/entity/btc/block"
	"github.com/1F47E/go-feesh/entity/btc/info"
	"github.com/1F47E/go-feesh/entity/btc/peer"
	"github.com/1F47E/go-feesh/entity/btc/tx"
	log "github.com/1F47E/go-feesh/logger"
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
	client      *http.Client
	host        string
	user        string
	password    string
	useGetblock bool
	retries     int
}

func NewClient(host, user, password string) (*Client, error) {
	// Check if the host is non-empty
	if host == "" {
		return nil, fmt.Errorf("RPC_HOST or BTC_GETBLOCK env var must be set")
	}

	// Check if this is a GetBlock-style URL (token in URL)
	useGetblock := false
	if host != "" && (user == "" || password == "") {
		// If host is provided but credentials are missing,
		// assume token is in the URL (GetBlock style)
		useGetblock = true
	}

	return &Client{
		client: &http.Client{
			Timeout: time.Second * 10, // getrawmempool verbose can take a long fucking time
		},
		host:        host,
		user:        user,
		password:    password,
		useGetblock: useGetblock,
		retries:     10,
	}, nil
}

func (c *Client) doRequest(r *RPCRequest) (*RPCResponse, error) {
	l := log.Log.WithField("context", "[RPC]")
	jr, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(jr)

	req, err := http.NewRequest(http.MethodPost, c.host, bodyReader)
	if err != nil {
		return nil, err
	}

	// Only set basic auth if we're not using a GetBlock-style URL
	if !c.useGetblock {
		req.SetBasicAuth(c.user, c.password)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	resp := new(http.Response)
	for i := 0; i <= c.retries; i++ {
		resp, err = c.client.Do(req)
		if err != nil {
			// retry on 5xx
			if resp != nil && resp.StatusCode >= 500 {
				l.Errorf("5xx error: %s", resp.Status)
				// sleep randomly to not overload the node. Parsing the pool can be 100k+ items.
				// TODO: make rate limiter
				time.Sleep(time.Duration(100*i+rand.Intn(300)) * time.Millisecond)
				continue
			}
			log.Log.Errorf("fatal error: %s", err.Error())
			return nil, err
		}
		defer resp.Body.Close()
		break
	}

	// debug
	// read response to bytes
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Log.Errorf("RPC cli reading body err: %s", err.Error())
		return nil, err
	}

	var ret RPCResponse
	err = json.Unmarshal(data, &ret)

	if err != nil {
		log.Log.Errorf("RPC cli parsing json err: %s\nbody data: %s", err.Error(), string(data))
		return nil, err
	}

	return &ret, nil
}

// getinfo request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) GetInfo() (*info.Info, error) {
	r := NewRPCRequest("getinfo", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}

	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		return nil, err
	}

	// parse into struct
	info := new(info.Info)
	err = json.Unmarshal(rawJson, &info)
	if err != nil {
		return nil, err
	}
	// utils.PrintStruct(info)
	return info, nil
}

// get best block
// curl: curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getbestblock","params":[],"id":1}' http://localhost:18334
/*
{
  "hash": "0000000000000013d40d7e4cfd271c223c93c134065e3fc857a3adf077da3dda",
  "height": 2443258
}
*/
type ResponseGetBestBlock struct {
	Hash   string `json:"hash"`
	Height int    `json:"height"`
}

func (c *Client) GetBestBlock() (*ResponseGetBestBlock, error) {
	r := NewRPCRequest("getbestblock", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}

	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Log.Errorf("Error marshalling back to raw JSON: %v", err)
		return nil, err
	}

	// parse into struct
	var info ResponseGetBestBlock
	err = json.Unmarshal(rawJson, &info)
	if err != nil {
		log.Log.Errorf("error unmarshalling response: %v", err)
		return nil, err
	}
	return &info, nil
}

// get block header by hash
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getblockheader","params":["0000000000000013d40d7e4cfd271c223c93c134065e3fc857a3adf077da3dda"],"id":1}' http://localhost:18334
/*
{
  "hash": "0000000000000013d40d7e4cfd271c223c93c134065e3fc857a3adf077da3dda",
  "confirmations": 1,
  "height": 2443258,
  "version": 551550976,
  "versionHex": "20e00000",
  "merkleroot": "028512b4b67ecea72c6c302490baae2a38b194091c94877b0f9e5a160efe0310",
  "time": 1690059411,
  "nonce": 1820305450,
  "bits": "192495f8",
  "difficulty": 117392538.8721802,
  "previousblockhash": "00000000000019b5d7df02caed57469c2fd082aa78f975de6379e2d9500f8234"
}
*/

func (c *Client) GetBlockHeader(hash string) (*block.Block, error) {
	r := NewRPCRequest("getblockheader", []interface{}{hash})
	data, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}

	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Log.Errorf("Error marshalling back to raw JSON: %v\n", err)
		return nil, err
	}

	// parse into struct
	ret := new(block.Block)
	err = json.Unmarshal(rawJson, ret)
	if err != nil {
		log.Log.Errorf("error unmarshalling response: %v\n", err)
		return nil, err
	}
	return ret, nil
}

// get block by hash
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getblock","params":["0000000000000013d40d7e4cfd271c223c93c134065e3fc857a3adf077da3dda"],"id":1}' http://localhost:18334
/*
// NOTE: same as block header but with tx list
{
  "hash": "0000000000000013d40d7e4cfd271c223c93c134065e3fc857a3adf077da3dda",
  "confirmations": 1,
  "strippedsize": 4685,
  "size": 8043,
  "weight": 22098,
  "height": 2443258,
  "version": 551550976,
  "versionHex": "20e00000",
  "merkleroot": "028512b4b67ecea72c6c302490baae2a38b194091c94877b0f9e5a160efe0310",
  "tx": [
    "0dfe7eff14bd3722eb75f98c623bfc8b8d98036314bb5761c3384323de28dfd8",
    "b24f7ec5145310f66d0832f77f366066fe153ce9a4c477f625406d7285594c50",
    "2f507b7c7ce5cd16b9179187c25d59815894e2d4ff1839f32f9cbfb1011b633a",
    "a994932a6c31c84d862f12dfa3060266519e6eca7949d994dad07c8baa5378b3",
    "72ebc76ee8ae9761494b0a2b240a3f9234f2f4f88ef864c4127b66ab8fb6b90b",
    "46bad08470906ef2bb10f22121647f21520bc4f4141d40328706569f04a0ec47",
    "694f407557696ab70f3bdb2a193e796425ae64b04caa4045675da10e3b9fd8fa",
    "29bc294896110f683845833b1f3112870ed6feb4d5e446cd8c9c2364d5076548",
    "71310bf10899b7d03a6a6b85f3b4747c67625c4505d09c837663df6a66f14b70"
  ],
  "time": 1690059411,
  "nonce": 1820305450,
  "bits": "192495f8",
  "difficulty": 117392538.8721802,
  "previousblockhash": "00000000000019b5d7df02caed57469c2fd082aa78f975de6379e2d9500f8234"
}
*/
func (c *Client) GetBlock(blockHash string) (*block.Block, error) {
	r := NewRPCRequest("getblock", []interface{}{blockHash})
	data, err := c.doRequest(r)
	if err != nil {
		log.Log.Errorf("error doing request: %v\n", err)
		return nil, err
	}
	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		return nil, err
	}

	// parse into struct
	ret := new(block.Block)
	err = json.Unmarshal(rawJson, ret)
	if err != nil {
		return nil, err
	}
	// utils.PrintStruct(ret)
	return ret, nil
}

// get transaction
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawtransaction","params":["6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"],"id":1}' http://localhost:18334
func (c *Client) TransactionGet(txid string) (*tx.Transaction, error) {
	if len(txid) != 64 {
		return nil, fmt.Errorf("TransactionGet invalid txid")
	}
	p1 := []interface{}{string(txid)}
	p2 := []interface{}{1}
	params := append(p1, p2...)

	r := NewRPCRequest("getrawtransaction", params)
	data, err := c.doRequest(r)
	if err != nil {
		return nil, fmt.Errorf("error on getrawtransaction: %v", err)
	}
	// check type of result is string
	if _, ok := data.Result.(map[string]interface{}); !ok {
		log.Log.Errorf("data.Result: %v\n", data.Result)
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		return nil, fmt.Errorf("error marshalling back to raw JSON: %v", err)
	}

	// parse into struct
	var resp tx.Transaction
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		log.Log.Errorf("error unmarshalling response: %v\n", err)
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}
	return &resp, nil
}

// func (c *Client) TransactionGetVin(t *tx.Transaction) (uint64, error) {
// 	// ===== find out amount from vin tx matching by vout index
// 	var in uint64
// 	for _, vin := range t.Vin {
// 		// mined
// 		if vin.Coinbase != "" {
// 			continue
// 		}
// 		txIn, err := c.TransactionGet(vin.Txid)
// 		if err != nil {
// 			log.Log.Errorf("error getting vin tx: %v\n", err)
// 			return 0, err
// 		}
// 		for _, vout := range txIn.Vout {
// 			if vout.N != vin.Vout {
// 				continue
// 			}
// 			in += uint64(vout.Value * 1_0000_0000)
// 		}
// 	}
// 	return in, nil
// }

// decode raw transaction
func (c *Client) TransactionDecode(txdata string) (*tx.Transaction, error) {
	r := NewRPCRequest("decoderawtransaction", []interface{}{txdata})
	data, err := c.doRequest(r)
	if err != nil {
		log.Log.Errorf("error doing request: %v\n", err)
		return nil, err
	}
	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		log.Log.Errorf("Error marshalling back to raw JSON: %v\n", err)
		return nil, err
	}

	// parse into struct
	var resp tx.Transaction
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		log.Log.Errorf("error unmarshalling response: %v", err)
		return nil, err
	}
	return &resp, nil
}

// get peer info
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getpeerinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) GetPeers() ([]*peer.Peer, error) {
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
