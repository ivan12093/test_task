package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	OAuth    OAuthConfig    `yaml:"oauth"`
}

type ServerConfig struct {
	Addr        string `yaml:"addr"`
	FrontendURL string `yaml:"frontend_url"`
	FullAddress string `yaml:"full_address"`
	CORSEnabled bool   `yaml:"cors_enabled"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type OAuthConfig struct {
	Google GoogleOAuthConfig `yaml:"google"`
}

type GoogleOAuthConfig struct {
	RedirectURL  string `yaml:"redirect_url"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

func Load(configPath string, envPath string) (*Config, error) {
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load .env file: %w", err)
			}
		}
	} else {
		_ = godotenv.Load()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	applyEnvOverrides(&config)

	return &config, nil
}

func applyEnvOverrides(config *Config) {
	if val := getEnvFirst("PORT"); val != "" {
		config.Server.Addr = ":" + val
	} else if val := getEnvFirst("SERVER_ADDR"); val != "" {
		config.Server.Addr = val
	}

	if val := getEnvFirst("FRONTEND_URL", "SERVER_FRONTEND_URL"); val != "" {
		config.Server.FrontendURL = val
	}

	if val := getEnvFirst("SERVER_URL", "SERVER_FULL_ADDRESS"); val != "" {
		config.Server.FullAddress = val
	}

	if val := getEnvFirst("MYSQLHOST", "DATABASE_HOST"); val != "" {
		config.Database.Host = val
	}

	if port := getEnvInt("MYSQLPORT", "DATABASE_PORT"); port > 0 {
		config.Database.Port = port
	}

	if val := getEnvFirst("MYSQLDATABASE", "DATABASE_DATABASE"); val != "" {
		config.Database.Database = val
	}

	if val := getEnvFirst("MYSQLUSER", "DATABASE_USERNAME"); val != "" {
		config.Database.Username = val
	}

	if val := getEnvFirst("MYSQLPASSWORD", "DATABASE_PASSWORD"); val != "" {
		config.Database.Password = val
	}

	if val := getEnvFirst("GOOGLE_CLIENT_ID", "OAUTH_GOOGLE_CLIENT_ID"); val != "" {
		config.OAuth.Google.ClientID = val
	}

	if val := getEnvFirst("GOOGLE_CLIENT_SECRET", "OAUTH_GOOGLE_CLIENT_SECRET"); val != "" {
		config.OAuth.Google.ClientSecret = val
	}

	if val := getEnvFirst("OAUTH_REDIRECT_URL", "OAUTH_GOOGLE_REDIRECT_URL"); val != "" {
		config.OAuth.Google.RedirectURL = val
	}

	if val := getEnvFirst("CORS_ENABLED", "ENABLE_CORS"); val != "" {
		config.Server.CORSEnabled = val == "true" || val == "1" || val == "yes"
	}
}

func getEnvFirst(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return ""
}

func getEnvInt(keys ...string) int {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			var port int
			if _, err := fmt.Sscanf(val, "%d", &port); err == nil {
				return port
			}
		}
	}
	return 0
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}
