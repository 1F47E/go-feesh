package peer

/*
	{
	      "id": 25,
	      "addr": "95.216.147.146:18333",
	      "addrlocal": "10.1.54.29:58486",
	      "services": "00001037",
	      "relaytxes": true,
	      "lastsend": 1689179306,
	      "lastrecv": 1689179304,
	      "bytessent": 5771,
	      "bytesrecv": 20164,
	      "conntime": 1689178790,
	      "timeoffset": 0,
	      "pingtime": 380349,
	      "version": 70015,
	      "subver": "/Satoshi:0.17.1/",
	      "inbound": false,
	      "startingheight": 2441519,
	      "currentheight": 2441519,
	      "banscore": 0,
	      "feefilter": 1000,
	      "syncnode": false
	    }
*/
type Peer struct {
	ID             int64  `json:"id"`
	Addr           string `json:"addr"`
	AddrLocal      string `json:"addrlocal"`
	Services       string `json:"services"`
	RelayTxes      bool   `json:"relaytxes"`
	LastSend       int64  `json:"lastsend"`
	LastRecv       int64  `json:"lastrecv"`
	BytesSent      int64  `json:"bytessent"`
	BytesRecv      int64  `json:"bytesrecv"`
	ConnTime       int64  `json:"conntime"`
	TimeOffset     int64  `json:"timeoffset"`
	PingTime       int64  `json:"pingtime"`
	Version        int64  `json:"version"`
	SubVer         string `json:"subver"`
	Inbound        bool   `json:"inbound"`
	StartingHeight int64  `json:"startingheight"`
	CurrentHeight  int64  `json:"currentheight"`
	BanScore       int64  `json:"banscore"`
	FeeFilter      int64  `json:"feefilter"`
	SyncNode       bool
}
