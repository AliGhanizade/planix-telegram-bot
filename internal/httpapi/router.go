// Package httpapi روتر HTTP کوچک برنامه را می‌سازد.
package httpapi

import (
	"net/http"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// NewRouter مسیرهای سلامت سرویس، قرارداد OpenAPI و وب‌هوک تلگرام را ثبت می‌کند.
func NewRouter(cfg config.Config, telegram *bot.Bot, log *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("./docs/openapi.yaml")
	})

	r.POST("/telegram/webhook", func(c *gin.Context) {
		if cfg.TelegramWebhookSecret != "" {
			if c.GetHeader("X-Telegram-Bot-Api-Secret-Token") != cfg.TelegramWebhookSecret {
				log.Warn("invalid telegram webhook secret")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		var update tgbotapi.Update
		if err := c.ShouldBindJSON(&update); err != nil {
			log.Error("invalid telegram update", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := telegram.HandleUpdate(c.Request.Context(), update); err != nil {
			log.Error("failed to handle update", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.Status(http.StatusOK)
	})

	return r
}

// requestLogger هر درخواست HTTP را با ساختار zap لاگ می‌کند.
func requestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log.Info(
			"http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}
