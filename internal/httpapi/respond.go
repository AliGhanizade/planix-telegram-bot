package httpapi

import (
	"strings"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// fail پاسخ خطای استاندارد {error: "..."}.
func fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// bearerToken توکن از هدر Authorization را برمی‌گرداند.
func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if token, ok := strings.CutPrefix(header, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return ""
}

// currentUser کاربر تاییدشده توسط میدل‌ور Auth را از کانتکست می‌خواند.
func currentUser(c *gin.Context) *domain.User {
	if v, ok := c.Get(ctxUser); ok {
		return v.(*domain.User)
	}
	return nil
}

// zapErr فیلد لاگ خطا.
func zapErr(err error) zap.Field { return zap.Error(err) }

// zapStr فیلد لاگ رشته‌ای.
func zapStr(key, value string) zap.Field { return zap.String(key, value) }
