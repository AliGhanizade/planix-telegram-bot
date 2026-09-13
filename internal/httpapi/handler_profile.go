package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ---- profile DTOs ----

type profileResponse struct {
	ID          string  `json:"id"`
	TelegramID  int64   `json:"telegram_id"`
	Username    string  `json:"username"`
	FirstName   string  `json:"first_name"`
	LastName    string  `json:"last_name"`
	Timezone    string  `json:"timezone"`
	DailyReport bool    `json:"daily_report"`
	Pending     int64   `json:"pending_tasks"`
	Completed   int64   `json:"completed_tasks"`
	Cancelled   int64   `json:"cancelled_tasks"`
	ProgressPct float64 `json:"progress_percent"`
}

type updateProfileReq struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Timezone  *string `json:"timezone"`
}

// registerProfile registers profile and stats routes (authed).
func (h *Handlers) registerProfile(authed *gin.RouterGroup) {
	g := authed.Group("/profile")
	{
		g.GET("", h.getProfile)
		g.PATCH("", h.updateProfile)
	}
	g.PUT("/daily-report", h.setDailyReport)
	g.GET("/stats", h.getStats)
}

// setDailyReport toggles the nightly report for the user.
func (h *Handlers) setDailyReport(c *gin.Context) {
	user := currentUser(c)
	var req struct {
		Enabled *bool `json:"enabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "enabled الزامی است")
		return
	}
	if err := h.profiles.SetDailyReport(c.Request.Context(), user.ID, *req.Enabled); err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	c.JSON(http.StatusOK, gin.H{"daily_report": *req.Enabled})
}

// getProfile GET /api/profile
func (h *Handlers) getProfile(c *gin.Context) {
	user := currentUser(c)
	u, st, err := h.profiles.Profile(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	c.JSON(http.StatusOK, profileResponse{
		ID:          u.ID.String(),
		TelegramID:  u.TelegramID,
		Username:    u.Username,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Timezone:    u.Timezone,
		DailyReport: u.DailyReport,
		Pending:     st.Pending,
		Completed:   st.Completed,
		Cancelled:   st.Cancelled,
		ProgressPct: st.CompletionRate(),
	})
}

// updateProfile PATCH /api/profile
func (h *Handlers) updateProfile(c *gin.Context) {
	user := currentUser(c)
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "ورودی نامعتبر است")
		return
	}
	if req.FirstName == nil && req.LastName == nil && req.Timezone == nil {
		fail(c, http.StatusBadRequest, "حداقل یک فیلد برای تغییر لازم است")
		return
	}

	u, err := h.profiles.UpdateProfile(c.Request.Context(), user.ID, req.FirstName, req.LastName, req.Timezone)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         u.ID.String(),
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"timezone":   u.Timezone,
		"message":    "اطلاعات بروزرسانی شد",
	})
}

// getStats GET /api/stats
func (h *Handlers) getStats(c *gin.Context) {
	user := currentUser(c)
	_, st, err := h.profiles.Profile(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pending_tasks":    st.Pending,
		"completed_tasks":  st.Completed,
		"cancelled_tasks":  st.Cancelled,
		"progress_percent": st.CompletionRate(),
	})
}
