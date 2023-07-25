package config

import (
	"os"
	"strconv"

	log "go-btc-scan/src/pkg/logger"
)

type Config struct {
	RpcUser  string
	RpcPass  string
	RpcHost  string
	ApiHost  string
	RpcLimit int // btc node config should be updated to allow more connections
}

func NewConfig() *Config {
	rpcUser := os.Getenv("RPC_USER")
	if rpcUser == "" {
		log.Log.Fatalln("RPC_USER env var is required")
	}
	rpcPass := os.Getenv("RPC_PASS")
	if rpcPass == "" {
		log.Log.Fatalln("RPC_PASS env var is required")
	}
	rpcHost := os.Getenv("RPC_HOST")
	if rpcHost == "" {
		log.Log.Fatalln("RPC_HOST env var is required")
	}

	rpcStr := os.Getenv("RPC_LIMIT")
	if rpcStr == "" {
		log.Log.Fatalln("RPC_LIMIT env var is required")
	}
	rpcLimit, err := strconv.Atoi(rpcStr)
	if err != nil {
		log.Log.Fatalln("error on parse RPC_LIMIT env var:", err)
	}
	if rpcLimit < 1 {
		log.Log.Fatalln("RPC_LIMIT env var should be greater than 0")
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		log.Log.Fatalln("API_HOST env var is required")
	}

	return &Config{
		RpcUser:  rpcUser,
		RpcPass:  rpcPass,
		RpcHost:  rpcHost,
		RpcLimit: rpcLimit,
		ApiHost:  apiHost,
	}
}
