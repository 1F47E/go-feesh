
```
            _____ _____ _____ _____ _____ 
           |   __|   __|   __|   __|  |  |
           |   __|   __|   __|__   |     |
           |__|  |_____|_____|_____|__|__|
           bitcoin mempool stats


```                                           

NOTE: In order to have close to realtime mempool info bitcoin node should be patched.

By default getrawmempool by default returns unsorted list of transactions.

## DEMO
https://demo.feesh.io/

# ENVS:
```                                           
export RPC_USER='rpcuser'
export RPC_PASS='rpcpass'
export RPC_HOST='http://localhost:18334'
export RPC_LIMIT=420
export API_HOST='localhost:8080'
```                                           




## TODO
- [ ] Add more block stats
- [ ] history data
- [ ] frontend
- [x] API pool
- [x] websockets

WIP
![feesh react front end](https://github.com/1F47E/react-feesh/raw/master/assets/screenshot.png)
