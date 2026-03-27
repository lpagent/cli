package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultBaseURL = "https://api.lpagent.io/open-api/v1"
	configDir      = ".lpagent"
	configFile     = "config.json"
)

type Config struct {
	APIKey       string `json:"api_key,omitempty"`
	BaseURL      string `json:"api_base_url,omitempty"`
	DefaultOwner string `json:"default_owner,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDir, configFile), nil
}

func Load() (*Config, error) {
	cfg := &Config{
		BaseURL:      DefaultBaseURL,
		OutputFormat: "json",
	}

	p, err := configPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.OutputFormat == "" {
		cfg.OutputFormat = "json"
	}

	// Env var overrides
	if v := os.Getenv("LPAGENT_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("LPAGENT_API_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("LPAGENT_DEFAULT_OWNER"); v != "" {
		cfg.DefaultOwner = v
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	p, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(p, data, 0600)
}

func (c *Config) ResolveOwner(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if v := os.Getenv("LPAGENT_DEFAULT_OWNER"); v != "" {
		return v, nil
	}
	if c.DefaultOwner != "" {
		return c.DefaultOwner, nil
	}
	return "", fmt.Errorf("owner is required. Use --owner flag, LPAGENT_DEFAULT_OWNER env, or run: lpagent auth set-default-owner <address>")
}

func (c *Config) GetAPIKey(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if c.APIKey != "" {
		return c.APIKey, nil
	}
	return "", fmt.Errorf("no API key configured. Run: lpagent auth set-key")
}
