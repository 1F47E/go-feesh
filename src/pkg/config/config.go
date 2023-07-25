package config

import (
	"os"
	"strconv"

	log "go-btc-scan/src/pkg/logger"
)

type Config struct {
	RpcLimit int // btc node config should be updated to allow more connections
}

func NewConfig() *Config {
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
	return &Config{
		RpcLimit: rpcLimit,
	}
}
