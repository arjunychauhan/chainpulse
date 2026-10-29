package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/arjunychauhan/chainpulse/internal/config"
	client "github.com/arjunychauhan/chainpulse/internal/rpc"
)

type SuccessResult struct {
	RPCName        string
	ChainID        uint64
	BlockNumber    uint64
	ChainIDLatency time.Duration
	BlockLatency   time.Duration
	TotalLatency   time.Duration
}

type FailedResult struct {
	RPCName string
	Error   error
}

type RPCResult struct {
	Success *SuccessResult
	Failed  *FailedResult
}

type RPCCheckResult struct {
	Index  int
	Result *RPCResult
}

type RPCCheckFunc func(
	rpcURL string,
	timeout time.Duration,
	name string,
) (*SuccessResult, error)

func check(args []string) int {
	return checkWithRPC(args, callRPCClient)
}

func callRPCClient(rpcURL string, timeout time.Duration, name string) (*SuccessResult, error) {
	rpcClient := client.NewClient(rpcURL)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	chainID, chainLatency, err := rpcClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	blockNumber, blockLatency, err := rpcClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	return &SuccessResult{
		ChainID:        chainID,
		BlockNumber:    blockNumber,
		BlockLatency:   blockLatency,
		ChainIDLatency: chainLatency,
		TotalLatency:   blockLatency + chainLatency,
		RPCName:        name,
	}, nil
}

func printRPCResult(result *RPCResult, resultLag *LagResult) {
	if result.Failed != nil {
		printFailedResult(result.Failed)
		return
	}

	if result.Success != nil && resultLag != nil {
		printSuccessResult(result.Success, resultLag)
	}
}

func printSuccessResult(data *SuccessResult, resultLag *LagResult) {
	fmt.Println("RPC: ", data.RPCName)
	fmt.Println("Chain ID: ", data.ChainID)
	fmt.Println("Block Number: ", data.BlockNumber)
	fmt.Println("Chain Latency: ", data.ChainIDLatency)
	fmt.Println("Block Latency: ", data.BlockLatency)
	fmt.Println("Total Latency: ", data.TotalLatency)

	if resultLag != nil {
		fmt.Println("Lag:", resultLag.Lag)
		fmt.Println("Status:", resultLag.Status)
	}

	fmt.Println("------------------------------------------")
}

func printFailedResult(result *FailedResult) {
	fmt.Println("RPC:", result.RPCName)
	fmt.Println("Status: FAILED")
	fmt.Println("Error:", result.Error)
	fmt.Println("------------------------------------------")
}

func checkRPCs(
	rpcs []config.RPCConfig,
	timeout time.Duration,
	checkRPC RPCCheckFunc,
) ([]*RPCResult, error) {
	resultsChan := make(chan RPCCheckResult, len(rpcs))

	for i, rpc := range rpcs {
		go func(i int, rpc config.RPCConfig) {
			data, err := checkRPC(rpc.URL, timeout, rpc.Name)

			if err != nil {
				resultsChan <- RPCCheckResult{
					Index: i,
					Result: &RPCResult{
						Failed: &FailedResult{
							RPCName: rpc.Name,
							Error:   err,
						},
					},
				}
				return
			}

			resultsChan <- RPCCheckResult{
				Index: i,
				Result: &RPCResult{
					Success: data,
				},
			}
		}(i, rpc)
	}

	results := make([]*RPCResult, len(rpcs))

	for i := 0; i < len(rpcs); i++ {
		result := <-resultsChan
		results[result.Index] = result.Result
	}

	return results, nil
}

func checkWithRPC(args []string, checkRPC RPCCheckFunc) int {
	checkCmd := flag.NewFlagSet("check", flag.ContinueOnError)

	// Suppress the flag package's automatic error and usage output.
	// We handle help and parsing errors explicitly below.
	checkCmd.SetOutput(io.Discard)

	rpcURL := checkCmd.String(
		"rpc",
		"",
		"EVM JSON-RPC endpoint",
	)
	configPath := checkCmd.String(
		"config",
		"",
		"Path to ChainPulse config file",
	)
	timeout := checkCmd.Duration(
		"timeout",
		10*time.Second,
		"RPC request timeout",
	)

	checkCmd.Usage = func() {
		for _, arg := range args {
			if arg == "--help" || arg == "-h" {
				fmt.Fprintln(os.Stdout, `Usage:
  chainpulse check --rpc <URL> [--timeout 10s]
  chainpulse check --config <file>

Options:`)

				checkCmd.VisitAll(func(f *flag.Flag) {
					valueType := "string"
					if f.Name == "timeout" {
						valueType = "duration"
					}

					fmt.Fprintf(os.Stdout, "  -%s %s\n        %s", f.Name, valueType, f.Usage)
					if f.Name == "timeout" {
						fmt.Fprintf(os.Stdout, " (default %s)", f.DefValue)
					}
					fmt.Fprintln(os.Stdout)
				})
				return
			}
		}
	}

	// Handle help before parsing, so Usage is called exactly once.
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			checkCmd.Usage()
			return 0
		}
	}

	if err := checkCmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	if *rpcURL == "" && *configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --rpc or --config is required")
		return 1
	}

	if *rpcURL != "" && *configPath != "" {
		fmt.Fprintln(os.Stderr, "Error: --rpc and --config cannot be used together")
		return 1
	}

	if *configPath != "" {
		cfg, err := config.Load(*configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading config:", err)
			return 1
		}

		results, err := checkRPCs(cfg.RPCs, cfg.Timeout, checkRPC)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}

		resultLags := analyzeLag(results, cfg.BlockLagThreshold)
		lagIndex := 0

		for _, result := range results {
			if result.Failed != nil {
				printRPCResult(result, nil)
				continue
			}

			printRPCResult(result, &resultLags[lagIndex])
			lagIndex++
		}

		if hasUnhealthyRPCs(results, cfg.BlockLagThreshold) {
			return 1
		}

		return 0
	}

	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "Error: --timeout must be greater than zero")
		return 1
	}

	if *rpcURL != "" {
		data, err := checkRPC(*rpcURL, *timeout, "No name")
		if err != nil {
			printFailedResult(&FailedResult{
				RPCName: "No name",
				Error:   err,
			})
			return 1
		}

		printSuccessResult(data, nil)
	}

	return 0
}
