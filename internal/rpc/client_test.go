package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("expected application/json, got %s", r.Header.Get("Content-Type"))
			}
			w.Header().Set("Content-Type", "application/json")

			w.Write([]byte(body))
		},
	))

	return server
}
func TestChainID(t *testing.T) {
	server := newTestServer(t, []byte(`{
                "jsonrpc": "2.0",
                "id": 1,
                "result": "0x2105"
            }`))
	defer server.Close()
	client := NewClient(server.URL)

	chainID, latency, err := client.ChainID(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if latency <= 0 {
		t.Fatalf("expected latency  > 0, got %v", latency)
	}

	if chainID != 8453 {
		t.Fatalf("expected chain ID 8453, got %d", chainID)
	}

}

func TestChainID_RPCError(t *testing.T) {
	server := newTestServer(t, []byte(`{
				"jsonrpc": "2.0",
				"id": 1,
				"error": {
					"code": -32601,
					"message": "method not found"
				}
			}`))
	defer server.Close()
	client := NewClient(server.URL)

	_, _, err := client.ChainID(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "method not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChainID_InvalidResult(t *testing.T) {
	server := newTestServer(t, []byte(`{
		"jsonrpc": "2.0",
		"id": 1,
		"result": "not-a-number"
	}`))
	defer server.Close()
	client := NewClient(server.URL)

	_, _, err := client.ChainID(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChainID_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	))
	defer server.Close()
	client := NewClient(server.URL)

	_, _, err := client.ChainID(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected HTTP 500 error, got %v", err)
	}

}

func TestChainID_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-time.After(100 * time.Millisecond):
				// deliberately don't respond
			case <-r.Context().Done():
				return
			}
		},
	))
	defer server.Close()

	// The timeout applies to the entire health check,
	// not to each individual RPC request.

	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()
	client := NewClient(server.URL)
	_, latency, err := client.ChainID(ctx)

	if latency <= 0 {
		t.Fatalf("expected latency > 0, got %v", latency)
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
