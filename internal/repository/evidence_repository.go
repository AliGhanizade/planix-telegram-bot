package repository

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EvidenceRepository holds queries for task evidence (photo proofs).
type EvidenceRepository struct{ db *gorm.DB }

func NewEvidence(db *gorm.DB) *EvidenceRepository { return &EvidenceRepository{db} }

// Create stores a new evidence record.
func (r *EvidenceRepository) Create(ctx context.Context, e *domain.TaskEvidence) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// LatestForTask returns the newest evidence of a task.
func (r *EvidenceRepository) LatestForTask(ctx context.Context, taskID uuid.UUID) (*domain.TaskEvidence, error) {
	var v domain.TaskEvidence
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("created_at desc").
		First(&v).Error
	return &v, err
}
