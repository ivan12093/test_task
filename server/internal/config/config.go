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
	if val := os.Getenv("SERVER_ADDR"); val != "" {
		config.Server.Addr = val
	}
	if val := os.Getenv("PORT"); val != "" {
		config.Server.Addr = ":" + val
	}
	if val := os.Getenv("SERVER_FRONTEND_URL"); val != "" {
		config.Server.FrontendURL = val
	}
	if val := os.Getenv("FRONTEND_URL"); val != "" {
		config.Server.FrontendURL = val
	}
	if val := os.Getenv("SERVER_FULL_ADDRESS"); val != "" {
		config.Server.FullAddress = val
	}
	if val := os.Getenv("SERVER_URL"); val != "" {
		config.Server.FullAddress = val
		if config.OAuth.Google.RedirectURL == "" || config.OAuth.Google.RedirectURL == "http://localhost:8080/api/auth/google/callback" {
			config.OAuth.Google.RedirectURL = val + "/api/auth/google/callback"
		}
	}

	if val := os.Getenv("DATABASE_HOST"); val != "" {
		config.Database.Host = val
	}
	if val := os.Getenv("MYSQLHOST"); val != "" {
		config.Database.Host = val
	}
	if val := os.Getenv("DATABASE_PORT"); val != "" {
		var port int
		if _, err := fmt.Sscanf(val, "%d", &port); err == nil {
			config.Database.Port = port
		}
	}
	if val := os.Getenv("MYSQLPORT"); val != "" {
		var port int
		if _, err := fmt.Sscanf(val, "%d", &port); err == nil {
			config.Database.Port = port
		}
	}
	if val := os.Getenv("DATABASE_DATABASE"); val != "" {
		config.Database.Database = val
	}
	if val := os.Getenv("MYSQLDATABASE"); val != "" {
		config.Database.Database = val
	}
	if val := os.Getenv("DATABASE_USERNAME"); val != "" {
		config.Database.Username = val
	}
	if val := os.Getenv("MYSQLUSER"); val != "" {
		config.Database.Username = val
	}
	if val := os.Getenv("DATABASE_PASSWORD"); val != "" {
		config.Database.Password = val
	}
	if val := os.Getenv("MYSQLPASSWORD"); val != "" {
		config.Database.Password = val
	}

	if val := os.Getenv("OAUTH_GOOGLE_CLIENT_ID"); val != "" {
		config.OAuth.Google.ClientID = val
	}
	if val := os.Getenv("GOOGLE_CLIENT_ID"); val != "" {
		config.OAuth.Google.ClientID = val
	}
	if val := os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"); val != "" {
		config.OAuth.Google.ClientSecret = val
	}
	if val := os.Getenv("GOOGLE_CLIENT_SECRET"); val != "" {
		config.OAuth.Google.ClientSecret = val
	}
	if val := os.Getenv("OAUTH_GOOGLE_REDIRECT_URL"); val != "" {
		config.OAuth.Google.RedirectURL = val
	}
	if val := os.Getenv("OAUTH_REDIRECT_URL"); val != "" {
		config.OAuth.Google.RedirectURL = val
	}

	if val := os.Getenv("ENABLE_CORS"); val != "" {
		config.Server.CORSEnabled = val == "true" || val == "1" || val == "yes"
	}
	if val := os.Getenv("CORS_ENABLED"); val != "" {
		config.Server.CORSEnabled = val == "true" || val == "1" || val == "yes"
	}
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
