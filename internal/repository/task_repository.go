package repository

import (
	"context"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskRepository encapsulates task table queries.
type TaskRepository struct{ db *gorm.DB }

func NewTask(db *gorm.DB) *TaskRepository { return &TaskRepository{db} }
func (r *TaskRepository) Create(c context.Context, v *domain.Task) error {
	return r.db.WithContext(c).Create(v).Error
}
func (r *TaskRepository) GetByID(c context.Context, id uuid.UUID) (*domain.Task, error) {
	var v domain.Task
	return &v, r.db.WithContext(c).First(&v, "id = ?", id).Error
}
func (r *TaskRepository) ListByAssignee(c context.Context, id uuid.UUID) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).Where("assignee_id = ?", id).Order("due_at nulls last, created_at desc").Find(&v).Error
}
func (r *TaskRepository) ListPendingByAssignee(c context.Context, id uuid.UUID) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).Where("assignee_id = ? and status = ?", id, "pending").Order("due_at nulls last").Find(&v).Error
}
func (r *TaskRepository) CountByAssigneeAndStatus(c context.Context, id uuid.UUID, status string) (int64, error) {
	var n int64
	q := r.db.WithContext(c).Model(&domain.Task{}).Where("assignee_id = ?", id)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return n, q.Count(&n).Error
}
func (r *TaskRepository) ListByAssigneePaged(c context.Context, id uuid.UUID, status string, limit, offset int) ([]domain.Task, error) {
	var v []domain.Task
	q := r.db.WithContext(c).Where("assignee_id = ?", id)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return v, q.Order("created_at desc").Limit(limit).Offset(offset).Find(&v).Error
}
func (r *TaskRepository) SearchByTitle(c context.Context, id uuid.UUID, query string, limit int) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).
		Where("assignee_id = ? AND lower(title) LIKE lower(?)", id, "%"+query+"%").
		Order("created_at desc").Limit(limit).Find(&v).Error
}
func (r *TaskRepository) ListDueSoon(c context.Context, from, to time.Time) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).
		Where("status = ? AND reminded_at IS NULL AND due_at >= ? AND due_at < ?", "pending", from, to).
		Find(&v).Error
}
func (r *TaskRepository) MarkReminded(c context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("reminded_at", at).Error
}
func (r *TaskRepository) ListByOwner(c context.Context, id uuid.UUID) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).Where("owner_id = ?", id).Find(&v).Error
}
func (r *TaskRepository) ListByPlanner(c context.Context, id uuid.UUID) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).Where("planner_id = ?", id).Find(&v).Error
}
func (r *TaskRepository) ListDueBetween(c context.Context, from, to time.Time) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).Where("due_at >= ? and due_at < ?", from, to).Find(&v).Error
}
func (r *TaskRepository) Update(c context.Context, v *domain.Task) error {
	return r.db.WithContext(c).Save(v).Error
}
func (r *TaskRepository) UpdateTitle(c context.Context, id uuid.UUID, title string) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("title", title).Error
}
func (r *TaskRepository) UpdateDescription(c context.Context, id uuid.UUID, text string) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("description", text).Error
}
func (r *TaskRepository) UpdatePriority(c context.Context, id uuid.UUID, p string) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("priority", p).Error
}
func (r *TaskRepository) UpdateDueAt(c context.Context, id uuid.UUID, due *time.Time) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("due_at", due).Error
}
func (r *TaskRepository) SetEvidenceRequired(c context.Context, id uuid.UUID, value bool) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("requires_evidence", value).Error
}
func (r *TaskRepository) Complete(c context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Updates(map[string]any{"status": "completed", "completed_at": at}).Error
}
func (r *TaskRepository) Reopen(c context.Context, id uuid.UUID) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Updates(map[string]any{"status": "pending", "completed_at": nil}).Error
}
func (r *TaskRepository) Cancel(c context.Context, id uuid.UUID) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("status", "cancelled").Error
}
func (r *TaskRepository) Reassign(c context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(c).Model(&domain.Task{}).Where("id = ?", id).Update("assignee_id", userID).Error
}
func (r *TaskRepository) Delete(c context.Context, id uuid.UUID) error {
	return r.db.WithContext(c).Delete(&domain.Task{}, "id = ?", id).Error
}

func (r *TaskRepository) CountMine(c context.Context, id uuid.UUID, status string) (int64, error) {
	var n int64
	q := r.db.WithContext(c).Model(&domain.Task{}).Where("assignee_id = ? AND owner_id = ?", id, id)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return n, q.Count(&n).Error
}

func (r *TaskRepository) ListMinePaged(c context.Context, id uuid.UUID, status string, limit, offset int) ([]domain.Task, error) {
	var v []domain.Task
	q := r.db.WithContext(c).Where("assignee_id = ? AND owner_id = ?", id, id)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	return v, q.Order("created_at desc").Limit(limit).Offset(offset).Find(&v).Error
}

func (r *TaskRepository) CountDelegatedToMe(c context.Context, id uuid.UUID) (int64, error) {
	var n int64
	return n, r.db.WithContext(c).Model(&domain.Task{}).
		Where("assignee_id = ? AND owner_id <> ? AND status = ?", id, id, "pending").
		Count(&n).Error
}

func (r *TaskRepository) ListDelegatedToMePaged(c context.Context, id uuid.UUID, limit, offset int) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).
		Where("assignee_id = ? AND owner_id <> ? AND status = ?", id, id, "pending").
		Order("created_at desc").Limit(limit).Offset(offset).Find(&v).Error
}

func (r *TaskRepository) CountHelpdesk(c context.Context, id uuid.UUID) (int64, error) {
	var n int64
	return n, r.db.WithContext(c).Model(&domain.Task{}).
		Where("owner_id = ? AND assignee_id <> ? AND status = ?", id, id, "pending").
		Count(&n).Error
}

func (r *TaskRepository) ListHelpdeskPaged(c context.Context, id uuid.UUID, limit, offset int) ([]domain.Task, error) {
	var v []domain.Task
	return v, r.db.WithContext(c).
		Where("owner_id = ? AND assignee_id <> ? AND status = ?", id, id, "pending").
		Order("created_at desc").Limit(limit).Offset(offset).Find(&v).Error
}
