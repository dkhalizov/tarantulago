package config

import (
	"fmt"
	"os"
)

type Config struct {
	TelegramToken string
	PostgresURL   string
	LogLevel      string
}

func LoadConfig() (*Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN environment variable is not set")
	}

	dbPath := os.Getenv("POSTGRES_URL")
	if dbPath == "" {
		return nil, fmt.Errorf("POSTGRES_URL environment variable is not set")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		TelegramToken: token,
		PostgresURL:   dbPath,
		LogLevel:      logLevel,
	}, nil
}
