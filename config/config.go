package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	envAPIKey  = "NEOCITIES_API_KEY"
	configDir  = ".config/neocities"
	configFile = "config.json"
)

// Config holds the application configuration.
type Config struct {
	APIKey  string `json:"api_key,omitempty"`
	SiteDir string `json:"site_dir,omitempty"`
}

// configPath returns the full path to the config file.
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load loads configuration from environment variable or config file.
// Environment variable takes precedence.
func Load() (*Config, error) {
	cfg := &Config{}

	// Check environment variable first
	if apiKey := os.Getenv(envAPIKey); apiKey != "" {
		cfg.APIKey = apiKey
		return cfg, nil
	}

	// Try to load from config file
	path, err := configPath()
	if err != nil {
		return cfg, nil // Return empty config, not an error
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Config file doesn't exist, return empty config
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to the config file.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GetAPIKey returns the API key from environment or config.
func GetAPIKey() string {
	if apiKey := os.Getenv(envAPIKey); apiKey != "" {
		return apiKey
	}

	cfg, err := Load()
	if err != nil {
		return ""
	}

	return cfg.APIKey
}

// SaveAPIKey saves an API key to the config file.
func SaveAPIKey(apiKey string) error {
	cfg, err := Load()
	if err != nil {
		cfg = &Config{}
	}

	cfg.APIKey = apiKey
	return Save(cfg)
}
