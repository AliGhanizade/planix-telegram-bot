// Package httpapi builds the http surface (telegram webhook + web panel api).
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

// Handlers holds shared dependencies for the http handlers.
type Handlers struct {
	cfg      config.Config
	log      *zap.Logger
	telegram *bot.Bot
	auth     *service.AuthService
	profiles *service.UserService
	tasks    *service.TaskService
	folders  *service.FolderService
}

// NewHandlers builds the handlers with injected dependencies.
func NewHandlers(cfg config.Config, log *zap.Logger, telegram *bot.Bot, auth *service.AuthService, profiles *service.UserService, tasks *service.TaskService, folders *service.FolderService) *Handlers {
	return &Handlers{cfg: cfg, log: log, telegram: telegram, auth: auth, profiles: profiles, tasks: tasks, folders: folders}
}

// register registers the public routes.
func (h *Handlers) register(r *gin.Engine) {
	r.GET("/healthz", h.health)
	r.POST("/telegram/webhook", h.telegramWebhook)
}

// health is the service health check.
func (h *Handlers) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// telegramWebhook receives telegram updates with secret validation.
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

	// the update is processed asynchronously; respond 200 right away.
	h.telegram.ProcessUpdate(c.Request.Context(), &update)
	c.Status(http.StatusOK)
}
