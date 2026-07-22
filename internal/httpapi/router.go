package httpapi

import (
	"net/http"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(cfg config.Config, telegram *bot.Bot, log *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

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
