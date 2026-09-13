// Package app اجزای برنامه را راه‌اندازی و به هم وصل می‌کند.
package app

import (
	"context"
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/AliGhanizade/planix-telegram-bot/internal/httpapi"
	"github.com/AliGhanizade/planix-telegram-bot/internal/platform/database"
	"github.com/AliGhanizade/planix-telegram-bot/internal/platform/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// App نگهدارنده‌ی اجزای اجرایی برنامه است.
type App struct {
	Config config.Config
	Logger *zap.Logger
	Router *gin.Engine
	Bot    *bot.Bot
	stop   context.CancelFunc
}

// New تنظیمات، لاگر، دیتابیس و بات را راه می‌اندازد.
func New() (*App, error) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}
	log, err := logger.New(c.LogLevel)
	if err != nil {
		return nil, err
	}
	db, err := database.Open(c.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	telegram, err := bot.New(c.TelegramBotToken, c.OwnerUsername, db, log)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	// بات تا لغو شدن کانتکست آپدیت‌ها را دریافت می‌کند.
	ctx, cancel := context.WithCancel(context.Background())
	go telegram.Run(ctx)

	return &App{
		Config: c,
		Logger: log,
		Router: httpapi.NewRouter(c, telegram, log),
		Bot:    telegram,
		stop:   cancel,
	}, nil
}

// Stop حلقه‌ی دریافت آپدیت‌ها را متوقف می‌کند.
func (a *App) Stop() { a.stop() }
