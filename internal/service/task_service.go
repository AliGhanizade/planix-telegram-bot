// Package service قواعد کسب‌وکار تسک‌ها و ثبت رویدادها را پیاده‌سازی می‌کند.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskService لایه‌ی سرویس تسک‌ها بین بات و مخازن داده است.
type TaskService struct {
	db    *gorm.DB
	tasks *repository.TaskRepository
}

// NewTask سرویس تسک را می‌سازد.
func NewTask(db *gorm.DB) *TaskService { return &TaskService{db: db, tasks: repository.NewTask(db)} }

// Create تسک جدید ثبت می‌کند؛ اگر موعدی مشخص نشده باشد، مهلت پیش‌فرض ۲۰ ساعت بعد در نظر گرفته می‌شود.
func (s *TaskService) Create(ctx context.Context, task *domain.Task) error {
	if task.Title == "" {
		return fmt.Errorf("task title is required")
	}
	if task.DueAt == nil || task.DueAt.IsZero() {
		due := time.Now().Add(20 * time.Hour)
		task.DueAt = &due
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return err
	}
	return s.log(ctx, &task.AssigneeID, "task", task.ID, "created", map[string]string{"title": task.Title})
}

// Complete تسک را انجام‌شده علامت می‌زند و نسخه‌ی به‌روزشده را برمی‌گرداند.
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

// Today تسک‌های باز کاربر را برمی‌گرداند.
func (s *TaskService) Today(ctx context.Context, userID uuid.UUID) ([]domain.Task, error) {
	return s.tasks.ListPendingByAssignee(ctx, userID)
}

// CountOpen تعداد تسک‌های باز کاربر را برمی‌گرداند.
func (s *TaskService) CountOpen(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&domain.Task{}).
		Where("assignee_id = ? AND status = ?", userID, "pending").
		Count(&n).Error
	return n, err
}

// log رویداد را در جدول ActivityLog ثبت می‌کند.
func (s *TaskService) log(ctx context.Context, userID *uuid.UUID, kind string, id uuid.UUID, action string, meta any) error {
	raw, _ := json.Marshal(meta)
	return s.db.WithContext(ctx).Create(&domain.ActivityLog{UserID: userID, EntityType: kind, EntityID: id, Action: action, Metadata: string(raw), OccurredAt: time.Now()}).Error
}

// GetByID یک تسک را با شناسه برمی‌گرداند.
func (s *TaskService) GetByID(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	return s.tasks.GetByID(ctx, taskID)
}

// ListByOwnerAndAssignee تسک‌هایی که مالک به کاربر داده است را فهرست می‌کند.
func (s *TaskService) ListByOwnerAndAssignee(ctx context.Context, ownerID, assigneeID uuid.UUID) ([]domain.Task, error) {
	var tasks []domain.Task
	if err := s.db.WithContext(ctx).Where("owner_id = ? AND assignee_id = ?", ownerID, assigneeID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// CreateForOtherUser تسک را به‌جای کاربر دیگر ثبت و رویدادش را لاگ می‌کند.
func (s *TaskService) CreateForOtherUser(ctx context.Context, t *domain.Task) error {
	if err := s.tasks.Create(ctx, t); err != nil {
		return err
	}
	_ = s.log(ctx, &t.OwnerID, "task", t.ID, "created_for_other", map[string]string{"title": t.Title})
	return nil
}

// Delete تسک را حذف و رویدادش را لاگ می‌کند.
func (s *TaskService) Delete(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Delete(&domain.Task{}, "id = ?", taskID).Error; err != nil {
		return err
	}
	return s.log(ctx, &task.OwnerID, "task", task.ID, "deleted", nil)
}

// UpdateTitle عنوان تسک را تغییر می‌دهد.
func (s *TaskService) UpdateTitle(ctx context.Context, taskID uuid.UUID, title string) error {
	if title == "" {
		return fmt.Errorf("title is required")
	}
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", taskID).Update("title", title).Error; err != nil {
		return err
	}
	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_title", map[string]string{"title": title})
}

// UpdateDescription توضیحات تسک را تغییر می‌دهد.
func (s *TaskService) UpdateDescription(ctx context.Context, taskID uuid.UUID, description string) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", taskID).Update("description", description).Error; err != nil {
		return err
	}
	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_description", nil)
}

