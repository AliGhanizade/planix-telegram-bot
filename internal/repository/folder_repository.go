package repository

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FolderRepository holds folder and task-folder link queries.
type FolderRepository struct{ db *gorm.DB }

func NewFolder(db *gorm.DB) *FolderRepository { return &FolderRepository{db} }

// Create stores a new folder.
func (r *FolderRepository) Create(ctx context.Context, f *domain.Folder) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// GetByID returns a folder by id.
func (r *FolderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Folder, error) {
	var v domain.Folder
	return &v, r.db.WithContext(ctx).First(&v, "id = ?", id).Error
}

// Delete removes a folder and its task links. Child folders become roots.
func (r *FolderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("folder_id = ?", id).Delete(&domain.TaskFolderLink{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&domain.Folder{}).Where("parent_id = ?", id).Update("parent_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.Folder{}, "id = ?", id).Error
	})
}

// ListOwned returns the folders a user created.
func (r *FolderRepository) ListOwned(ctx context.Context, ownerID uuid.UUID) ([]domain.Folder, error) {
	var v []domain.Folder
	return v, r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at asc").Find(&v).Error
}

// ListShared returns folders created by others that contain at least one
// task assigned to this user by that folder owner (delegation folders).
func (r *FolderRepository) ListShared(ctx context.Context, userID uuid.UUID) ([]domain.Folder, error) {
	var v []domain.Folder
	err := r.db.WithContext(ctx).
		Where("owner_id <> ? AND id IN (?)", userID,
			r.db.Model(&domain.TaskFolderLink{}).
				Select("DISTINCT task_folder_links.folder_id").
				Joins("JOIN tasks ON tasks.id = task_folder_links.task_id").
				Where("tasks.assignee_id = ? AND tasks.owner_id <> ?", userID, userID),
		).
		Order("created_at asc").Find(&v).Error
	return v, err
}

// Link puts a task into a folder.
func (r *FolderRepository) Link(ctx context.Context, taskID, folderID uuid.UUID) error {
	link := domain.TaskFolderLink{TaskID: taskID, FolderID: folderID}
	return r.db.WithContext(ctx).Where(link).FirstOrCreate(&link).Error
}

// Unlink removes a task from a folder.
func (r *FolderRepository) Unlink(ctx context.Context, taskID, folderID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("task_id = ? AND folder_id = ?", taskID, folderID).
		Delete(&domain.TaskFolderLink{}).Error
}

// LinksForTask returns the folder ids a task belongs to.
func (r *FolderRepository) LinksForTask(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	var links []domain.TaskFolderLink
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Find(&links).Error
	ids := make([]uuid.UUID, 0, len(links))
	for _, l := range links {
		ids = append(ids, l.FolderID)
	}
	return ids, err
}

// TasksInFolder returns the tasks of a folder that the user may see
// (assignee or owner), newest first.
func (r *FolderRepository) TasksInFolder(ctx context.Context, folderID, userID uuid.UUID) ([]domain.Task, error) {
	var v []domain.Task
	err := r.db.WithContext(ctx).
		Where("id IN (?) AND (assignee_id = ? OR owner_id = ?)",
			r.db.Model(&domain.TaskFolderLink{}).Select("task_id").Where("folder_id = ?", folderID),
			userID, userID,
		).
		Order("created_at desc").Find(&v).Error
	return v, err
}
