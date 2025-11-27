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

	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		config.OAuth.Google.ClientID = clientID
	}
	if clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET"); clientSecret != "" {
		config.OAuth.Google.ClientSecret = clientSecret
	}
	if dbPassword := os.Getenv("MYSQLPASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}
	if dbHost := os.Getenv("MYSQLHOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbPort := os.Getenv("MYSQLPORT"); dbPort != "" {
		var port int
		if _, err := fmt.Sscanf(dbPort, "%d", &port); err == nil {
			config.Database.Port = port
		}
	}
	if dbName := os.Getenv("MYSQLDATABASE"); dbName != "" {
		config.Database.Database = dbName
	}
	if dbUser := os.Getenv("MYSQLUSER"); dbUser != "" {
		config.Database.Username = dbUser
	}
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		config.Server.FrontendURL = frontendURL
	}

	return &config, nil
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
