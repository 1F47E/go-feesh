![feesh](assets/banner.jpg)

## DEMO
https://demo.feesh.io/


## ENVS:
```                                           
export RPC_USER='rpcuser'
export RPC_PASS='rpcpass'
export RPC_HOST='http://localhost:18334'
export RPC_LIMIT=420
export API_HOST='localhost:8080'
```                                           

## System requierments
```
735 Gb of space (as of 8.08.2023)
8 Gb RAM
Go 1.17
```

## How to install (patched btcd version)
```
Install patched btcd node client https://github.com/1F47E/btcd-fork2
Edit config to have index=1 if you want to get info about all tx, not only mempool

Testnet
/root/go/bin/btcd -u rpcuser -P rpcpass --testnet

Mainnet
/root/go/bin/btcd -u rpcuser -P rpcpass

Full mainnet sync can take up to a week.
```

## rawmempool patch
```
In order to have realtime mempool info with tx fee right from the node - it should be patched.
By default getrawmempool returns unsorted list of transactions, hash only.
After patch it will return full tx info in a sorted array by time.
```





## TODO
- [ ] Add more block stats
- [ ] history pool data
- [ ] pool tx update via websocket
- [x] basic pool frontend 
- [x] API pool
- [x] websockets

## react front end (WIP)
![feesh react front end](https://github.com/1F47E/react-feesh/raw/master/assets/screenshot.png)



```
            _____ _____ _____ _____ _____ 
           |   __|   __|   __|   __|  |  |
           |   __|   __|   __|__   |     |
           |__|  |_____|_____|_____|__|__|
           bitcoin mempool stats


```      
