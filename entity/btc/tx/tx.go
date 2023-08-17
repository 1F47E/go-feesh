package tx

/*
Decoding transaction is 2 step process:

1. get raw tx data
curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawtransaction","params":["6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454", 1],"id":1}' http://localhost:18334

{
  "hex": "01000000016e7e943d341edcd4c2425e875548d0abb1c7042e3ab28bff8deefd17d9755187030000006b483045022100dbd7d430ab70ed41e4da2c9ce98814081b3f1224e7e969017690dc40d2b48767022022925f78101669aa247d3943cb3e69ead996673f38e8ea666e8ab50d5f1d12c80121037435c194e9b01b3d7f7a2802d6684a3af68d05bbf4ec8f17021980d777691f1dfdffffff040000000000000000536a4c5054325b2066a044835c5e167dbd133f018c5f8cd536d59feaeaca86f666cc1100606a7e9f5ec16ec7cef200fe9df6d3ccad67152e42edaf8eba1e47f4174b3c0b3e2fc0002547e10005002543a7000a4b10270000000000001976a914000000000000000000000000000000000000000088ac10270000000000001976a914000000000000000000000000000000000000000088acd972cb10000000001976a914ba27f99e007c7f605a8305e318c1abde3cd220ac88ac00000000",
  "txid": "afe8727e41cfde28c9162a68bf27f9172b05a2615e34d7a6891f6f7594b21d0c",
  "hash": "afe8727e41cfde28c9162a68bf27f9172b05a2615e34d7a6891f6f7594b21d0c",
  "size": 352,
  "vsize": 352,
  "weight": 1408,
  "version": 1,
  "locktime": 0,
  "vin": [
    {
      "txid": "875175d917fdee8dff8bb23a2e04c7b1abd04855875e42c2d4dc1e343d947e6e",
      "vout": 3,
      "scriptSig": {
        "asm": "3045022100dbd7d430ab70ed41e4da2c9ce98814081b3f1224e7e969017690dc40d2b48767022022925f78101669aa247d3943cb3e69ead996673f38e8ea666e8ab50d5f1d12c801 037435c194e9b01b3d7f7a2802d6684a3af68d05bbf4ec8f17021980d777691f1d",
        "hex": "483045022100dbd7d430ab70ed41e4da2c9ce98814081b3f1224e7e969017690dc40d2b48767022022925f78101669aa247d3943cb3e69ead996673f38e8ea666e8ab50d5f1d12c80121037435c194e9b01b3d7f7a2802d6684a3af68d05bbf4ec8f17021980d777691f1d"
      },
      "sequence": 4294967293
    }
  ],
  "vout": [
    {
      "value": 0,
      "n": 0,
      "scriptPubKey": {
        "asm": "OP_RETURN 54325b2066a044835c5e167dbd133f018c5f8cd536d59feaeaca86f666cc1100606a7e9f5ec16ec7cef200fe9df6d3ccad67152e42edaf8eba1e47f4174b3c0b3e2fc0002547e10005002543a7000a4b",
        "hex": "6a4c5054325b2066a044835c5e167dbd133f018c5f8cd536d59feaeaca86f666cc1100606a7e9f5ec16ec7cef200fe9df6d3ccad67152e42edaf8eba1e47f4174b3c0b3e2fc0002547e10005002543a7000a4b",
        "type": "nulldata"
      }
    },
    {
      "value": 0.0001,
      "n": 1,
      "scriptPubKey": {
        "asm": "OP_DUP OP_HASH160 0000000000000000000000000000000000000000 OP_EQUALVERIFY OP_CHECKSIG",
        "hex": "76a914000000000000000000000000000000000000000088ac",
        "reqSigs": 1,
        "type": "pubkeyhash",
        "addresses": [
          "mfWxJ45yp2SFn7UciZyNpvDKrzbhyfKrY8"
        ]
      }
    },
    {
      "value": 0.0001,
      "n": 2,
      "scriptPubKey": {
        "asm": "OP_DUP OP_HASH160 0000000000000000000000000000000000000000 OP_EQUALVERIFY OP_CHECKSIG",
        "hex": "76a914000000000000000000000000000000000000000088ac",
        "reqSigs": 1,
        "type": "pubkeyhash",
        "addresses": [
          "mfWxJ45yp2SFn7UciZyNpvDKrzbhyfKrY8"
        ]
      }
    },
    {
      "value": 2.81768665,
      "n": 3,
      "scriptPubKey": {
        "asm": "OP_DUP OP_HASH160 ba27f99e007c7f605a8305e318c1abde3cd220ac OP_EQUALVERIFY OP_CHECKSIG",
        "hex": "76a914ba27f99e007c7f605a8305e318c1abde3cd220ac88ac",
        "reqSigs": 1,
        "type": "pubkeyhash",
        "addresses": [
          "mxVFsFW5N4mu1HPkxPttorvocvzeZ7KZyk"
        ]
      }
    }
  ],
  "blockhash": "0000000000000021a4fc5a8d7b408c9506ee832bb7e89759c7196f8df2b544d7",
  "confirmations": 3,
  "time": 1690044562,
  "blocktime": 1690044562
}

*/

type Transaction struct {
	Txid string `json:"txid"`
	// Hash          string `json:"hash"`
	Version       int    `json:"version"`
	Locktime      int    `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	Size          int    `json:"size"`
	Weight        int    `json:"weight"`
	Blockhash     string `json:"blockhash"`
	Confirmations int    `json:"confirmations"`
	Time          int    `json:"time"`
	Blocktime     int    `json:"blocktime"`
}

type Vin struct {
	Txid        string    `json:"txid"`
	Vout        int       `json:"vout"`
	ScriptSig   ScriptSig `json:"scriptSig"`
	Txinwitness []string  `json:"txinwitness"`
	Sequence    uint64    `json:"sequence"`
	Coinbase    string    `json:"coinbase"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

type ScriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	ReqSigs   int      `json:"reqSigs"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

// get total out amount
func (t *Transaction) GetTotalOut() uint64 {
	var total float64
	for _, v := range t.Vout {
		total += v.Value
	}
	return uint64(total * 1_0000_0000)
}
