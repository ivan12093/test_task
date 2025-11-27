package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	API    APIConfig    `yaml:"api"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type APIConfig struct {
	BaseURL string `yaml:"base_url"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if apiBaseURL := os.Getenv("API_BASE_URL"); apiBaseURL != "" {
		config.API.BaseURL = apiBaseURL
	}

	return &config, nil
}
