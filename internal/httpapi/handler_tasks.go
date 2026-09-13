package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// errInvalidStatus خطای وضعیت نامعتبر تسک.
var errInvalidStatus = errors.New("invalid status")

// ---- DTO های تسک ----

type taskResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	DueAt       *time.Time `json:"due_at"`
	CompletedAt *time.Time `json:"completed_at"`
	OwnerID     string     `json:"owner_id"`
	AssigneeID  string     `json:"assignee_id"`
	CreatedAt   time.Time  `json:"created_at"`
}

type createTaskReq struct {
	Title       string     `json:"title" binding:"required,max=240"`
	Description string     `json:"description"`
	Priority    string     `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	DueAt       *time.Time `json:"due_at"`
}

type updateTaskReq struct {
	Title       *string          `json:"title" binding:"omitempty,max=240"`
	Description *string          `json:"description"`
	Priority    *string          `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	DueAt       *json.RawMessage `json:"due_at"` // مقدار یا null برای حذف موعد
	Status      *string          `json:"status" binding:"omitempty,oneof=pending completed cancelled"`
}

type taskListResponse struct {
	Items    []taskResponse `json:"items"`
	Page     int            `json:"page"`
	Pages    int            `json:"pages"`
	Total    int64          `json:"total"`
	PageSize int            `json:"page_size"`
}

// webPageSize اندازه صفحه‌ی API — جدا از رابط کاربری بات.
const webPageSize = 20

// registerTasks مسیرهای تسک برای پنل وب (همه احراز هویت‌شده).
func (h *Handlers) registerTasks(authed *gin.RouterGroup) {
	g := authed.Group("/tasks")
	{
		g.GET("", h.listTasks)
		g.POST("", h.createTask)
		g.GET("/:id", h.getTask)
		g.PATCH("/:id", h.updateTask)
		g.DELETE("/:id", h.deleteTask)
		g.POST("/:id/complete", h.completeTask)
		g.POST("/:id/reopen", h.reopenTask)
	}
}

// toTaskResponse مدل دامنه را به DTO تسک تبدیل می‌کند.
func toTaskResponse(t *domain.Task) taskResponse {
	return taskResponse{
		ID:          t.ID.String(),
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Status:      t.Status,
		DueAt:       t.DueAt,
		CompletedAt: t.CompletedAt,
		OwnerID:     t.OwnerID.String(),
		AssigneeID:  t.AssigneeID.String(),
		CreatedAt:   t.CreatedAt,
	}
}

// listTasks GET /api/tasks?status=&page=&q=
func (h *Handlers) listTasks(c *gin.Context) {
	user := currentUser(c)
	status := c.DefaultQuery("status", "pending")
	switch status {
	case "pending", "completed", "all", "cancelled":
	default:
		fail(c, http.StatusBadRequest, "status نامعتبر است")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// جستجو: q
	if q := c.Query("q"); q != "" {
		tasks, err := h.tasks.Search(c.Request.Context(), user.ID, q, 50)
		if err != nil {
			fail(c, http.StatusInternalServerError, "خطای داخلی")
			return
		}
		items := make([]taskResponse, 0, len(tasks))
		for i := range tasks {
			items = append(items, toTaskResponse(&tasks[i]))
		}
		c.JSON(http.StatusOK, taskListResponse{Items: items, Page: 1, Pages: 1, Total: int64(len(items)), PageSize: len(items)})
		return
	}

	tasks, total, err := h.tasks.ListFiltered(c.Request.Context(), user.ID, status, page, webPageSize)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	items := make([]taskResponse, 0, len(tasks))
	for i := range tasks {
		items = append(items, toTaskResponse(&tasks[i]))
	}
	pages := int((total + webPageSize - 1) / webPageSize)
	if pages < 1 {
		pages = 1
	}
	c.JSON(http.StatusOK, taskListResponse{Items: items, Page: page, Pages: pages, Total: total, PageSize: webPageSize})
}

// createTask POST /api/tasks
func (h *Handlers) createTask(c *gin.Context) {
	user := currentUser(c)
	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "عنوان الزامی است")
		return
	}

	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}
	task := &domain.Task{
		OwnerID:     user.ID,
		AssigneeID:  user.ID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		Status:      "pending",
		DueAt:       req.DueAt,
	}
	if err := h.tasks.Create(c.Request.Context(), task); err != nil {
		fail(c, http.StatusInternalServerError, "خطا در ثبت تسک")
		return
	}
	c.JSON(http.StatusCreated, toTaskResponse(task))
}

