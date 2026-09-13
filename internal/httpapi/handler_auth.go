package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	"github.com/gin-gonic/gin"
)

// ---- auth DTOs ----

type loginRequestReq struct {
	Username string `json:"username" binding:"required"`
}

type loginVerifyReq struct {
	Code string `json:"code" binding:"required,len=6"`
}

type loginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      userResponse `json:"user"`
}

type userResponse struct {
	ID           string `json:"id"`
	TelegramID   int64  `json:"telegram_id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Timezone     string `json:"timezone"`
	DailyReport  bool   `json:"daily_report"`
	PendingTasks int64  `json:"pending_tasks"`
}

// ---- public auth routes ----

func (h *Handlers) registerAuth(public *gin.RouterGroup) {
	g := public.Group("/auth")
	{
		g.POST("/login/request", h.requestLogin)
		g.POST("/login/verify", h.verifyLogin)
	}
}

// requestLogin is the web initiated flow: user gives a username, the bot dms the code.
func (h *Handlers) requestLogin(c *gin.Context) {
	var req loginRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "username الزامی است")
		return
	}

	username := strings.TrimPrefix(strings.TrimSpace(req.Username), "@")
	user, code, err := h.auth.RequestLoginByUsername(c.Request.Context(), username)
	if err != nil {
		requestLoggerOf(c).Warn("login request failed", zapErr(err), zapStr("username", username))
		fail(c, http.StatusNotFound, "کاربر پیدا نشد؛ اول در تلگرام با بات استارت کن")
		return
	}

	// deliver the code to the user through the bot.
	if err := h.telegram.SendLoginCode(c.Request.Context(), user.TelegramID, code.Code, 10); err != nil {
		requestLoggerOf(c).Error("send login code failed", zapErr(err))
		fail(c, http.StatusInternalServerError, "ارسال کد توسط بات ناموفق بود")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "کد ورود توسط بات در تلگرام برای تو ارسال شد",
		"expires_in": int(service.CodeTTL.Seconds()),
	})
}

// verifyLogin checks the code and issues a session.
func (h *Handlers) verifyLogin(c *gin.Context) {
	var req loginVerifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "کد ۶ رقمی الزامی است")
		return
	}

	user, session, err := h.auth.VerifyLoginCode(c.Request.Context(), req.Code, c.Request.UserAgent())
	if err != nil {
		requestLoggerOf(c).Warn("login verify failed", zapErr(err))
		fail(c, http.StatusUnauthorized, "کد نامعتبر یا منقضی است")
		return
	}

	u, _, err := h.profiles.Profile(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	c.JSON(http.StatusOK, loginResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt,
		User:      toUserResponse(u, 0),
	})
}

// ---- authed routes ----

func (h *Handlers) registerAuthed(authed *gin.RouterGroup) {
	g := authed.Group("/auth")
	{
		g.GET("/me", h.me)
		g.POST("/logout", h.logout)
	}
}

// me returns the logged in user.
func (h *Handlers) me(c *gin.Context) {
	user := currentUser(c)

	u, st, err := h.profiles.Profile(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	c.JSON(http.StatusOK, toUserResponse(u, st.Pending))
}

// logout revokes the token.
func (h *Handlers) logout(c *gin.Context) {
	token := bearerToken(c)
	if err := h.auth.Logout(c.Request.Context(), token); err != nil {
		fail(c, http.StatusInternalServerError, "خطا در خروج")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "خروج انجام شد"})
}

// toUserResponse converts the domain model into the user DTO.
func toUserResponse(u *domain.User, pending int64) userResponse {
	return userResponse{
		ID:           u.ID.String(),
		TelegramID:   u.TelegramID,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Timezone:     u.Timezone,
		DailyReport:  u.DailyReport,
		PendingTasks: pending,
	}
}
