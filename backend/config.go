package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey  string `json:"api_key"`
	APIBase string `json:"api_base"`
	Model   string `json:"model"`
}

func LoadConfig() Config {
	cfg := Config{}

	// 1. global config: ~/.deltascope/config.json
	if home, err := os.UserHomeDir(); err == nil {
		globalPath := filepath.Join(home, ".deltascope", "config.json")
		if data, err := os.ReadFile(globalPath); err == nil {
			_ = json.Unmarshal(data, &cfg)
		}
	}

	// 2. local config: ./.deltascope.json
	if data, err := os.ReadFile(".deltascope.json"); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// 3. environment variables override config files
	if v := os.Getenv("DELTASCOPE_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("DELTASCOPE_API_BASE"); v != "" {
		cfg.APIBase = v
	}
	if v := os.Getenv("DELTASCOPE_MODEL"); v != "" {
		cfg.Model = v
	}

	return cfg
}

func SaveConfig(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".deltascope")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
