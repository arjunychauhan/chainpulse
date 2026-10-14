package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {

	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "test_config.yaml")

	yamlData := []byte(`
timeout: 5s
block_lag_threshold: 3
rpcs:
  - name: New rpc
    url: https://example.com
`)
	err := os.WriteFile(testFilePath, yamlData, 0644)

	if err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	cfg, err := Load(testFilePath)

	if err != nil {
		t.Fatalf("failed to load yaml: %v", err)
	}

	if len(cfg.RPCs) != 1 {
		t.Fatalf("Expected 1 RPC, got %d", len(cfg.RPCs))
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Expected Timeout %q, got %q", 5*time.Second, cfg.Timeout)
	}

	if cfg.RPCs[0].Name != "New rpc" {
		t.Errorf("Expected Name %q, got %q", "New rpc", cfg.RPCs[0].Name)
	}

	if cfg.RPCs[0].URL != "https://example.com" {
		t.Errorf("Expected URL %q, got %q", "https://example.com", cfg.RPCs[0].URL)
	}
	if cfg.BlockLagThreshold != 3 {
		t.Errorf(
			"expected BlockLagThreshold %d, got %d",
			uint64(3),
			cfg.BlockLagThreshold,
		)
	}

}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("does-not-exist.yaml")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoad_ExplicitZero(t *testing.T) {
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "test_config.yaml")

	yamlData := []byte(`
timeout: 5s
block_lag_threshold: 0
rpcs:
  - name: New rpc
    url: https://example.com
`)
	err := os.WriteFile(testFilePath, yamlData, 0644)

	if err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	cfg, err := Load(testFilePath)

	if err != nil {
		t.Fatalf("failed to load yaml: %v", err)
	}

	if cfg.BlockLagThreshold != 0 {
		t.Errorf(
			"expected BlockLagThreshold %d, got %d",
			uint64(0),
			cfg.BlockLagThreshold,
		)
	}
}

func TestLoad_DefaultBlockLagThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "test_config.yaml")

	yamlData := []byte(`
timeout: 5s
rpcs:
  - name: ethereum
    url: https://example.com
`)

	if err := os.WriteFile(testFilePath, yamlData, 0644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	cfg, err := Load(testFilePath)
	if err != nil {
		t.Fatalf("failed to load yaml: %v", err)
	}

	if cfg.BlockLagThreshold != 3 {
		t.Errorf(
			"expected default threshold 3, got %d",
			cfg.BlockLagThreshold,
		)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "invalid.yaml")

	yamlData := []byte(`
timeout: [invalid
rpcs:
  - name: test
`)

	err := os.WriteFile(testFilePath, yamlData, 0644)

	if err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	_, err = Load(testFilePath)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

}
