package repository

import (
	"context"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repositories keep database queries out of Telegram and HTTP handlers.
type UserRepository struct{ db *gorm.DB }

func NewUser(db *gorm.DB) *UserRepository { return &UserRepository{db} }
func (r *UserRepository) Create(c context.Context, u *domain.User) error {
	return r.db.WithContext(c).Create(u).Error
}
func (r *UserRepository) GetByID(c context.Context, id uuid.UUID) (*domain.User, error) {
	var v domain.User
	return &v, r.db.WithContext(c).First(&v, "id = ?", id).Error
}
func (r *UserRepository) GetByTelegramID(c context.Context, id int64) (*domain.User, error) {
	var v domain.User
	return &v, r.db.WithContext(c).Where("telegram_id = ?", id).First(&v).Error
}
func (r *UserRepository) GetByUsername(c context.Context, username string) (*domain.User, error) {
	var v domain.User
	return &v, r.db.WithContext(c).Where("lower(username) = lower(?)", username).First(&v).Error
}
func (r *UserRepository) List(c context.Context, limit, offset int) ([]domain.User, error) {
	var v []domain.User
	return v, r.db.WithContext(c).Limit(limit).Offset(offset).Find(&v).Error
}
func (r *UserRepository) ListActive(c context.Context) ([]domain.User, error) {
	var v []domain.User
	return v, r.db.WithContext(c).Where("is_active = true").Find(&v).Error
}
func (r *UserRepository) Update(c context.Context, u *domain.User) error {
	return r.db.WithContext(c).Save(u).Error
}
func (r *UserRepository) UpdateName(c context.Context, id uuid.UUID, first, last string) error {
	return r.db.WithContext(c).Model(&domain.User{}).Where("id = ?", id).Updates(map[string]any{"first_name": first, "last_name": last}).Error
}
func (r *UserRepository) UpdateUsername(c context.Context, id uuid.UUID, username string) error {
	return r.db.WithContext(c).Model(&domain.User{}).Where("id = ?", id).Update("username", username).Error
}
func (r *UserRepository) UpdateTimezone(c context.Context, id uuid.UUID, tz string) error {
	return r.db.WithContext(c).Model(&domain.User{}).Where("id = ?", id).Update("timezone", tz).Error
}
func (r *UserRepository) MarkSeen(c context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(c).Model(&domain.User{}).Where("id = ?", id).Update("last_seen_at", at).Error
}
func (r *UserRepository) Activate(c context.Context, id uuid.UUID) error {
	return r.db.WithContext(c).Model(&domain.User{}).Where("id = ?", id).Update("is_active", true).Error
}
func (r *UserRepository) Deactivate(c context.Context, id uuid.UUID) error {
	return r.db.WithContext(c).Model(&domain.User{}).Where("id = ?", id).Update("is_active", false).Error
}
func (r *UserRepository) Count(c context.Context) (int64, error) {
	var n int64
	return n, r.db.WithContext(c).Model(&domain.User{}).Count(&n).Error
}
func (r *UserRepository) Delete(c context.Context, id uuid.UUID) error {
	return r.db.WithContext(c).Delete(&domain.User{}, "id = ?", id).Error
}
func (r *UserRepository) UpsertTelegramUser(c context.Context, u *domain.User) error {
	return r.db.WithContext(c).Where("telegram_id = ?", u.TelegramID).Assign(map[string]any{"username": u.Username, "first_name": u.FirstName, "last_name": u.LastName, "language_code": u.LanguageCode, "last_seen_at": time.Now()}).FirstOrCreate(u).Error
}
func (r *UserRepository) ListTaskContacts(ctx context.Context, id uuid.UUID) ([]domain.User, error) {
	var users []domain.User

	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where(`
			id IN (
				SELECT assignee_id FROM tasks WHERE owner_id = ? OR planner_id = ?
				UNION
				SELECT owner_id FROM tasks WHERE assignee_id = ?
				UNION
				SELECT planner_id FROM tasks WHERE assignee_id = ?
			)
		`, id, id, id, id).
		Where("id <> ?", id).
		Find(&users).Error

	return users, err
}

func (r *UserRepository) ListUsersAssignedTo(ctx context.Context, id uuid.UUID) ([]domain.User, error) {
	var users []domain.User

	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id IN (?)",
			r.db.Model(&domain.Task{}).
				Select("DISTINCT assignee_id").
				Where("owner_id = ? OR planner_id = ?", id, id),
		).
		Find(&users).Error

	return users, err
}

func (r *UserRepository) ListUsersWhoAssignedTo(ctx context.Context, id uuid.UUID) ([]domain.User, error) {
	var users []domain.User

	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where(`
			id IN (
				SELECT owner_id FROM tasks WHERE assignee_id = ?
				UNION
				SELECT planner_id FROM tasks WHERE assignee_id = ?
			)
		`, id, id).
		Where("id <> ?", id).
		Find(&users).Error

	return users, err
}

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
