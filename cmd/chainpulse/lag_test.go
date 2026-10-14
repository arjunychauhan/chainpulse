package main

import (
	"fmt"
	"testing"
)

func TestHighestBlockByChain(t *testing.T) {
	results := []*RPCResult{
		{
			Success: &SuccessResult{
				RPCName:     "eth-a",
				ChainID:     1,
				BlockNumber: 1000,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-b",
				ChainID:     1,
				BlockNumber: 997,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "bsc-a",
				ChainID:     56,
				BlockNumber: 5000,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-c",
				ChainID:     1,
				BlockNumber: 1002,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "bsc-b",
				ChainID:     56,
				BlockNumber: 4998,
			},
		},
	}

	highestBlock := highestBlockByChain(results)

	if highestBlock == nil {
		t.Fatal("map is empty")
	}

	if len(highestBlock) != 2 {
		t.Fatalf("expected 2 chains, got %d", len(highestBlock))
	}

	if highestBlock[56] != 5000 {
		t.Errorf("expected block 5000, but got %v", highestBlock[56])
	}

	if highestBlock[1] != 1002 {
		t.Errorf("expected block 1002, but got %v", highestBlock[1])
	}

}
func TestHighestBlockByChain_IgnoresFailedRPC(t *testing.T) {
	results := []*RPCResult{
		{
			Failed: &FailedResult{
				RPCName: "eth-broken",
				Error:   fmt.Errorf("invalid rpc"),
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-good",
				ChainID:     1,
				BlockNumber: 1000,
			},
		},
	}

	highestBlock := highestBlockByChain(results)

	if len(highestBlock) != 1 {
		t.Fatal("map should be of 1 length")
	}
	if highestBlock[1] != 1000 {
		t.Errorf("expected block 1000, got %d", highestBlock[1])
	}

}

func TestHighestBlockByChain_EmptyInput(t *testing.T) {
	results := []*RPCResult{}

	highestBlock := highestBlockByChain(results)

	if len(highestBlock) != 0 {
		t.Fatal("map should be empty")
	}

}

func TestCalculateLag(t *testing.T) {
	results := []*RPCResult{
		{
			Success: &SuccessResult{
				RPCName:     "eth-a",
				ChainID:     1,
				BlockNumber: 1000,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-b",
				ChainID:     1,
				BlockNumber: 997,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "bsc-a",
				ChainID:     56,
				BlockNumber: 5000,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-c",
				ChainID:     1,
				BlockNumber: 1002,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "bsc-b",
				ChainID:     56,
				BlockNumber: 4998,
			},
		},
	}

	expected := map[string]uint64{
		"eth-a": 2,
		"eth-b": 5,
		"eth-c": 0,
		"bsc-a": 0,
		"bsc-b": 2,
	}

	highestBlock := highestBlockByChain(results)

	if highestBlock == nil {
		t.Fatal("map is empty")
	}

	lag := calculateLag(results, highestBlock)

	for _, result := range results {
		name := result.Success.RPCName
		got := lag[name]
		want := expected[name]

		if got != want {
			t.Errorf("RPC %s: expected lag %d, got %d", name, want, got)
		}
	}

}

func TestCalculateLag_IgnoresFailedRPC(t *testing.T) {

	results := []*RPCResult{
		{
			Failed: &FailedResult{
				RPCName: "eth-broken",
				Error:   fmt.Errorf("invalid rpc"),
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-good",
				ChainID:     1,
				BlockNumber: 1000,
			},
		},
	}

	highestBlock := highestBlockByChain(results)

	if highestBlock == nil {
		t.Fatal("map is empty")
	}

	lag := calculateLag(results, highestBlock)

	if lag["eth-good"] != 0 {
		t.Errorf("expected lag 0 for eth-good, got %d", lag["eth-good"])
	}

}

func TestCalculateLag_EmptyInput(t *testing.T) {
	results := []*RPCResult{}

	highestBlock := highestBlockByChain(results)
	if len(highestBlock) != 0 {
		t.Fatal("map should be empty")
	}

	lag := calculateLag(results, highestBlock)

	if len(lag) != 0 {
		t.Errorf("expected length is 0, but got %d", len(lag))
	}
}

