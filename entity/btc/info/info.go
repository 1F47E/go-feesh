package info

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

type Info struct {
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
