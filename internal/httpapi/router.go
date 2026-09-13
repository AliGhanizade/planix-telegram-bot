// Package httpapi رابط HTTP برنامه (وب‌هوک تلگرام + API پنل وب) را می‌سازد.
package httpapi

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/bot"
	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRouter روتر Gin را با میدل‌ورها و مسیرها می‌سازد.
func NewRouter(
	cfg config.Config,
	log *zap.Logger,
	telegram *bot.Bot,
	auth *service.AuthService,
	profiles *service.UserService,
	tasks *service.TaskService,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(CORS(cfg.WebCORSOrigin))
	r.Use(RequestLogger(log))

	h := NewHandlers(cfg, log, telegram, auth, profiles, tasks)
	h.register(r)

	// API پنل وب: مسیرهای عمومی با محدودیت نرخ + مسیرهای احراز هویت‌شده.
	public := r.Group("/api", RateLimit(30))
	authed := r.Group("/api", Auth(auth))
	h.registerAuth(public)
	h.registerAuthed(authed)
	h.registerTasks(authed)
	h.registerProfile(authed)

	return r
}
