// Package service implements task business rules and event logging.
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

// TaskService is the task service layer between bot and repositories.
type TaskService struct {
	db       *gorm.DB
	tasks    *repository.TaskRepository
	evidence *repository.EvidenceRepository
	users    *repository.UserRepository
}

// NewTask builds the task service.
func NewTask(db *gorm.DB) *TaskService {
	return &TaskService{db: db, tasks: repository.NewTask(db), evidence: repository.NewEvidence(db), users: repository.NewUser(db)}
}

// SavePhotoEvidence stores a photo the user sent as proof for a task.
func (s *TaskService) SavePhotoEvidence(ctx context.Context, taskID, userID uuid.UUID, fileID string) error {
	e := &domain.TaskEvidence{
		TaskID:         taskID,
		SubmittedByID:  userID,
		Kind:           "photo",
		TelegramFileID: fileID,
	}
	if err := s.evidence.Create(ctx, e); err != nil {
		return err
	}
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil
	}
	return s.log(ctx, &task.AssigneeID, "task", task.ID, "proof_added", nil)
}

// LatestEvidence returns the newest proof of a task, if any.
func (s *TaskService) LatestEvidence(ctx context.Context, taskID uuid.UUID) (*domain.TaskEvidence, error) {
	return s.evidence.LatestForTask(ctx, taskID)
}

// Create saves a new task; when no due date is given it defaults to 20 hours later.
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

// Complete marks a task done and returns the updated copy.
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

// Today returns the open tasks of a user.
func (s *TaskService) Today(ctx context.Context, userID uuid.UUID) ([]domain.Task, error) {
	return s.tasks.ListPendingByAssignee(ctx, userID)
}

// CountOpen counts the open tasks of a user.
func (s *TaskService) CountOpen(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Model(&domain.Task{}).
		Where("assignee_id = ? AND status = ?", userID, "pending").
		Count(&n).Error
	return n, err
}

// log writes an event into ActivityLog.
func (s *TaskService) log(ctx context.Context, userID *uuid.UUID, kind string, id uuid.UUID, action string, meta any) error {
	raw, _ := json.Marshal(meta)
	return s.db.WithContext(ctx).Create(&domain.ActivityLog{UserID: userID, EntityType: kind, EntityID: id, Action: action, Metadata: string(raw), OccurredAt: time.Now()}).Error
}

// GetByID returns a task by id.
func (s *TaskService) GetByID(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	return s.tasks.GetByID(ctx, taskID)
}

// ListByOwnerAndAssignee lists tasks an owner gave to a user.
func (s *TaskService) ListByOwnerAndAssignee(ctx context.Context, ownerID, assigneeID uuid.UUID) ([]domain.Task, error) {
	var tasks []domain.Task
	if err := s.db.WithContext(ctx).Where("owner_id = ? AND assignee_id = ?", ownerID, assigneeID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// CreateForOtherUser saves a task on behalf of another user and logs the event.
func (s *TaskService) CreateForOtherUser(ctx context.Context, t *domain.Task) error {
	if err := s.tasks.Create(ctx, t); err != nil {
		return err
	}
	_ = s.log(ctx, &t.OwnerID, "task", t.ID, "created_for_other", map[string]string{"title": t.Title})
	return nil
}

// Delete removes a task and logs the event.
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

// UpdateTitle changes the task title.
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

// UpdateDescription changes the task description.
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

// dueRange turns a time filter into a due date window in the user timezone.
func dueRange(filter string, now time.Time, tz string) (time.Time, time.Time, bool) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	n := now.In(loc)
	switch filter {
	case "today":
		y, m, d := n.Date()
		start := time.Date(y, m, d, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 0, 1), true
	case "tomorrow":
		start := n.AddDate(0, 0, 1)
		y, m, d := start.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, loc), time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, 1), true
	case "upcoming":
		return n, n.AddDate(0, 0, 7), true
	case "overdue":
		return time.Time{}, n, true
	}
	return n, n, false
}

// TaskStats holds the task stats of a user.
type TaskStats struct {
	Pending   int64
	Completed int64
	Cancelled int64
}

