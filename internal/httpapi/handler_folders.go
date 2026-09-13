package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// registerFolders routes for folder management (authed).
func (h *Handlers) registerFolders(authed *gin.RouterGroup, _ *gin.RouterGroup) {
	g := authed.Group("/folders")
	{
		g.GET("", h.listFolders)
		g.POST("", h.createFolder)
		g.DELETE("/:id", h.deleteFolder)
		g.GET("/:id/tasks", h.folderTasks)
	}
	authed.POST("/tasks/:id/folders", h.linkTaskFolder)
	authed.DELETE("/tasks/:id/folders/:folderId", h.unlinkTaskFolder)
}

type folderResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Kind     string  `json:"kind"` // own | shared
	ParentID *string `json:"parent_id"`
}

type createFolderReq struct {
	Name     string `json:"name" binding:"required,max=80"`
	ParentID string `json:"parent_id"`
}

// listFolders GET /api/folders
func (h *Handlers) listFolders(c *gin.Context) {
	user := currentUser(c)
	views, err := h.folders.List(c.Request.Context(), user.ID)
	if err != nil {
		fail(c, http.StatusInternalServerError, "خطای داخلی")
		return
	}
	items := make([]folderResponse, 0, len(views))
	for _, v := range views {
		items = append(items, folderResponse{
			ID:   v.Folder.ID.String(),
			Name: v.Folder.Name,
			Kind: v.Kind,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// createFolder POST /api/folders
func (h *Handlers) createFolder(c *gin.Context) {
	user := currentUser(c)
	var req createFolderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "نام پوشه الزامی است")
		return
	}
	var parentID *uuid.UUID
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err != nil {
			fail(c, http.StatusBadRequest, "پوشه‌ی والد نامعتبر است")
			return
		}
		parentID = &pid
	}
	f, err := h.folders.Create(c.Request.Context(), user.ID, req.Name, parentID)
	if err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, folderResponse{ID: f.ID.String(), Name: f.Name, Kind: "own"})
}

// deleteFolder DELETE /api/folders/:id
func (h *Handlers) deleteFolder(c *gin.Context) {
	user := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return
	}
	if err := h.folders.Delete(c.Request.Context(), user.ID, id); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// folderTasks GET /api/folders/:id/tasks
func (h *Handlers) folderTasks(c *gin.Context) {
	user := currentUser(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return
	}
	tasks, err := h.folders.TasksInFolder(c.Request.Context(), user.ID, id)
	if err != nil {
		fail(c, http.StatusNotFound, "پوشه پیدا نشد")
		return
	}
	items := make([]taskResponse, 0, len(tasks))
	for i := range tasks {
		items = append(items, h.toTaskResponse(c.Request.Context(), &tasks[i]))
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// linkTaskFolder puts a task into a folder.
func (h *Handlers) linkTaskFolder(c *gin.Context) {
	user := currentUser(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return
	}
	var req struct {
		FolderID string `json:"folder_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "folder_id الزامی است")
		return
	}
	folderID, err := uuid.Parse(req.FolderID)
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return
	}
	if err := h.folders.Link(c.Request.Context(), user.ID, taskID, folderID); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "تسک به پوشه اضافه شد"})
}

// unlinkTaskFolder removes a task from a folder.
func (h *Handlers) unlinkTaskFolder(c *gin.Context) {
	user := currentUser(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return
	}
	folderID, err := uuid.Parse(c.Param("folderId"))
	if err != nil {
		fail(c, http.StatusBadRequest, "شناسه نامعتبر است")
		return
	}
	if err := h.folders.Unlink(c.Request.Context(), user.ID, taskID, folderID); err != nil {
		fail(c, http.StatusBadRequest, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
