package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
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
	Id      interface{} `json:"id"`
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
	// GetBlock.io uses JsonRPC 2.0 format
	return &RPCRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      "getblock.io",
	}
}

// ===== CLIENT
type Client struct {
	client      *http.Client
	host        string
	user        string
	password    string
	useGetblock bool
	debug       bool
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

	// Check for RPC_DEBUG environment variable
	debug := os.Getenv("RPC_DEBUG") == "1"
	if debug {
		log.Log.Info("RPC debug logging enabled")
	}

	return &Client{
		client: &http.Client{
			Timeout: time.Second * 10, // getrawmempool verbose can take a long fucking time
		},
		host:        host,
		user:        user,
		password:    password,
		useGetblock: useGetblock,
		debug:       debug,
		retries:     10,
	}, nil
}

func (c *Client) doRequest(r *RPCRequest) (*RPCResponse, error) {
	l := log.Log.WithField("context", "[RPC]")

	if c.debug {
		l.Info("============= RPC REQUEST =============")
		l.Infof("Host: %s", c.host)
		l.Infof("Method: %s", r.Method)
		l.Infof("UseGetblock: %v", c.useGetblock)
	}

	jr, err := json.Marshal(r)
	if err != nil {
		l.Errorf("Error marshaling request: %v", err)
		return nil, err
	}

	// Debug log of the request payload
	if c.debug {
		l.Infof("Request payload: %s", string(jr))
	} else {
		l.Debugf("Sending request to %s: %s", c.host, string(jr))
	}

	bodyReader := bytes.NewReader(jr)

	req, err := http.NewRequest(http.MethodPost, c.host, bodyReader)
	if err != nil {
		l.Errorf("Error creating request: %v", err)
		return nil, err
	}

	// Only set basic auth if we're not using a GetBlock-style URL
	if !c.useGetblock {
		req.SetBasicAuth(c.user, c.password)
	}

	// Add common headers required by most APIs
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-feesh/1.0")

	// For GetBlock, add some additional headers that might be needed
	if c.useGetblock {
		// Some providers require explicitly setting these headers
		req.Header.Set("Accept", "application/json")
	}

	req.Close = true

	if c.debug {
		l.Info("============= REQUEST HEADERS =============")
		for k, v := range req.Header {
			l.Infof("%s: %v", k, v)
		}
	}

	resp := new(http.Response)
	for i := 0; i <= c.retries; i++ {
		l.Debugf("Making API request attempt %d/%d", i+1, c.retries+1)
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
			l.Errorf("Network error: %s", err.Error())
			return nil, err
		}

		// Handle 403 errors specially with more diagnostics
		if resp.StatusCode == 403 {
			l.Errorf("HTTP 403 Forbidden - API access denied. This could be due to:")
			l.Errorf("1. Invalid API key or token in URL")
			l.Errorf("2. Requested method (%s) not supported by GetBlock", r.Method)
			l.Errorf("3. IP address restrictions")

			// Try to read the response body for any helpful error messages
			responseData, readErr := io.ReadAll(resp.Body)
			if readErr == nil && len(responseData) > 0 {
				l.Errorf("Response body: %s", string(responseData))
			}
			resp.Body.Close()

			// Return a more specific error
			return nil, fmt.Errorf("API access denied (HTTP 403) when calling method: %s", r.Method)
		}

		if c.debug {
			l.Infof("Response status: %s", resp.Status)
			l.Info("============= RESPONSE HEADERS =============")
			for k, v := range resp.Header {
				l.Infof("%s: %v", k, v)
			}
		} else {
			l.Debugf("Received response with status: %s", resp.Status)
		}

		defer resp.Body.Close()
		break
	}

	// Read response to bytes
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("Error reading response body: %s", err.Error())
		return nil, err
	}

	// Debug log the response data
	if c.debug {
		l.Info("============= RESPONSE BODY =============")
		l.Info(string(data))
	} else {
		l.Debugf("Received response body: %s", string(data))
	}

	// Check if response body is empty
	if len(data) == 0 {
		l.Error("Empty response body received")
		return nil, fmt.Errorf("empty response from server")
	}

	var ret RPCResponse
	err = json.Unmarshal(data, &ret)

	if err != nil {
		l.Errorf("Error parsing JSON response: %s\nBody data: %s", err.Error(), string(data))
		return nil, err
	}

	// Check for errors in the response
	if ret.Error != nil {
		errorDetails, _ := json.Marshal(ret.Error)
		l.Errorf("RPC error in response: %s", string(errorDetails))
		return nil, fmt.Errorf("RPC error: %s", string(errorDetails))
	}

	return &ret, nil
}

