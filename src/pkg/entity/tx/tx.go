package tx

/*
Decoding transaction is 2 step process:

1. get raw tx data
curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"getrawtransaction","params":["6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454"],"id":1}' http://localhost:18334

{"jsonrpc":"1.0","result":"0100000000010562fb78e4e369736e010af64b59d0637a9be2c9f04f84a5ea423e05299992df200200000000ffffffff14b06d44d8a6efba98a7ad96a1f1b61fea8c3f2d07fb0532be2ac623c02ef9640300000000ffffffff2eff276971ae1c08173d771123cf44d82d265b7bd07c9369d36b959b6f8da3bb0100000000ffffffff5b31ce928e3819ceb115ed6f277e036e195cea08ca04bba6bd4ea2b97ee24bca0300000000ffffffff5f1e949d6591ede9ca05a0bbe342339a7bba5ec1a1b0c1e937aaa2dac13dcce91000000000ffffffff05a08601000000000016001425c0ba912783971919bbc284b61f20a739782b87a086010000000000160014861c8be9238544867e95c37da4f61f61a73dda47a08601000000000016001495c6636db2ad0736a6c6f45f89e7cb4a1086d2b2a086010000000000160014acdffd0c60d0acb4da9cee54b82ae4a47cb568aba086010000000000160014b467890d42c087827dd791527ba248af4967732002473044022024a6df551024ec3dd1e9c93ab3492c48fa544ee97fea5c50328e97611a9bf1f1022071bc4aaacba73562891a5164c5be95760c2c6d62e1fdd066e6aebe1252815e69012102267c60eaa91f9eb30edd796cfa183acebbd51b3df7f6eef64b5a5761ef52dda502483045022100bd32e550ca2959769516bd56ece5c05d6af63e606d70689573205e23e348ec9502203264f03e9d19b2a1d2fe21ddd61c2cde7fd3479b9e0a8f5b93aec3f631e50e32012103b8a21dc846dd366f1524c8fa3f89ee257ddd7932c9a421b70f6e985dda2ec98c02483045022100c7a9ad39bd8bc98b13664d7e099ad6afcb58395ec13bc791e9296ba9131dfd3a02207a190acfae15c7386f6a95596a2c425c56687fa27fe34552c1c8532de748894401210364065598df65cb98bcb45f7675c31e0d6238e1fa6006b28e97d3147504001374024730440220615c29dbfc3cdb31b7a0e79f1cfb3ab33b785e360cca7c2ae6d13781c5c932d202207c4e9a44b60984177cf7f48daa4e22c0cba5c509684601e250beae007a32cf3f012103ddf05e4783cf17dcec4c6ccf40b245620ec4bb5f3269d34b0d6cedf10da11434024730440220275e63e774778abb8d3627594e39aa3f058d382bfe3b61031f95cd33d30cdeba02202fb1517dacdd3e44f6e11071a6064f51ad6c69bbe97ec1f9d485a7e5daf27cd2012103d7f98f2ce64e89f9306630e777b07de45689944a710f984389ff8e04d0de568f00000000","error":null,"id":1}


2. decode raw tx data
curl -X POST -H 'Content-Type: application/json' -u 'rpcuser:rpcpass' -d '{"jsonrpc":"1.0","method":"decoderawtransaction","params":["0100000000010562fb78e4e369736e010af64b59d0637a9be2c9f04f84a5ea423e05299992df200200000000ffffffff14b06d44d8a6efba98a7ad96a1f1b61fea8c3f2d07fb0532be2ac623c02ef9640300000000ffffffff2eff276971ae1c08173d771123cf44d82d265b7bd07c9369d36b959b6f8da3bb0100000000ffffffff5b31ce928e3819ceb115ed6f277e036e195cea08ca04bba6bd4ea2b97ee24bca0300000000ffffffff5f1e949d6591ede9ca05a0bbe342339a7bba5ec1a1b0c1e937aaa2dac13dcce91000000000ffffffff05a08601000000000016001425c0ba912783971919bbc284b61f20a739782b87a086010000000000160014861c8be9238544867e95c37da4f61f61a73dda47a08601000000000016001495c6636db2ad0736a6c6f45f89e7cb4a1086d2b2a086010000000000160014acdffd0c60d0acb4da9cee54b82ae4a47cb568aba086010000000000160014b467890d42c087827dd791527ba248af4967732002473044022024a6df551024ec3dd1e9c93ab3492c48fa544ee97fea5c50328e97611a9bf1f1022071bc4aaacba73562891a5164c5be95760c2c6d62e1fdd066e6aebe1252815e69012102267c60eaa91f9eb30edd796cfa183acebbd51b3df7f6eef64b5a5761ef52dda502483045022100bd32e550ca2959769516bd56ece5c05d6af63e606d70689573205e23e348ec9502203264f03e9d19b2a1d2fe21ddd61c2cde7fd3479b9e0a8f5b93aec3f631e50e32012103b8a21dc846dd366f1524c8fa3f89ee257ddd7932c9a421b70f6e985dda2ec98c02483045022100c7a9ad39bd8bc98b13664d7e099ad6afcb58395ec13bc791e9296ba9131dfd3a02207a190acfae15c7386f6a95596a2c425c56687fa27fe34552c1c8532de748894401210364065598df65cb98bcb45f7675c31e0d6238e1fa6006b28e97d3147504001374024730440220615c29dbfc3cdb31b7a0e79f1cfb3ab33b785e360cca7c2ae6d13781c5c932d202207c4e9a44b60984177cf7f48daa4e22c0cba5c509684601e250beae007a32cf3f012103ddf05e4783cf17dcec4c6ccf40b245620ec4bb5f3269d34b0d6cedf10da11434024730440220275e63e774778abb8d3627594e39aa3f058d382bfe3b61031f95cd33d30cdeba02202fb1517dacdd3e44f6e11071a6064f51ad6c69bbe97ec1f9d485a7e5daf27cd2012103d7f98f2ce64e89f9306630e777b07de45689944a710f984389ff8e04d0de568f00000000"],"id":1}' http://localhost:18334 | jq
{
    "txid": "6dcf241891cd43d3508ef6ee8f260fe5a9f3b0337f83874c4123bf6eb2c17454",
    "version": 1,
    "locktime": 0,
    "vin": [
      {
        "txid": "20df929929053e42eaa5844ff0c9e29b7a63d0594bf60a016e7369e3e478fb62",
        "vout": 2,
        "scriptSig": {
          "asm": "",
          "hex": ""
        },
        "txinwitness": [
          "3044022024a6df551024ec3dd1e9c93ab3492c48fa544ee97fea5c50328e97611a9bf1f1022071bc4aaacba73562891a5164c5be95760c2c6d62e1fdd066e6aebe1252815e6901",
          "02267c60eaa91f9eb30edd796cfa183acebbd51b3df7f6eef64b5a5761ef52dda5"
        ],
        "sequence": 4294967295
      },
      {
        "txid": "64f92ec023c62abe3205fb072d3f8cea1fb6f1a196ada798baefa6d8446db014",
        "vout": 3,
        "scriptSig": {
          "asm": "",
          "hex": ""
        },
        "txinwitness": [
          "3045022100bd32e550ca2959769516bd56ece5c05d6af63e606d70689573205e23e348ec9502203264f03e9d19b2a1d2fe21ddd61c2cde7fd3479b9e0a8f5b93aec3f631e50e3201",
          "03b8a21dc846dd366f1524c8fa3f89ee257ddd7932c9a421b70f6e985dda2ec98c"
        ],
        "sequence": 4294967295
      },
      {
        "txid": "bba38d6f9b956bd369937cd07b5b262dd844cf2311773d17081cae716927ff2e",
        "vout": 1,
        "scriptSig": {
          "asm": "",
          "hex": ""
        },
        "txinwitness": [
          "3045022100c7a9ad39bd8bc98b13664d7e099ad6afcb58395ec13bc791e9296ba9131dfd3a02207a190acfae15c7386f6a95596a2c425c56687fa27fe34552c1c8532de748894401",
          "0364065598df65cb98bcb45f7675c31e0d6238e1fa6006b28e97d3147504001374"
        ],
        "sequence": 4294967295
      },
      {
        "txid": "ca4be27eb9a24ebda6bb04ca08ea5c196e037e276fed15b1ce19388e92ce315b",
        "vout": 3,
        "scriptSig": {
          "asm": "",
          "hex": ""
        },
        "txinwitness": [
          "30440220615c29dbfc3cdb31b7a0e79f1cfb3ab33b785e360cca7c2ae6d13781c5c932d202207c4e9a44b60984177cf7f48daa4e22c0cba5c509684601e250beae007a32cf3f01",
          "03ddf05e4783cf17dcec4c6ccf40b245620ec4bb5f3269d34b0d6cedf10da11434"
        ],
        "sequence": 4294967295
      },
      {
        "txid": "e9cc3dc1daa2aa37e9c1b0a1c15eba7b9a3342e3bba005cae9ed91659d941e5f",
        "vout": 16,
        "scriptSig": {
          "asm": "",
          "hex": ""
        },
        "txinwitness": [
          "30440220275e63e774778abb8d3627594e39aa3f058d382bfe3b61031f95cd33d30cdeba02202fb1517dacdd3e44f6e11071a6064f51ad6c69bbe97ec1f9d485a7e5daf27cd201",
          "03d7f98f2ce64e89f9306630e777b07de45689944a710f984389ff8e04d0de568f"
        ],
        "sequence": 4294967295
      }
    ],
    "vout": [
      {
        "value": 0.001,
        "n": 0,
        "scriptPubKey": {
          "asm": "0 25c0ba912783971919bbc284b61f20a739782b87",
          "hex": "001425c0ba912783971919bbc284b61f20a739782b87",
          "reqSigs": 1,
          "type": "witness_v0_keyhash",
          "addresses": [
            "tb1qyhqt4yf8swt3jxdmc2ztv8eq5uuhs2u85ja2pc"
          ]
        }
      },
      {
        "value": 0.001,
        "n": 1,
        "scriptPubKey": {
          "asm": "0 861c8be9238544867e95c37da4f61f61a73dda47",
          "hex": "0014861c8be9238544867e95c37da4f61f61a73dda47",
          "reqSigs": 1,
          "type": "witness_v0_keyhash",
          "addresses": [
            "tb1qscwgh6frs4zgvl54cd76faslvxnnmkj8vnkmk4"
          ]
        }
      },
      {
        "value": 0.001,
        "n": 2,
        "scriptPubKey": {
          "asm": "0 95c6636db2ad0736a6c6f45f89e7cb4a1086d2b2",
          "hex": "001495c6636db2ad0736a6c6f45f89e7cb4a1086d2b2",
          "reqSigs": 1,
          "type": "witness_v0_keyhash",
          "addresses": [
            "tb1qjhrxxmdj45rndfkx730cne7tfgggd54jurs4xa"
          ]
        }
      },
      {
        "value": 0.001,
        "n": 3,
        "scriptPubKey": {
          "asm": "0 acdffd0c60d0acb4da9cee54b82ae4a47cb568ab",
          "hex": "0014acdffd0c60d0acb4da9cee54b82ae4a47cb568ab",
          "reqSigs": 1,
          "type": "witness_v0_keyhash",
          "addresses": [
            "tb1q4n0l6rrq6zktfk5uae2ts2hy537t269tr58t8q"
          ]
        }
      },
      {
        "value": 0.001,
        "n": 4,
        "scriptPubKey": {
          "asm": "0 b467890d42c087827dd791527ba248af49677320",
          "hex": "0014b467890d42c087827dd791527ba248af49677320",
          "reqSigs": 1,
          "type": "witness_v0_keyhash",
          "addresses": [
            "tb1qk3ncjr2zczrcylwhj9f8hgjg4aykwueqnk5yyj"
          ]
        }
      }
    ]
}

*/

type Transaction struct {
	Txid     string `json:"txid"`
	Version  int    `json:"version"`
	Locktime int    `json:"locktime"`
	Vin      []Vin  `json:"vin"`
	Vout     []Vout `json:"vout"`
}

type Vin struct {
	Txid        string    `json:"txid"`
	Vout        int       `json:"vout"`
	ScriptSig   ScriptSig `json:"scriptSig"`
	Txinwitness []string  `json:"txinwitness"`
	Sequence    uint64    `json:"sequence"`
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
