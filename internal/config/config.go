package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv           string
	HTTPAddr         string
	DatabaseURL      string
	TelegramBotToken string
	LogLevel         string
}

func Load() (Config, error) {
	c := Config{
		AppEnv:           env("APP_ENV", "development"),
		HTTPAddr:         env("HTTP_ADDR", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		LogLevel:         env("LOG_LEVEL", "info"),
	}
	if c.TelegramBotToken == "" {
		return c, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	return c, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