// getinfo request
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getinfo","params":[],"id":1}' http://localhost:18334
func (c *Client) GetInfo() (*info.Info, error) {
	l := log.Log.WithField("context", "[RPC.GetInfo]")
	l.Debug("Getting blockchain info")

	// Try getblockchaininfo first (for newer Bitcoin Core and services like GetBlock)
	r := NewRPCRequest("getblockchaininfo", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		// If getblockchaininfo fails, try the legacy getinfo method
		l.Warnf("getblockchaininfo failed, trying legacy getinfo: %v", err)
		r = NewRPCRequest("getinfo", []interface{}{})
		data, err = c.doRequest(r)
		if err != nil {
			l.Errorf("Both getblockchaininfo and getinfo failed: %v", err)
			return nil, err
		}
	}

	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		l.Errorf("Unexpected result type: %T", data.Result)
		return nil, fmt.Errorf("unexpected type for result")
	}

	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		l.Errorf("Error marshalling result back to JSON: %v", err)
		return nil, err
	}

	// First try parsing as the standard info format
	nodeInfo := new(info.Info)
	err = json.Unmarshal(rawJson, &nodeInfo)

	if err != nil {
		// If that fails, try parsing as blockchain info and convert
		l.Warnf("Failed to parse response as Info, trying blockchain info format: %v", err)
		var blockchainInfo struct {
			Chain                string  `json:"chain"`
			Blocks               int     `json:"blocks"`
			Headers              int     `json:"headers"`
			BestBlockHash        string  `json:"bestblockhash"`
			Difficulty           float64 `json:"difficulty"`
			MedianTime           int     `json:"mediantime"`
			VerificationProgress float64 `json:"verificationprogress"`
			InitialBlockDownload bool    `json:"initialblockdownload"`
			ChainWork            string  `json:"chainwork"`
			SizeOnDisk           int64   `json:"size_on_disk"`
			Pruned               bool    `json:"pruned"`
		}

		if err := json.Unmarshal(rawJson, &blockchainInfo); err != nil {
			l.Errorf("Failed to parse response as blockchain info: %v", err)
			return nil, err
		}

		// Convert to the legacy info format
		nodeInfo = &info.Info{
			Version:         0, // Not available in blockchaininfo
			ProtocolVersion: 0, // Not available in blockchaininfo
			Blocks:          blockchainInfo.Blocks,
			Timeoffset:      0,  // Not available in blockchaininfo
			Connections:     0,  // Not available in blockchaininfo
			Proxy:           "", // Not available in blockchaininfo
			Difficulty:      blockchainInfo.Difficulty,
			Testnet:         blockchainInfo.Chain != "main",
			Relayfee:        0,  // Not available in blockchaininfo
			Errors:          "", // Not available in blockchaininfo
		}
	}

	l.Debugf("Successfully parsed info, height: %d", nodeInfo.Blocks)
	return nodeInfo, nil
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
	l := log.Log.WithField("context", "[RPC.GetBestBlock]")

	// First try getbestblockhash (widely supported including by GetBlock)
	l.Debug("Getting best block hash")
	r := NewRPCRequest("getbestblockhash", []interface{}{})
	data, err := c.doRequest(r)

	if err != nil {
		l.Warnf("getbestblockhash failed, trying fallback to getblockchaininfo: %v", err)

		// Try getblockchaininfo as fallback
		r = NewRPCRequest("getblockchaininfo", []interface{}{})
		data, err = c.doRequest(r)
		if err != nil {
			l.Errorf("Both getbestblockhash and getblockchaininfo failed: %v", err)

			// As a last resort, try the original getbestblock method
			r = NewRPCRequest("getbestblock", []interface{}{})
			data, err = c.doRequest(r)
			if err != nil {
				l.Errorf("All block query methods failed: %v", err)
				return nil, fmt.Errorf("failed to get best block info: %v", err)
			}
		}

		// If we got blockchaininfo, extract what we need
		if r.Method == "getblockchaininfo" {
			// Parse blockchaininfo
			if _, ok := data.Result.(map[string]interface{}); !ok {
				return nil, fmt.Errorf("unexpected type for result")
			}

			rawJson, _ := json.Marshal(data.Result)
			var blockchainInfo struct {
				Blocks        int    `json:"blocks"`
				BestBlockHash string `json:"bestblockhash"`
			}

			if err := json.Unmarshal(rawJson, &blockchainInfo); err != nil {
				return nil, err
			}

			// Create the response from blockchaininfo data
			return &ResponseGetBestBlock{
				Hash:   blockchainInfo.BestBlockHash,
				Height: blockchainInfo.Blocks,
			}, nil
		}
	}

	// If we got here with getbestblockhash, we need to get the height separately
	if r.Method == "getbestblockhash" {
		// We have the hash, but need to get the height
		hash, ok := data.Result.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected result type for getbestblockhash")
		}

		// Get the block header to find its height
		r = NewRPCRequest("getblockheader", []interface{}{hash})
		headerData, err := c.doRequest(r)
		if err != nil {
			return nil, fmt.Errorf("failed to get block header: %v", err)
		}

		if _, ok := headerData.Result.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("unexpected type for block header result")
		}

		rawJson, _ := json.Marshal(headerData.Result)
		var blockHeader struct {
			Height int `json:"height"`
		}

		if err := json.Unmarshal(rawJson, &blockHeader); err != nil {
			return nil, err
		}

		return &ResponseGetBestBlock{
			Hash:   hash,
			Height: blockHeader.Height,
		}, nil
	}

	// For the original getbestblock method, parse directly
	// check type of result
	if _, ok := data.Result.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		l.Errorf("Error marshalling back to raw JSON: %v", err)
		return nil, err
	}

	// parse into struct
	var info ResponseGetBestBlock
	err = json.Unmarshal(rawJson, &info)
	if err != nil {
		l.Errorf("error unmarshalling response: %v", err)
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
