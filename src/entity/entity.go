package entity

// getinfo

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

// GET BLOCK
/*
curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getblock","params":["00000000000000048e1b327dd79f72fab6395cc09a049e54fe2c0b90aa837914"],"id":1}' http://localhost:18334

{
  "hash": "00000000000000048e1b327dd79f72fab6395cc09a049e54fe2c0b90aa837914",
  "confirmations": 1,
  "strippedsize": 471,
  "size": 507,
  "weight": 1920,
  "height": 2441511,
  "version": 538968064,
  "versionHex": "20200000",
  "merkleroot": "cb391927c3107c15fb401ae8b47706ca4a829bfc60c49b23300b6ba57b81ae90",
  "tx": [
    "8bfcb7d438a438f13c2c2adb88d7785ff5178d3dc7bd4fa0d3c876457ec96793",
    "f1c2f32d50eb242679d0a09bc4ba67721cb5feee09650e734eb8cb5c9e913e95"
  ],
  "time": 1689174691,
  "nonce": 2672535381,
  "bits": "192495f8",
  "difficulty": 117392538.8721802,
  "previousblockhash": "000000000000000353b4f3f4d4e354607c6bf8f43483421e498ef7b947a39b85"
}
*/

type Block struct {
	Hash              string   `json:"hash"`
	Confirmations     int      `json:"confirmations"`
	Strippedsize      int      `json:"strippedsize"`
	Size              int      `json:"size"`
	Weight            int      `json:"weight"`
	Height            int      `json:"height"`
	Version           int      `json:"version"`
	VersionHex        string   `json:"versionHex"`
	Merkleroot        string   `json:"merkleroot"`
	Tx                []string `json:"tx"`
	Time              int      `json:"time"`
	Nonce             int      `json:"nonce"`
	Bits              string   `json:"bits"`
	Difficulty        float64  `json:"difficulty"`
	Previousblockhash string   `json:"previousblockhash"`
}
