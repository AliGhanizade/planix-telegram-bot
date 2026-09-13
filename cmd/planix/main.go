// Planix Telegram Bot — بات تلگرامی فارسی برای برنامه‌ریزی و پیگیری تسک‌ها.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// برای خواندن تایم‌زون‌ها (مثل Asia/Tehran) در محیط‌های بدون tzdata سیستمی.
	_ "time/tzdata"

	"github.com/AliGhanizade/planix-telegram-bot/internal/app"
	"go.uber.org/zap"
)

func main() {
	application, err := app.New()
	if err != nil {
		panic(err)
	}
	defer application.Logger.Sync()

	server := &http.Server{Addr: application.Config.HTTPAddr, Handler: application.Router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		application.Logger.Info("http server started", zap.String("address", application.Config.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			application.Logger.Fatal("http server failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	application.Stop()
	if err := server.Shutdown(ctx); err != nil {
		application.Logger.Error("graceful shutdown failed", zap.Error(err))
	}
}
