package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: chainpulse <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		check(os.Args[2:])

	case "version":
		fmt.Println("ChainPulse v0.1.0")

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
