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

// context keys برای مقادیر درج‌شده در درخواست.
const (
	ctxRequestID = "planix.request_id"
	ctxLogger    = "planix.logger"
	ctxUser      = "planix.user"
)

// RequestID برای هر درخواست شناسه یکتا تولید و در هدر پاسخ برمی‌گرداند
// و در کانتکست قرار می‌دهد تا همه‌ی لاگ‌های همان درخواست قابل ردیابی باشند.
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

// RequestLogger پایان هر درخواست را با لاگر اختصاصی همان درخواست لاگ می‌کند.
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

// CORS هدرهای اشتراک‌گذاری منابع را برای پنل وب تنظیم می‌کند.
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

// requestIDOf شناسه‌ی درخواست را از کانتکست می‌خواند.
func requestIDOf(c *gin.Context) string {
	if id, ok := c.Get(ctxRequestID); ok {
		return id.(string)
	}
	return ""
}

// requestLoggerOf لاگر اختصاصی درخواست را از کانتکست می‌خواند.
func requestLoggerOf(c *gin.Context) *zap.Logger {
	if l, ok := c.Get(ctxLogger); ok {
		return l.(*zap.Logger)
	}
	return zap.NewNop()
}

// Auth میدل‌ور نشست وب: توکن Bearer را بررسی و کاربر را در کانتکست می‌گذارد.
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

// RateLimit محدودکننده‌ی ساده‌ی نرخ درخواست بر اساس IP (برای مسیرهای عمومی احراز هویت).
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
