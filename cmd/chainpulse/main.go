package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: chainpulse <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		os.Exit(check(os.Args[2:]))

	case "version":
		fmt.Printf("ChainPulse %s\n", version)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
