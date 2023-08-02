package client

import (
	"encoding/json"
	"fmt"

	"github.com/1F47E/go-feesh/pkg/entity/btc/txpool"
	log "github.com/1F47E/go-feesh/pkg/logger"
)

// rawmempool request list of tx
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawmempool","params":[],"id":1}' http://localhost:18334
// NOTE: In order to have close to realtime mempool info bitcoin node should be patched.
// By default getrawmempool by default returns unsorted list of transactions.

// custom response format with additional data
/*
{
    "txid": "0798ca60f8e42bc8ca4bf38b3449f05f6605136a30991e3962657b33fd2b035f",
    "time": 1690167042,
    "weight": 544,
    "fee": 496,
    "fee_kb": 3647
  }
*/
func (c *Client) RawMempool() ([]txpool.TxPool, error) {
	r := NewRPCRequest("getrawmempool", []interface{}{})
	data, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}
	// check type of result
	if _, ok := data.Result.([]interface{}); !ok {
		log.Log.Debugf("rawmempool result type: %T\n", data.Result)
		log.Log.Debugf("rawmempool result: %+v\n", data.Result)
		return nil, fmt.Errorf("unexpected type for result")
	}
	// Convert back to raw JSON
	rawJson, err := json.Marshal(data.Result)
	if err != nil {
		return nil, err
	}

	var ret []txpool.TxPool
	err = json.Unmarshal(rawJson, &ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// rawmempool request extended
// curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawmempool","params":[true],"id":1}' http://localhost:18334
// NOTE: takes a long time. 1+ min for the pool of 80k txs
func (c *Client) RawMempoolVerbose() ([]txpool.TxPoolVerbose, error) {
	// extended
	r := NewRPCRequest("getrawmempool", []interface{}{true})
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

	var resp map[string]txpool.TxPoolVerbose
	err = json.Unmarshal(rawJson, &resp)
	if err != nil {
		return nil, err
	}

	res := make([]txpool.TxPoolVerbose, 0)
	log.Log.Debugf("raw mempool transactions found %d\n", len(resp))
	for k, v := range resp {
		v.Hash = k
		// log.Printf("txid: %s, fee: %f\n", k, v.Fee)
		res = append(res, v)
	}
	return res, nil
}
