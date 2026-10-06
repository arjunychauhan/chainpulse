package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	client "github.com/arjunychauhan/chainpulse/internal/rpc"
)

func check(args []string) {
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)

	rpcURL := checkCmd.String(
		"rpc",
		"",
		"EVM JSON-RPC endpoint",
	)
	timeout := checkCmd.Duration(
		"timeout",
		10*time.Second,
		"RPC request timeout",
	)

	checkCmd.Parse(args)

	if *rpcURL == "" {
		fmt.Println("Error: --rpc is required")
		os.Exit(1)
	}
	if *timeout <= 0 {
		fmt.Println("Error: --timeout must be greater than zero")
		os.Exit(1)
	}

	rpcClient := client.NewClient(*rpcURL)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		*timeout,
	)
	defer cancel()

	chainID, chainLatency, err := rpcClient.ChainID(ctx)
	if err != nil {
		fmt.Println("Error:", err, "chainLatency", chainLatency)
		os.Exit(1)
	}

	fmt.Println("Chain ID:", chainID)
	fmt.Println("Chain ID Latency:", chainLatency)

	blockNumber, blockLatency, err := rpcClient.BlockNumber(ctx)
	if err != nil {
		fmt.Println("Error:", err, "blockLatency", blockLatency)
		os.Exit(1)
	}
	fmt.Println("Block Number:", blockNumber)
	fmt.Println("Block Number Latency:", blockLatency)
	fmt.Println("RPC:", *rpcURL)
	fmt.Println("Total RPC Latency", chainLatency+blockLatency)
}