// TaskStats آمار تسک‌های یک کاربر است.
type TaskStats struct {
	Pending   int64
	Completed int64
	Cancelled int64
}

// CompletionRate درصد پیشرفت کاربر را برمی‌گرداند.
func (t TaskStats) CompletionRate() float64 {
	total := t.Pending + t.Completed + t.Cancelled
	if total == 0 {
		return 0
	}
	return float64(t.Completed) / float64(total) * 100
}

// Reopen تسکِ انجام‌شده را دوباره باز می‌کند.
func (s *TaskService) Reopen(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status == "pending" {
		return task, nil
	}
	if err = s.tasks.Reopen(ctx, taskID); err != nil {
		return nil, err
	}
	task.Status = "pending"
	task.CompletedAt = nil
	task.RemindedAt = nil
	_ = s.log(ctx, &task.AssigneeID, "task", task.ID, "reopened", nil)
	return task, nil
}

// UpdatePriority اولویت تسک را تغییر می‌دهد.
func (s *TaskService) UpdatePriority(ctx context.Context, taskID uuid.UUID, priority string) error {
	switch priority {
	case "low", "normal", "high", "urgent":
	default:
		return fmt.Errorf("invalid priority: %s", priority)
	}
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.tasks.UpdatePriority(ctx, taskID, priority); err != nil {
		return err
	}
	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_priority", map[string]string{"priority": priority})
}

// UpdateDueAt موعد تسک را تغییر می‌دهد؛ nil یعنی حذف موعد.
func (s *TaskService) UpdateDueAt(ctx context.Context, taskID uuid.UUID, due *time.Time) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.tasks.UpdateDueAt(ctx, taskID, due); err != nil {
		return err
	}
	// با تغییر موعد، یادآوری قبلی بی‌معنا می‌شود.
	_ = s.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", taskID).Update("reminded_at", nil).Error
	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_due", nil)
}

// ListFiltered تسک‌های کاربر را با فیلتر وضعیت و صفحه‌بندی برمی‌گرداند.
// filter می‌تواند pending، completed یا all باشد.
func (s *TaskService) ListFiltered(ctx context.Context, userID uuid.UUID, filter string, page, size int) ([]domain.Task, int64, error) {
	status := ""
	switch filter {
	case "pending":
		status = "pending"
	case "completed":
		status = "completed"
	}
	if page < 1 {
		page = 1
	}
	total, err := s.tasks.CountByAssigneeAndStatus(ctx, userID, status)
	if err != nil {
		return nil, 0, err
	}
	tasks, err := s.tasks.ListByAssigneePaged(ctx, userID, status, size, (page-1)*size)
	return tasks, total, err
}

// Search تسک‌های کاربر را با جستجوی عنوان برمی‌گرداند.
func (s *TaskService) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]domain.Task, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	return s.tasks.SearchByTitle(ctx, userID, query, limit)
}

// Stats آمار تسک‌های کاربر را برمی‌گرداند.
func (s *TaskService) Stats(ctx context.Context, userID uuid.UUID) (*TaskStats, error) {
	st := &TaskStats{}
	var err error
	if st.Pending, err = s.tasks.CountByAssigneeAndStatus(ctx, userID, "pending"); err != nil {
		return nil, err
	}
	if st.Completed, err = s.tasks.CountByAssigneeAndStatus(ctx, userID, "completed"); err != nil {
		return nil, err
	}
	if st.Cancelled, err = s.tasks.CountByAssigneeAndStatus(ctx, userID, "cancelled"); err != nil {
		return nil, err
	}
	return st, nil
}

// DueSoon تسک‌هایی که موعدشان نزدیک است و یادآوری نشده‌اند را برمی‌گرداند.
func (s *TaskService) DueSoon(ctx context.Context, from, to time.Time) ([]domain.Task, error) {
	return s.tasks.ListDueSoon(ctx, from, to)
}

// MarkReminded ثبت می‌کند که برای تسک یادآوری فرستاده شده است.
func (s *TaskService) MarkReminded(ctx context.Context, taskID uuid.UUID, at time.Time) error {
	return s.tasks.MarkReminded(ctx, taskID, at)
}
