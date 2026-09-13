// Package app اجزای برنامه را راه‌اندازی و به هم وصل می‌کند.
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

// App نگهدارنده‌ی اجزای اجرایی برنامه است.
type App struct {
	Config config.Config
	Logger *zap.Logger
	Router *gin.Engine
	Bot    *bot.Bot
	cron   *cron.Cron
	stop   context.CancelFunc
}

// New تنظیمات، لاگر، دیتابیس، بات، زمان‌بند گزارش روزانه و یادآوری‌ها را راه می‌اندازد.
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
	// سرویس‌های مشترک بات و پنل وب.
	authSvc := service.NewAuthService(db, log)
	profileSvc := service.NewUserService(db, log)

	telegram, err := bot.New(c.TelegramBotToken, c.OwnerUsername, db, log, authSvc, profileSvc)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	// بات تا لغو شدن کانتکست آپدیت‌ها را دریافت می‌کند.
	ctx, cancel := context.WithCancel(context.Background())
	go telegram.Run(ctx)

	// زمان‌بند: گزارش روزانه و یادآوری موعد تسک‌ها.
	scheduler := cron.New(cron.WithSeconds())
	if _, err = scheduler.AddFunc(c.DailyReportCron, func() {
		telegram.SendDailyReports(ctx)
	}); err != nil {
		cancel()
		return nil, fmt.Errorf("invalid DAILY_REPORT_CRON: %w", err)
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
		Router: httpapi.NewRouter(c, log, telegram, authSvc, profileSvc, telegram.Tasks()),
		Bot:    telegram,
		cron:   scheduler,
		stop:   cancel,
	}, nil
}

// Stop زمان‌بند و حلقه‌ی دریافت آپدیت‌ها را متوقف می‌کند.
func (a *App) Stop() {
	a.stop()
	a.cron.Stop()
}
