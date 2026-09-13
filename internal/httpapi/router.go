// Package httpapi روتر HTTP کوچک برنامه را می‌سازد.
package httpapi

import (
	"net/http"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

// NewRouter مسیرهای سلامت سرویس، قرارداد OpenAPI و وب‌هوک تلگرام را ثبت می‌کند.
func NewRouter(cfg config.Config, telegram *bot.Bot, log *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(CORS(cfg.WebCORSOrigin))
	r.Use(RequestLogger(log))

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

		var update models.Update
		if err := c.ShouldBindJSON(&update); err != nil {
			log.Error("invalid telegram update", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// پردازش آپدیت ناهمگام انجام می‌شود؛ بلافاصله 200 برمی‌گردد.
		telegram.ProcessUpdate(c.Request.Context(), &update)
		c.Status(http.StatusOK)
	})

	return r
}