// loadAccessibleTask تسک را واکشی و دسترسی کاربر را بررسی می‌کند.
func (h *Handlers) loadAccessibleTask(c *gin.Context) (*domain.Task, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return nil, false
	}
	user := currentUser(c)
	task, err := h.tasks.GetByID(c.Request.Context(), id)
	if err != nil {
		fail(c, http.StatusNotFound, "تسک پیدا نشد")
		return nil, false
	}
	if task.AssigneeID != user.ID && task.OwnerID != user.ID {
		fail(c, http.StatusForbidden, "به این تسک دسترسی نداری")
		return nil, false
	}
	return task, true
}

// getTask GET /api/tasks/:id
func (h *Handlers) getTask(c *gin.Context) {
	task, ok := h.loadAccessibleTask(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(task))
}

// updateTask PATCH /api/tasks/:id
func (h *Handlers) updateTask(c *gin.Context) {
	task, ok := h.loadAccessibleTask(c)
	if !ok {
		return
	}
	var req updateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "ورودی نامعتبر است")
		return
	}
	ctx := c.Request.Context()

	if req.Title != nil {
		if err := h.tasks.UpdateTitle(ctx, task.ID, *req.Title); err != nil {
			fail(c, http.StatusBadRequest, "عنوان نامعتبر است")
			return
		}
	}
	if req.Description != nil {
		if err := h.tasks.UpdateDescription(ctx, task.ID, *req.Description); err != nil {
			fail(c, http.StatusInternalServerError, "خطا در بروزرسانی")
			return
		}
	}
	if req.Priority != nil {
		if err := h.tasks.UpdatePriority(ctx, task.ID, *req.Priority); err != nil {
			fail(c, http.StatusBadRequest, "اولویت نامعتبر است")
			return
		}
	}
	if req.DueAt != nil {
		if string(*req.DueAt) == "null" {
			if err := h.tasks.UpdateDueAt(ctx, task.ID, nil); err != nil {
				fail(c, http.StatusInternalServerError, "خطا در حذف موعد")
				return
			}
		} else {
			var due time.Time
			if err := json.Unmarshal(*req.DueAt, &due); err != nil {
				fail(c, http.StatusBadRequest, "due_at باید ISO 8601 یا null باشد")
				return
			}
			if err := h.tasks.UpdateDueAt(ctx, task.ID, &due); err != nil {
				fail(c, http.StatusInternalServerError, "خطا در ثبت موعد")
				return
			}
		}
	}
	if req.Status != nil {
		if _, err := h.applyStatus(ctx, task.ID, *req.Status); err != nil {
			fail(c, http.StatusBadRequest, "وضعیت نامعتبر است")
			return
		}
	}

	updated, err := h.tasks.GetByID(ctx, task.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(updated))
}

// applyStatus وضعیت تسک را با سرویس تغییر می‌دهد.
func (h *Handlers) applyStatus(ctx context.Context, taskID uuid.UUID, status string) (*domain.Task, error) {
	switch status {
	case "completed":
		return h.tasks.Complete(ctx, taskID)
	case "pending":
		return h.tasks.Reopen(ctx, taskID)
	case "cancelled":
		if err := h.tasks.Cancel(ctx, taskID); err != nil {
			return nil, err
		}
		return h.tasks.GetByID(ctx, taskID)
	}
	return nil, errInvalidStatus
}

// deleteTask DELETE /api/tasks/:id
func (h *Handlers) deleteTask(c *gin.Context) {
	task, ok := h.loadAccessibleTask(c)
	if !ok {
		return
	}
	if err := h.tasks.Delete(c.Request.Context(), task.ID); err != nil {
		fail(c, http.StatusInternalServerError, "خطا در حذف")
		return
	}
	c.Status(http.StatusNoContent)
}

// completeTask POST /api/tasks/:id/complete
func (h *Handlers) completeTask(c *gin.Context) {
	task, ok := h.loadAccessibleTask(c)
	if !ok {
		return
	}
	updated, err := h.tasks.Complete(c.Request.Context(), task.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطا در انجام تسک")
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(updated))
}

// reopenTask POST /api/tasks/:id/reopen
func (h *Handlers) reopenTask(c *gin.Context) {
	task, ok := h.loadAccessibleTask(c)
	if !ok {
		return
	}
	updated, err := h.tasks.Reopen(c.Request.Context(), task.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطا در بازگشایی")
		return
	}
	c.JSON(http.StatusOK, toTaskResponse(updated))
}
