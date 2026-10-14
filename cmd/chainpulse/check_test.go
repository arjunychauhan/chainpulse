package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/arjunychauhan/chainpulse/internal/config"
)

func TestCheckRPCs(t *testing.T) {
	rpcs := []config.RPCConfig{
		{
			Name: "rpc-1",
			URL:  "https://rpc-1.example.com",
		},
		{
			Name: "rpc-2",
			URL:  "https://rpc-2.example.com",
		},
		{
			Name: "rpc-3",
			URL:  "https://rpc-3.example.com",
		},
	}

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		if name == "rpc-2" {
			return nil, fmt.Errorf("connection failed")
		}

		return &SuccessResult{
			RPCName: name,
			ChainID: 1,
		}, nil
	}

	results, err := checkRPCs(rpcs, 5*time.Second, fakeChecker)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Success == nil {
		t.Fatal("rpc-1: expected success result")
	}

	if results[1].Failed == nil {
		t.Fatal("rpc-2: expected failed result")
	}

	if results[1].Failed.RPCName != "rpc-2" {
		t.Errorf(
			"expected failed RPC name %q, got %q",
			"rpc-2",
			results[1].Failed.RPCName,
		)
	}

	if results[1].Failed.Error == nil {
		t.Fatal("rpc-2: expected error")
	}

	if results[2].Success == nil {
		t.Fatal("rpc-3: expected success result")
	}
}

func TestCheck_MissingRPCAndConfig(t *testing.T) {
	got := check([]string{})

	if got != 1 {
		t.Errorf("check() = %d, want 1", got)
	}
}

func TestCheck_RPCAndConfigProvided(t *testing.T) {
	got := check([]string{
		"--rpc", "https://example.com",
		"--config", "chainpulse.yaml",
	})

	if got != 1 {
		t.Errorf("check() = %d, want 1", got)
	}
}

func TestCheck_InvalidTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{
			name:    "zero timeout",
			timeout: "0s",
		},
		{
			name:    "negative timeout",
			timeout: "-1s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := check([]string{
				"--rpc", "https://example.com",
				"--timeout", tt.timeout,
			})

			if got != 1 {
				t.Errorf("check() = %d, want 1", got)
			}
		})
	}
}

func TestCheck_ConfigFileNotFound(t *testing.T) {
	got := check([]string{
		"--config", "nonexistent-chainpulse-config.yaml",
	})

	if got != 1 {
		t.Errorf("check() = %d, want 1", got)
	}
}

func TestCheckRPCs_AllSuccess(t *testing.T) {
	rpcs := []config.RPCConfig{
		{
			Name: "rpc-1",
			URL:  "https://rpc-1.example.com",
		},
		{
			Name: "rpc-2",
			URL:  "https://rpc-2.example.com",
		},
		{
			Name: "rpc-3",
			URL:  "https://rpc-3.example.com",
		},
	}

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		return &SuccessResult{
			RPCName: name,
			ChainID: 1,
		}, nil
	}

	results, err := checkRPCs(rpcs, 5*time.Second, fakeChecker)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Success == nil {
		t.Fatal("rpc-1: expected success result")
	}

	if results[1].Success == nil {
		t.Fatal("rpc-2: expected success result")
	}

	if results[2].Success == nil {
		t.Fatal("rpc-3: expected success result")
	}

	for i := 0; i < len(rpcs); i++ {
		if rpcs[i].Name != results[i].Success.RPCName {
			t.Errorf(
				"expected name %q, got %q",
				rpcs[i].Name,
				results[i].Success.RPCName,
			)
		}
	}
}
func TestCheckRPCs_PassesRPCDetails(t *testing.T) {
	rpcs := []config.RPCConfig{
		{
			Name: "rpc-1",
			URL:  "https://rpc-1.example.com",
		},
	}

	expectedTimeout := 5 * time.Second

	var receivedURL string
	var receivedTimeout time.Duration
	var receivedName string

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		receivedURL = rpcURL
		receivedTimeout = timeout
		receivedName = name

		return &SuccessResult{
			RPCName: name,
			ChainID: 1,
		}, nil
	}

	_, err := checkRPCs(rpcs, expectedTimeout, fakeChecker)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if receivedURL != rpcs[0].URL {
		t.Errorf("expected URL %q, got %q", rpcs[0].URL, receivedURL)
	}

	if receivedTimeout != expectedTimeout {
		t.Errorf("expected timeout %v, got %v", expectedTimeout, receivedTimeout)
	}

	if receivedName != rpcs[0].Name {
		t.Errorf("expected name %q, got %q", rpcs[0].Name, receivedName)
	}
}