func TestDetermineStatus(t *testing.T) {
	tests := []struct {
		name      string
		lag       uint64
		threshold uint64
		want      RPCStatus
	}{
		{"zero lag", 0, 3, StatusHealthy},
		{"below threshold", 2, 3, StatusHealthy},
		{"at threshold", 3, 3, StatusHealthy},
		{"above threshold", 4, 3, StatusLagging},
		{"far above threshold", 10, 3, StatusLagging},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineStatus(tt.lag, tt.threshold)

			if got != tt.want {
				t.Errorf(
					"determineStatus(%d, %d) = %s; want %s",
					tt.lag, tt.threshold, got, tt.want,
				)
			}
		})
	}
}

func TestAnalyzeLag(t *testing.T) {
	results := []*RPCResult{
		{
			Success: &SuccessResult{
				RPCName:     "eth-a",
				ChainID:     1,
				BlockNumber: 1002,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "eth-b",
				ChainID:     1,
				BlockNumber: 998,
			},
		},
		{
			Success: &SuccessResult{
				RPCName:     "bsc-a",
				ChainID:     56,
				BlockNumber: 5000,
			},
		},
		{
			Failed: &FailedResult{
				RPCName: "eth-broken",
				Error:   fmt.Errorf("invalid rpc"),
			},
		},
	}
	expectedLag := []uint64{0, 4, 0}
	expectedStatus := []RPCStatus{StatusHealthy, StatusLagging, StatusHealthy}
	expectedNames := []string{"eth-a", "eth-b", "bsc-a"}
	const threshold uint64 = 3

	lageResults := analyzeLag(results, threshold)

	if len(lageResults) != len(expectedLag) {
		t.Fatalf("expected length is %d, but got %d", len(expectedLag), len(lageResults))
	}

	for i, lageResult := range lageResults {
		if lageResult.Lag != expectedLag[i] {
			t.Errorf("Expected lag is %d, but got %d", expectedLag[i], lageResult.Lag)
		}
		if lageResult.Status != expectedStatus[i] {
			t.Errorf("Expected status is %s, but got %s", expectedStatus[i], lageResult.Status)
		}
		if lageResult.RPCName != expectedNames[i] {
			t.Errorf("expected RPC %s, got %s", expectedNames[i], lageResult.RPCName)
		}
	}

}
func TestHasUnhealthyRPCs(t *testing.T) {
	tests := []struct {
		name      string
		results   []*RPCResult
		threshold uint64
		want      bool
	}{
		{
			name:      "empty results",
			results:   []*RPCResult{},
			threshold: 3,
			want:      false,
		},
		{
			name: "all RPCs healthy",
			results: []*RPCResult{
				{
					Success: &SuccessResult{
						RPCName:     "eth-a",
						ChainID:     1,
						BlockNumber: 100,
					},
				},
				{
					Success: &SuccessResult{
						RPCName:     "eth-b",
						ChainID:     1,
						BlockNumber: 98,
					},
				},
			},
			threshold: 3,
			want:      false,
		},
		{
			name: "at least one RPC failed",
			results: []*RPCResult{
				{
					Success: &SuccessResult{
						RPCName:     "eth-a",
						ChainID:     1,
						BlockNumber: 100,
					},
				},
				{
					Failed: &FailedResult{
						RPCName: "eth-broken",
						Error:   fmt.Errorf("connection failed"),
					},
				},
			},
			threshold: 3,
			want:      true,
		},
		{
			name: "at least one RPC is lagging",
			results: []*RPCResult{
				{
					Success: &SuccessResult{
						RPCName:     "eth-a",
						ChainID:     1,
						BlockNumber: 100,
					},
				},
				{
					Success: &SuccessResult{
						RPCName:     "eth-b",
						ChainID:     1,
						BlockNumber: 96,
					},
				},
			},
			threshold: 3,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasUnhealthyRPCs(tt.results, tt.threshold)

			if got != tt.want {
				t.Errorf(
					"hasUnhealthyRPCs() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}
