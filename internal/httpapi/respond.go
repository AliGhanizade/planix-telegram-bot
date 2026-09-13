package httpapi

import (
	"strings"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// fail writes the standard error body {error: "..."}.
func fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// bearerToken extracts the token from the Authorization header.
func bearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if token, ok := strings.CutPrefix(header, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return ""
}

// currentUser reads the user stored by the Auth middleware.
func currentUser(c *gin.Context) *domain.User {
	if v, ok := c.Get(ctxUser); ok {
		return v.(*domain.User)
	}
	return nil
}

// zapErr builds an error log field.
func zapErr(err error) zap.Field { return zap.Error(err) }

// zapStr builds a string log field.
func zapStr(key, value string) zap.Field { return zap.String(key, value) }
