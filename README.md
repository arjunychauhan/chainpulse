# ChainPulse

A lightweight Go CLI for monitoring and troubleshooting EVM-compatible blockchain RPC endpoints.

ChainPulse checks RPC availability, chain IDs, block heights, response latency, and block lag across multiple endpoints.

## Features

- **RPC health checks** — verify that EVM JSON-RPC endpoints respond successfully.
- **Multi-RPC checks** — check multiple endpoints concurrently using a YAML configuration file.
- **Block lag detection** — compare block heights between endpoints belonging to the same chain.
- **Configurable lag threshold** — set the maximum acceptable block lag, defaulting to 3 blocks.
- **Latency measurement** — measure chain ID and block number request latency.
- **Timeout support** — use context-aware RPC requests with configurable timeouts.
- **Failure handling** — distinguish RPC failures from endpoints that are lagging.
- **Unit tests** — test RPC behavior, configuration, concurrency, and lag analysis.

## Requirements

- Go 1.27.1 or compatible toolchain for the version specified in `go.mod`.
- Access to EVM-compatible JSON-RPC endpoints.

## Installation

Clone the repository:

```bash
git clone https://github.com/arjunychauhan/chainpulse.git
cd chainpulse
```

Run the CLI directly:

```bash
go run ./cmd/chainpulse version
```

Build the executable:

```bash
go build -o chainpulse ./cmd/chainpulse
```

## Usage

### Check a single RPC endpoint

```bash
go run ./cmd/chainpulse check \
  --rpc https://your-rpc-endpoint \
  --timeout 10s
```

Example output:

```text
RPC:  No name
Chain ID:  8453
Block Number:  51902461
Chain Latency:  20ms
Block Latency:  15ms
Total Latency:  35ms
```

The values above are illustrative.

### Check multiple RPC endpoints

Create a `chainpulse.yaml` configuration file:

```yaml
timeout: 5s
block_lag_threshold: 3

rpcs:
  - name: ethereum
    url: https://your-primary-ethereum-rpc

  - name: ethereum-backup
    url: https://your-backup-ethereum-rpc
```

Run the checks:

```bash
go run ./cmd/chainpulse check --config chainpulse.yaml
```

ChainPulse checks the configured RPCs concurrently and prints the results in configuration order.

### Configuration options

| Option                | Description                    | Default                   |
| --------------------- | ------------------------------ | ------------------------- |
| `timeout`             | Timeout for each RPC check     | Required in configuration |
| `block_lag_threshold` | Maximum acceptable block lag   | `3`                       |
| `rpcs`                | List of RPC endpoints to check | At least one required     |
| `rpcs[].name`         | Unique name for an endpoint    | Required                  |
| `rpcs[].url`          | Endpoint URL                   | Required                  |

The lag threshold is measured in blocks. A threshold of `3` means that an endpoint may be up to three blocks behind the highest block reported by successful endpoints on the same chain and still be considered healthy. An explicit threshold of `0` is supported.

Configuration validation rejects nonpositive timeouts, empty RPC lists, empty names or URLs, and duplicate RPC names.

## Block lag detection

ChainPulse groups successful RPC results by chain ID. For each chain, it finds the highest observed block number and calculates each endpoint's lag:

```text
lag = highest observed block - endpoint block number
```

For example, with a threshold of `3`:

| RPC     | Block | Lag | Status    |
| ------- | ----: | --: | --------- |
| `eth-a` |   100 |   0 | `HEALTHY` |
| `eth-b` |    98 |   2 | `HEALTHY` |
| `eth-c` |    96 |   4 | `LAGGING` |

RPCs on different chain IDs are analyzed separately.

**Important limitation:** the highest observed block is a comparison reference, not proof of the canonical chain head. An endpoint reporting an anomalously high block can distort the comparison. ChainPulse's current lag calculation does not implement quorum-based or canonical-head verification.

## Statuses and exit codes

### RPC statuses

| Status    | Meaning                                                               |
| --------- | --------------------------------------------------------------------- |
| `HEALTHY` | RPC check succeeded and block lag is within the configured threshold. |
| `LAGGING` | RPC check succeeded, but block lag exceeds the threshold.             |
| `FAILED`  | The RPC request failed.                                               |

### Exit codes

| Code | Meaning                                                                          |
| ---: | -------------------------------------------------------------------------------- |
|  `0` | All checked RPCs are healthy.                                                    |
|  `1` | At least one RPC failed, an RPC is lagging, or the command encountered an error. |

These exit codes make ChainPulse suitable for basic shell scripts and CI checks.

## Development

Run all tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Check RPC package coverage:

```bash
go test -cover ./internal/rpc
```

Build the CLI:

```bash
go build ./cmd/chainpulse
```

## Project structure

```text
chainpulse/
├── cmd/
│   └── chainpulse/
│       ├── main.go
│       ├── check.go
│       ├── check_test.go
│       ├── lag.go
│       └── lag_test.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   └── rpc/
│       ├── client.go
│       └── client_test.go
├── chainpulse.yaml
├── go.mod
├── go.sum
└── README.md
```

## Roadmap

- [x] Multiple RPC endpoint checks
- [x] Block height and block lag detection
- [x] Response latency measurement
- [ ] Continuous RPC health monitoring
- [ ] Prometheus metrics
- [ ] Webhook alerts
- [ ] REST API
- [ ] Docker support
- [ ] Production release binaries

## Why ChainPulse?

Blockchain applications depend on reliable RPC infrastructure. An endpoint being reachable does not necessarily mean it is synchronized or responsive enough for an application.

ChainPulse provides a lightweight foundation for checking RPC availability, latency, and relative block synchronization from the command line.

## License

A license has not yet been selected. Until one is added, don't assume that the repository grants permission to redistribute or modify the software.
