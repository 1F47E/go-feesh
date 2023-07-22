package mempool

// RAW MEMPOOL RESPONSE extended
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

type MemPoolTxList struct {
  
}