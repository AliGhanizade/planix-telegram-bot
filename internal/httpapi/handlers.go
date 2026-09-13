// Package httpapi رابط HTTP برنامه (وب‌هوک تلگرام + API پنل وب) را می‌سازد.
package httpapi

import (
	"net/http"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

// Handlers نگهدارنده‌ی وابستگی‌های مشترک هندلرهای HTTP است.
type Handlers struct {
	cfg      config.Config
	log      *zap.Logger
	telegram *bot.Bot
	auth     *service.AuthService
	profiles *service.UserService
	tasks    *service.TaskService
}

// NewHandlers هندلرها را با وابستگی‌های تزریق‌شده می‌سازد.
func NewHandlers(cfg config.Config, log *zap.Logger, telegram *bot.Bot, auth *service.AuthService, profiles *service.UserService, tasks *service.TaskService) *Handlers {
	return &Handlers{cfg: cfg, log: log, telegram: telegram, auth: auth, profiles: profiles, tasks: tasks}
}

// register مسیرهای عمومی را ثبت می‌کند.
func (h *Handlers) register(r *gin.Engine) {
	r.GET("/healthz", h.health)
	r.GET("/openapi.yaml", h.openapi)
	r.POST("/telegram/webhook", h.telegramWebhook)
}

// health بررسی سلامت سرویس.
func (h *Handlers) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// openapi سرو قرارداد OpenAPI.
func (h *Handlers) openapi(c *gin.Context) {
	c.File("./docs/openapi.yaml")
}

// telegramWebhook دریافت آپدیت تلگرام با اعتبارسنجی secret.
func (h *Handlers) telegramWebhook(c *gin.Context) {
	if h.cfg.TelegramWebhookSecret != "" {
		if c.GetHeader("X-Telegram-Bot-Api-Secret-Token") != h.cfg.TelegramWebhookSecret {
			h.log.Warn("invalid telegram webhook secret")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}

	var update models.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		requestLoggerOf(c).Error("invalid telegram update", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// پردازش آپدیت ناهمگام انجام می‌شود؛ بلافاصله 200 برمی‌گردد.
	h.telegram.ProcessUpdate(c.Request.Context(), &update)
	c.Status(http.StatusOK)
}