// CompletionRate returns the user progress percentage.
func (t TaskStats) CompletionRate() float64 {
	total := t.Pending + t.Completed + t.Cancelled
	if total == 0 {
		return 0
	}
	return float64(t.Completed) / float64(total) * 100
}

// Reopen reopens a completed task. Only tasks that require evidence
// can be reopened, so the assignee sends a new proof.
func (s *TaskService) Reopen(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	task, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status == "pending" {
		return task, nil
	}
	if !task.RequiresEvidence {
		return nil, fmt.Errorf("reopen is only allowed for tasks that require evidence")
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

// UpdatePriority changes the task priority.
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

// UpdateDueAt changes the task due date; nil clears it.
func (s *TaskService) UpdateDueAt(ctx context.Context, taskID uuid.UUID, due *time.Time) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.tasks.UpdateDueAt(ctx, taskID, due); err != nil {
		return err
	}
	// changing the due date voids an earlier reminder.
	_ = s.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", taskID).Update("reminded_at", nil).Error
	return s.log(ctx, &task.OwnerID, "task", task.ID, "updated_due", nil)
}

// ListFiltered returns user tasks with filter and pagination.
// filter is one of pending (own), from_others, helpdesk, completed, cancelled or all.
func (s *TaskService) ListFiltered(ctx context.Context, userID uuid.UUID, filter string, page, size int) ([]domain.Task, int64, error) {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * size

	var tasks []domain.Task
	var total int64
	var err error

	switch filter {
	case "today", "tomorrow", "upcoming", "overdue":
		user, uerr := s.users.GetByID(ctx, userID)
		if uerr != nil {
			return nil, 0, uerr
		}
		from, to, ok := dueRange(filter, time.Now(), user.Timezone)
		if !ok {
			return nil, 0, fmt.Errorf("invalid filter: %s", filter)
		}
		if total, terr := s.tasks.CountByDueRange(ctx, userID, from, to); terr == nil {
			tasks, err = s.tasks.ListByDueRange(ctx, userID, from, to, size, offset)
			return tasks, total, err
		}
	case "from_others":
		if total, err = s.tasks.CountDelegatedToMe(ctx, userID); err == nil {
			tasks, err = s.tasks.ListDelegatedToMePaged(ctx, userID, size, offset)
		}
	case "helpdesk":
		if total, err = s.tasks.CountHelpdesk(ctx, userID); err == nil {
			tasks, err = s.tasks.ListHelpdeskPaged(ctx, userID, size, offset)
		}
	default:
		status := ""
		switch filter {
		case "completed":
			status = "completed"
		case "cancelled":
			status = "cancelled"
		case "pending":
			status = "pending"
		}
		if total, err = s.tasks.CountMine(ctx, userID, status); err == nil {
			tasks, err = s.tasks.ListMinePaged(ctx, userID, status, size, offset)
		}
	}
	return tasks, total, err
}

// Search returns user tasks matching a title query.
func (s *TaskService) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]domain.Task, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	return s.tasks.SearchByTitle(ctx, userID, query, limit)
}

// Stats returns the task stats of a user.
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

// DueSoon returns tasks due soon that have not been reminded yet.
func (s *TaskService) DueSoon(ctx context.Context, from, to time.Time) ([]domain.Task, error) {
	return s.tasks.ListDueSoon(ctx, from, to)
}

// SetEvidenceRequired turns the evidence requirement of a task on or off.
func (s *TaskService) SetEvidenceRequired(ctx context.Context, taskID uuid.UUID, required bool) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.tasks.SetEvidenceRequired(ctx, taskID, required); err != nil {
		return err
	}
	action := "evidence_off"
	if required {
		action = "evidence_on"
	}
	return s.log(ctx, &task.OwnerID, "task", task.ID, action, nil)
}

// MarkReminded marks that a reminder was sent for a task.
func (s *TaskService) MarkReminded(ctx context.Context, taskID uuid.UUID, at time.Time) error {
	return s.tasks.MarkReminded(ctx, taskID, at)
}

// Cancel marks a task as cancelled.
func (s *TaskService) Cancel(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", taskID).Update("status", "cancelled").Error; err != nil {
		return err
	}
	return s.log(ctx, &task.OwnerID, "task", task.ID, "cancelled", nil)
}
