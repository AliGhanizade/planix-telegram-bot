// Package config تنظیمات اجرایی بات را از متغیرهای محیطی می‌خواند.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config مجموعه‌ی تنظیمات لازم برای اجرای پلنیکس است.
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

// Load تنظیمات را می‌خواند و مقادیر اجباری را بررسی می‌کند.
// توکن بات هرگز مقدار پیش‌فرض ندارد و همیشه باید از متغیر محیطی خوانده شود.
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

// env مقدار متغیر محیطی را برمی‌گرداند یا در نبودش مقدار پیش‌فرض را.
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// envInt مقدار عددی متغیر محیطی را برمی‌گرداند یا در نبود/نامعتبری آن مقدار پیش‌فرض را.
func envInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
