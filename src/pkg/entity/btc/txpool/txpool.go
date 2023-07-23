package txpool

// struct to parse response from rawmempool true (verbose)
// with simplified fields
// time field is only avaiable via this method.
// if we parse just txid and then get tx via getrawtransaction - there is no time field
// So basically doint pool parsing with verbose mode to have ordered pool list of txs
// Also having fee is good
/*
{
    "size": 219,
    "vsize": 219,
    "weight": 876,
    "fee": 0.000219,
    "time": 1690133895,
    "height": 2443385,
    "startingpriority": 0,
    "currentpriority": 0,
    "depends": [
      "89c4151288c2c4a48d01752a66d5d7dbe210bb5c097b3a95a1a1be04451871a1"
    ]
  }
*/
type TxPool struct {
	Hash         string   `json:"hash"` // hash is the key from the response
	Size         int      `json:"size"`
	VSize        int      `json:"vsize"`
	Weight       int      `json:"weight"`
	Fee          string   `json:"fee"`
	Time         int64    `json:"time"`
	Height       int      `json:"height"`
	StartingPrio int      `json:"startingpriority"`
	CurrentPrio  int      `json:"currentpriority"`
	Depends      []string `json:"depends"`
}
