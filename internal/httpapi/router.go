// Package httpapi builds the http surface (telegram webhook + web panel api).
package httpapi

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRouter builds the gin router with middlewares and routes.
func NewRouter(
	cfg config.Config,
	log *zap.Logger,
	telegram *bot.Bot,
	auth *service.AuthService,
	profiles *service.UserService,
	tasks *service.TaskService,
	folders *service.FolderService,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(CORS(cfg.WebCORSOrigin))
	r.Use(RequestLogger(log))

	h := NewHandlers(cfg, log, telegram, auth, profiles, tasks, folders)
	h.register(r)

	// web panel api: public routes with rate limiting plus authed routes.
	public := r.Group("/api", RateLimit(30))
	authed := r.Group("/api", Auth(auth))
	h.registerAuth(public)
	h.registerAuthed(authed)
	h.registerTasks(authed)
	h.registerProfile(authed)
	h.registerFolders(authed, public)

	return r
}
