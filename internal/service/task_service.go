package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskService struct {
	db    *gorm.DB
	tasks *repository.TaskRepository
}

func NewTask(db *gorm.DB) *TaskService { return &TaskService{db: db, tasks: repository.NewTask(db)} }

func (s *TaskService) Create(ctx context.Context, task *domain.Task) error {
	if task.Title == "" {
		return fmt.Errorf("task title is required")
	}
	
	due := time.Now().Add(20 * time.Hour)
	task.DueAt = &due
	if task.DueAt.IsZero() {
		due := time.Now().Add(20 * time.Hour)
		task.DueAt = &due
	}

	if err := s.tasks.Create(ctx, task); err != nil {
		return err
	}

	return s.log(ctx, &task.AssigneeID, "task", task.ID, "created", map[string]string{"title": task.Title})
}

func (s *TaskService) Complete(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status == "completed" {
		return task, nil
	}
	now := time.Now()
	if err = s.tasks.Complete(ctx, taskID, now); err != nil {
		return nil, err
	}
	task.Status = "completed"
	task.CompletedAt = &now
	_ = s.log(ctx, &task.AssigneeID, "task", task.ID, "completed", nil)
	return task, nil
}

func (s *TaskService) Today(ctx context.Context, userID uuid.UUID) ([]domain.Task, error) {
	return s.tasks.ListPendingByAssignee(ctx, userID)
}

func (s *TaskService) log(ctx context.Context, userID *uuid.UUID, kind string, id uuid.UUID, action string, meta any) error {
	raw, _ := json.Marshal(meta)
	return s.db.WithContext(ctx).Create(&domain.ActivityLog{UserID: userID, EntityType: kind, EntityID: id, Action: action, Metadata: string(raw), OccurredAt: time.Now()}).Error
}

func (s *TaskService) GetByID(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	return s.tasks.GetByID(ctx, taskID)
}

func (s *TaskService) ListByOwnerAndAssignee(ctx context.Context, ownerID, assigneeID uuid.UUID) ([]domain.Task, error) {
	var tasks []domain.Task
	if err := s.db.WithContext(ctx).Where("owner_id = ? AND assignee_id = ?", ownerID, assigneeID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskService) CreateForOtherUser(ctx context.Context, t *domain.Task) error {

	if err := s.tasks.Create(ctx, t); err != nil {
		return err
	}
	_ = s.log(ctx, &t.OwnerID, "task", t.ID, "created_for_other", map[string]string{"title": t.Title})
	return nil
}

func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).
		Delete(&domain.Task{}, "id = ?", taskID).Error; err != nil {
		return err
	}

	return s.log(ctx, &task.OwnerID, "task", task.ID, "deleted", nil)
}

func (s *TaskService) UpdateTitle(ctx context.Context, taskID uuid.UUID, title string) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}

	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).
		Model(&domain.Task{}).
		Where("id = ?", taskID).
		Update("title", title).Error; err != nil {
		return err
	}

	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_title", map[string]string{
		"title": title,
	})
}

func (s *TaskService) UpdateDescription(ctx context.Context, taskID uuid.UUID, description string) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if err := s.db.WithContext(ctx).
		Model(&domain.Task{}).
		Where("id = ?", taskID).
		Update("description", description).Error; err != nil {
		return err
	}

	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_description", nil)
}
