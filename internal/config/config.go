package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Timeout           time.Duration `yaml:"timeout"`
	RPCs              []RPCConfig   `yaml:"rpcs"`
	BlockLagThreshold uint64        `yaml:"block_lag_threshold"`
}

type rawConfig struct {
	Timeout           time.Duration `yaml:"timeout"`
	RPCs              []RPCConfig   `yaml:"rpcs"`
	BlockLagThreshold *uint64       `yaml:"block_lag_threshold"`
}

type RPCConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}

	if len(c.RPCs) == 0 {
		return fmt.Errorf("at least one RPC must be configured")

	}

	seenNames := make(map[string]struct{}, len(c.RPCs))

	for i, rpc := range c.RPCs {
		name := strings.TrimSpace(rpc.Name)
		url := strings.TrimSpace(rpc.URL)

		if name == "" {
			return fmt.Errorf("RPC at index %d has an empty name", i)
		}

		if url == "" {
			return fmt.Errorf("RPC %q has an empty URL", name)
		}

		if _, exists := seenNames[name]; exists {
			return fmt.Errorf("duplicate RPC name %q", name)
		}

		seenNames[name] = struct{}{}
	}

	return nil
}

func Load(path string) (*Config, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw rawConfig

	if err := yaml.Unmarshal(fileBytes, &raw); err != nil {
		return nil, err
	}

	cfg := &Config{
		Timeout: raw.Timeout,
		RPCs:    raw.RPCs,
	}

	if raw.BlockLagThreshold == nil {
		cfg.BlockLagThreshold = 3
	} else {
		cfg.BlockLagThreshold = *raw.BlockLagThreshold
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
