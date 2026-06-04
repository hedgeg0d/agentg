package config

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Config struct {
	Token          string        `json:"token"`
	DataDir        string        `json:"data_dir"`
	UploadDir      string        `json:"upload_dir"`
	CommandTimeout int           `json:"command_timeout_seconds"`
	Access         Access        `json:"access"`
	Notifications  Notifications `json:"notifications"`
	Shell          Shell         `json:"shell"`
}

type Shell struct {
	StreamOutput *bool `json:"stream_output"`
}

func (s Shell) Streaming() bool {
	return s.StreamOutput == nil || *s.StreamOutput
}

type Notifications struct {
	DBusEnabled       bool `json:"dbus_enabled"`
	SystemBus         bool `json:"system_bus"`
	ReplaceNotifySend bool `json:"replace_notify_send"`
}

type Access struct {
	Admins           []int64  `json:"admins"`
	AllowedUsers     []int64  `json:"allowed_users"`
	AllowedUsernames []string `json:"allowed_usernames"`
	Password         string   `json:"password"`
	SessionTTLMin    int      `json:"session_ttl_minutes"`
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
	if cfg.UploadDir == "" {
		cfg.UploadDir = "."
	}
	if cfg.CommandTimeout <= 0 {
		cfg.CommandTimeout = 30
	}
	return &cfg, nil
}

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.CommandTimeout) * time.Second
}

func (a Access) SessionTTL() time.Duration {
	return time.Duration(a.SessionTTLMin) * time.Minute
}
