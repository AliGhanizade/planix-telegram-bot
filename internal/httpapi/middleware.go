package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// context keys for values stored on the request.
const (
	ctxRequestID = "planix.request_id"
	ctxLogger    = "planix.logger"
	ctxUser      = "planix.user"
)

// RequestID generates a unique id per request, echoes it in the response header
// and stores it in the context so all logs of the request are traceable.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ctxRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// RequestLogger logs every request with its own correlated logger.
func RequestLogger(root *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqLog := root.With(
			zap.String("request_id", requestIDOf(c)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
		c.Set(ctxLogger, reqLog)

		c.Next()

		reqLog.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

// CORS sets cross origin headers for the web panel.
func CORS(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// requestIDOf reads the request id from the context.
func requestIDOf(c *gin.Context) string {
	if id, ok := c.Get(ctxRequestID); ok {
		return id.(string)
	}
	return ""
}

// requestLoggerOf reads the per request logger from the context.
func requestLoggerOf(c *gin.Context) *zap.Logger {
	if l, ok := c.Get(ctxLogger); ok {
		return l.(*zap.Logger)
	}
	return zap.NewNop()
}

// Auth validates the bearer token and stores the user in the context.
func Auth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "توکن ارسال نشده است"})
			return
		}
		user, err := auth.ValidateSession(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "نشست نامعتبر یا منقضی است"})
			return
		}
		c.Set(ctxUser, user)
		c.Next()
	}
}

// RateLimit is a simple per ip request limiter for public auth routes.
func RateLimit(perMinute int) gin.HandlerFunc {
	type bucket struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(time.Minute)}
			buckets[ip] = b
		}
		b.count++
		allowed := b.count <= perMinute
		mu.Unlock()

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "درخواست‌های زیاد؛ کمی بعد تلاش کن"})
			return
		}
		c.Next()
	}
}
