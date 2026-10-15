package main

import (
	"fmt"
	"io"
	"os"
)

var version = "dev"

const helpText = `ChainPulse - EVM RPC monitoring CLI

Usage:
  chainpulse <command> [options]

Commands:
  check      Check one or more EVM RPC endpoints
  version    Print the ChainPulse version
  help       Show this help message

Examples:
  chainpulse check --rpc https://your-rpc-endpoint --timeout 10s
  chainpulse check --config chainpulse.yaml
  chainpulse version
  chainpulse help

Run "chainpulse check --help" for check command options.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, helpText)
		return 1
	}

	switch command := args[0]; command {
	case "check":
		return check(args[1:])

	case "version":
		fmt.Fprintf(stdout, "ChainPulse %s\n", version)
		return 0

	case "help", "--help", "-h":
		fmt.Fprint(stdout, helpText)
		return 0

	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n\n", command)
		fmt.Fprint(stdout, helpText)
		return 1
	}
}
