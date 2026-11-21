package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

// Config is the top-level structure in config.json.
type Config struct {
	Targets []Target `json:"targets"`
}

// Target describes one HTTP or TCP health check.
type Target struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Address        string `json:"address"`
	ExpectedStatus int    `json:"expected_status,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
	Contains       string `json:"contains,omitempty"`
	RetryCount     int    `json:"retry_count,omitempty"`
}

// ReadConfig loads and validates a JSON configuration file.
func ReadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if len(config.Targets) == 0 {
		return Config{}, fmt.Errorf("config must contain at least one target")
	}
	for i, target := range config.Targets {
		if err := target.Validate(); err != nil {
			return Config{}, fmt.Errorf("target %d: %w", i+1, err)
		}
	}
	return config, nil
}

// Validate rejects invalid settings before any network checks are started.
func (t Target) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if t.RetryCount < 0 || t.RetryCount > 3 {
		return fmt.Errorf("retry_count must be between 0 and 3")
	}

	switch strings.ToLower(t.Type) {
	case "http":
		parsed, err := url.ParseRequestURI(t.Address)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("http address must be a valid http or https URL")
		}
		if t.ExpectedStatus < 0 || t.ExpectedStatus > 599 {
			return fmt.Errorf("expected_status must be between 0 and 599")
		}
	case "tcp":
		if _, _, err := net.SplitHostPort(t.Address); err != nil {
			return fmt.Errorf("tcp address must be host:port: %w", err)
		}
	default:
		return fmt.Errorf("unsupported type %q (use http or tcp)", t.Type)
	}
	return nil
}
