package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Config struct {
	Token          string `json:"token"`
	DataDir        string `json:"data_dir"`
	CommandTimeout int    `json:"command_timeout_seconds"`
}

func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if cfg.Token == "" {
		return nil, errors.New("config: token is required")
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
	}
	if cfg.CommandTimeout <= 0 {
		cfg.CommandTimeout = 30
	}
	return &cfg, nil
}

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.CommandTimeout) * time.Second
}
