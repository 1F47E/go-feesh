package config

import (
	"os"
	"strconv"

	log "github.com/1F47E/go-feesh/logger"
)

const BLOCK_SIZE = 4_000_000

type Config struct {
	RpcUser            string
	RpcPass            string
	RpcHost            string
	BtcGetblock        string // For services like GetBlock where auth token is in URL
	UseGetblock        bool   // Flag to indicate if we should use GetBlock style auth
	ApiHost            string
	RpcLimit           int // btc node config should be updated to allow more connections
	BlocksParsingDepth int
}

func NewConfig() *Config {
	// Check if BTC_GETBLOCK is set first
	btcGetblock := os.Getenv("BTC_GETBLOCK")
	useGetblock := btcGetblock != ""

	// If we're using GetBlock, we don't need RPC credentials
	var rpcUser, rpcPass, rpcHost string
	if !useGetblock {
		rpcUser = os.Getenv("RPC_USER")
		if rpcUser == "" {
			log.Log.Fatal("RPC_USER env var is required when not using BTC_GETBLOCK")
		}
		rpcPass = os.Getenv("RPC_PASS")
		if rpcPass == "" {
			log.Log.Fatal("RPC_PASS env var is required when not using BTC_GETBLOCK")
		}
		rpcHost = os.Getenv("RPC_HOST")
		if rpcHost == "" {
			log.Log.Fatal("RPC_HOST env var is required when not using BTC_GETBLOCK")
		}
	} else {
		rpcHost = btcGetblock
	}

	rpcStr := os.Getenv("RPC_LIMIT")
	if rpcStr == "" {
		log.Log.Fatal("RPC_LIMIT env var is required")
	}
	rpcLimit, err := strconv.Atoi(rpcStr)
	if err != nil {
		log.Log.Fatalf("error on parse RPC_LIMIT env var:", err)
	}
	if rpcLimit < 1 {
		log.Log.Fatal("RPC_LIMIT env var should be greater than 0")
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		log.Log.Fatal("API_HOST env var is required")
	}

	blocksDepthStr := os.Getenv("BLOCKS_PARSING_DEPTH")
	if blocksDepthStr == "" {
		log.Log.Fatal("BLOCKS_PARSING_DEPTH env var is required")
	}
	blocksDepth, err := strconv.Atoi(blocksDepthStr)
	if err != nil {
		log.Log.Fatalf("error on parse BLOCKS_PARSING_DEPTH env var:", err)
	}

	return &Config{
		RpcUser:            rpcUser,
		RpcPass:            rpcPass,
		RpcHost:            rpcHost,
		BtcGetblock:        btcGetblock,
		UseGetblock:        useGetblock,
		RpcLimit:           rpcLimit,
		ApiHost:            apiHost,
		BlocksParsingDepth: blocksDepth,
	}
}