func TestCheckRPCs_Concurrent(t *testing.T) {
	rpcs := []config.RPCConfig{
		{Name: "rpc-1", URL: "https://rpc-1.example.com"},
		{Name: "rpc-2", URL: "https://rpc-2.example.com"},
		{Name: "rpc-3", URL: "https://rpc-3.example.com"},
	}

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		time.Sleep(200 * time.Millisecond)

		return &SuccessResult{
			RPCName: name,
			ChainID: 1,
		}, nil
	}

	start := time.Now()

	_, err := checkRPCs(rpcs, 5*time.Second, fakeChecker)

	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if elapsed >= 500*time.Millisecond {
		t.Errorf("expected concurrent execution under 500ms, took %v", elapsed)
	}
}

func TestCheckWithRPC_AllHealthy(t *testing.T) {
	configContent := `
timeout: 5s
block_lag_threshold: 3
rpcs:
  - name: eth-a
    url: https://eth-a.example.com
  - name: eth-b
    url: https://eth-b.example.com
`

	configPath := writeTestConfig(t, configContent)

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		blockNumber := uint64(100)
		if name == "eth-b" {
			blockNumber = 98
		}

		return &SuccessResult{
			RPCName:     name,
			ChainID:     1,
			BlockNumber: blockNumber,
		}, nil
	}

	got := checkWithRPC(
		[]string{"--config", configPath},
		fakeChecker,
	)

	if got != 0 {
		t.Errorf("checkWithRPC() = %d, want 0", got)
	}
}

func TestCheckWithRPC_LaggingRPC(t *testing.T) {
	configContent := `
timeout: 5s
block_lag_threshold: 3
rpcs:
  - name: eth-a
    url: https://eth-a.example.com
  - name: eth-b
    url: https://eth-b.example.com
`

	configPath := writeTestConfig(t, configContent)

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		blockNumber := uint64(100)
		if name == "eth-b" {
			blockNumber = 96
		}

		return &SuccessResult{
			RPCName:     name,
			ChainID:     1,
			BlockNumber: blockNumber,
		}, nil
	}

	got := checkWithRPC(
		[]string{"--config", configPath},
		fakeChecker,
	)

	if got != 1 {
		t.Errorf("checkWithRPC() = %d, want 1", got)
	}
}

func TestCheckWithRPC_FailedRPC(t *testing.T) {
	configContent := `
timeout: 5s
block_lag_threshold: 3
rpcs:
  - name: eth-a
    url: https://eth-a.example.com
  - name: eth-b
    url: https://eth-b.example.com
`

	configPath := writeTestConfig(t, configContent)

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		if name == "eth-b" {
			return nil, fmt.Errorf("connection failed")
		}

		return &SuccessResult{
			RPCName:     name,
			ChainID:     1,
			BlockNumber: 100,
		}, nil
	}

	got := checkWithRPC(
		[]string{"--config", configPath},
		fakeChecker,
	)

	if got != 1 {
		t.Errorf("checkWithRPC() = %d, want 1", got)
	}
}

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()

	configPath := t.TempDir() + "/chainpulse.yaml"

	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	return configPath
}

func TestCheckWithRPC_ConfigUsesConfiguredTimeout(t *testing.T) {
	configContent := `
timeout: 7s
block_lag_threshold: 3
rpcs:
  - name: eth-a
    url: https://eth-a.example.com
`

	configPath := writeTestConfig(t, configContent)

	var receivedTimeout time.Duration

	fakeChecker := func(
		rpcURL string,
		timeout time.Duration,
		name string,
	) (*SuccessResult, error) {
		receivedTimeout = timeout

		return &SuccessResult{
			RPCName:     name,
			ChainID:     1,
			BlockNumber: 100,
		}, nil
	}

	got := checkWithRPC(
		[]string{
			"--config", configPath,
			"--timeout", "1s",
		},
		fakeChecker,
	)

	if got != 0 {
		t.Fatalf("checkWithRPC() = %d, want 0", got)
	}

	if receivedTimeout != 7*time.Second {
		t.Errorf(
			"checker received timeout %v, want %v",
			receivedTimeout,
			7*time.Second,
		)
	}
}
