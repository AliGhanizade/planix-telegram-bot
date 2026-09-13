// Package app wires the application components together.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/AliGhanizade/planix-telegram-bot/internal/httpapi"
	"github.com/AliGhanizade/planix-telegram-bot/internal/platform/database"
	"github.com/AliGhanizade/planix-telegram-bot/internal/platform/logger"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// App holds the running parts of the application.
type App struct {
	Config config.Config
	Logger *zap.Logger
	Router *gin.Engine
	Bot    *bot.Bot
	cron   *cron.Cron
	stop   context.CancelFunc
}

// New sets up config, logger, database, bot and the cron scheduler.
func New() (*App, error) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}
	log, err := logger.New(c.LogLevel, c.AppEnv)
	if err != nil {
		return nil, err
	}
	db, err := database.Open(c.DatabaseURL, log)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	// services shared by the bot and the web panel.
	authSvc := service.NewAuthService(db, log)
	profileSvc := service.NewUserService(db, log)
	folderSvc := service.NewFolderService(db, log)

	telegram, err := bot.New(c.TelegramBotToken, c.OwnerUsername, db, log, authSvc, profileSvc, folderSvc)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	// the bot receives updates until the context is cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	go telegram.Run(ctx)

	// scheduler: due date reminders and per user daily report times.
	scheduler := cron.New(cron.WithSeconds())
	if _, err = scheduler.AddFunc("0 * * * * *", func() {
		telegram.SendDailyReportsAtMinute(ctx, bot.TehranHHMM())
	}); err != nil {
		cancel()
		return nil, fmt.Errorf("invalid report schedule: %w", err)
	}
	if _, err = scheduler.AddFunc(c.ReminderCron, func() {
		telegram.SendDueReminders(ctx, time.Duration(c.ReminderLeadMinutes)*time.Minute)
	}); err != nil {
		cancel()
		return nil, fmt.Errorf("invalid REMINDER_CRON: %w", err)
	}
	scheduler.Start()

	return &App{
		Config: c,
		Logger: log,
		Router: httpapi.NewRouter(c, log, telegram, authSvc, profileSvc, telegram.Tasks(), folderSvc),
		Bot:    telegram,
		cron:   scheduler,
		stop:   cancel,
	}, nil
}

// Stop stops the scheduler and the update loop.
func (a *App) Stop() {
	a.stop()
	a.cron.Stop()
}
