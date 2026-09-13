// Package config reads bot settings from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the settings needed to run planix.
type Config struct {
	AppEnv                string
	HTTPAddr              string
	DatabaseURL           string
	TelegramBotToken      string
	TelegramWebhookSecret string
	OwnerUsername         string
	LogLevel              string
	DailyReportCron       string
	ReminderCron          string
	ReminderLeadMinutes   int
	WebCORSOrigin         string
}

// Load reads the settings and checks required values.
// the bot token never has a default and always comes from the environment.
func Load() (Config, error) {
	c := Config{
		AppEnv:                env("APP_ENV", "development"),
		HTTPAddr:              env("HTTP_ADDR", ":8080"),
		DatabaseURL:           env("DATABASE_URL", "postgres://planix:planix@localhost:5432/planix?sslmode=disable"),
		TelegramBotToken:      os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecret: os.Getenv("TELEGRAM_WEBHOOK_SECRET"),
		OwnerUsername:         env("TELEGRAM_OWNER_USERNAME", "AliGhanizade"),
		LogLevel:              env("LOG_LEVEL", "info"),
		DailyReportCron:       env("DAILY_REPORT_CRON", "0 0 21 * * *"),
		ReminderCron:          env("REMINDER_CRON", "0 */15 * * * *"),
		ReminderLeadMinutes:   envInt("REMINDER_LEAD_MINUTES", 120),
		WebCORSOrigin:         env("WEB_CORS_ORIGIN", "*"),
	}
	if c.TelegramBotToken == "" {
		return c, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	return c, nil
}

// env returns an environment variable or a fallback.
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// envInt returns an int env var or the fallback when missing or invalid.
func envInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
