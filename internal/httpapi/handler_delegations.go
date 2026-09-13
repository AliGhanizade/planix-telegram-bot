package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// registerDelegations routes for the help desk view (authed).
func (h *Handlers) registerDelegations(authed *gin.RouterGroup) {
	authed.GET("/delegations", h.listDelegations)
}

// listDelegations GET /api/delegations?status=&q=&page=
// tasks this user delegated to others, filterable by assignee name,
// username, telegram id or task title.
func (h *Handlers) listDelegations(c *gin.Context) {
	user := currentUser(c)
	status := c.DefaultQuery("status", "pending")
	switch status {
	case "pending", "completed", "cancelled", "all":
	default:
		fail(c, http.StatusBadRequest, "status نامعتبر است")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	q := c.Query("q")

	tasks, total, err := h.tasks.ListDelegatedFiltered(c.Request.Context(), user.ID, status, q, page, webPageSize)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}

	items := make([]taskResponse, 0, len(tasks))
	for i := range tasks {
		items = append(items, h.toTaskResponseFor(c.Request.Context(), &tasks[i], user.ID))
	}
	pages := int((total + webPageSize - 1) / webPageSize)
	if pages < 1 {
		pages = 1
	}
	c.JSON(http.StatusOK, taskListResponse{Items: items, Page: page, Pages: pages, Total: total, PageSize: webPageSize})
}
