# ChainPulse

A lightweight CLI tool for monitoring and troubleshooting EVM-compatible blockchain RPC endpoints.

## Status

🚧 **Early development**

ChainPulse is currently being developed as a practical Go project focused on learning and building production-style blockchain infrastructure tooling.

## Current Features

* Check EVM chain ID
* Check latest block number
* JSON-RPC communication over HTTP
* Request timeout support
* Context-aware RPC requests
* RPC error handling
* Unit tests using `httptest`

## Usage

### Check an RPC endpoint

```bash
go run ./cmd/chainpulse check --rpc https://your-rpc-endpoint --timeout 10s
```

Example:

```bash
go run ./cmd/chainpulse check --rpc http://localhost:8545 --timeout 10s
```

Example output:

```text
Chain ID: 8453
Block Number: 51902461
RPC: https://your-rpc-endpoint
```

## Project Structure

```text
chainpulse/
├── cmd/
│   └── chainpulse/
│       └── main.go
├── internal/
│   └── rpc/
│       ├── client.go
│       └── client_test.go
├── .gitignore
├── go.mod
└── README.md
```

## Development

Run the tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./internal/rpc
```

Check test coverage:

```bash
go test -cover ./internal/rpc
```

Build the CLI:

```bash
go build ./cmd/chainpulse
```

## Roadmap

Planned features include:

* [ ] Multiple RPC endpoint monitoring
* [ ] Block height and block lag detection
* [ ] Continuous RPC health monitoring
* [ ] Response latency measurement
* [ ] Prometheus metrics
* [ ] Webhook alerts
* [ ] REST API
* [ ] Docker support
* [ ] Production release binaries

## Why ChainPulse?

Blockchain applications often depend heavily on RPC infrastructure. A healthy RPC endpoint should not only be reachable, but should also respond correctly, remain synchronized, and provide predictable latency.

ChainPulse aims to provide a lightweight tool for checking these properties from the command line.

## License

License to be decided.
