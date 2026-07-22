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
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type App struct {
	Config config.Config
	Logger *zap.Logger
	Router *gin.Engine
	cron   *cron.Cron
}

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
	telegram, err := bot.New(c.TelegramBotToken, db, log)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	// Start Telegram Long Polling
	go telegram.StartPolling(context.Background())
	scheduler := cron.New(cron.WithSeconds())
	if _, err = scheduler.AddFunc(c.DailyReportCron, func() {
		log.Info("daily report schedule started; report delivery will be added with the monitoring workflow")
	}); err != nil {
		return nil, fmt.Errorf("invalid DAILY_REPORT_CRON: %w", err)
	}
	scheduler.Start()
	return &App{Config: c, Logger: log, Router: httpapi.NewRouter(c, telegram, log), cron: scheduler}, nil
}
func (a *App) Stop() { a.cron.Stop() }
