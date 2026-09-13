package repository

import (
	"context"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoginCodeRepository holds queries for web login codes.
type LoginCodeRepository struct{ db *gorm.DB }

func NewLoginCode(db *gorm.DB) *LoginCodeRepository { return &LoginCodeRepository{db} }

// Create stores a new code and voids the user's previous unused ones.
func (r *LoginCodeRepository) Create(ctx context.Context, c *domain.LoginCode) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", c.UserID).
		Delete(&domain.LoginCode{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(c).Error
}

// FindActive returns a valid unused code.
func (r *LoginCodeRepository) FindActive(ctx context.Context, code string) (*domain.LoginCode, error) {
	var v domain.LoginCode
	err := r.db.WithContext(ctx).
		Where("code = ? AND used_at IS NULL AND expires_at > ?", code, time.Now()).
		First(&v).Error
	return &v, err
}

// MarkUsed marks a code as used.
func (r *LoginCodeRepository) MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.LoginCode{}).Where("id = ?", id).Update("used_at", at).Error
}

// WebSessionRepository holds queries for web sessions.
type WebSessionRepository struct{ db *gorm.DB }

func NewWebSession(db *gorm.DB) *WebSessionRepository { return &WebSessionRepository{db} }

// Create stores a new session.
func (r *WebSessionRepository) Create(ctx context.Context, s *domain.WebSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetByToken returns an active session by token.
func (r *WebSessionRepository) GetByToken(ctx context.Context, token string) (*domain.WebSession, error) {
	var v domain.WebSession
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&v).Error
	return &v, err
}

// Touch updates last activity and extends the session when expiry is near.
// the second return value says whether the session was extended.
func (r *WebSessionRepository) Touch(ctx context.Context, s *domain.WebSession, now time.Time, ttl, extendWindow time.Duration) (bool, error) {
	updates := map[string]any{"last_seen_at": now}
	extended := false
	if s.ExpiresAt.Sub(now) < extendWindow {
		updates["expires_at"] = now.Add(ttl)
		extended = true
	}
	if err := r.db.WithContext(ctx).Model(&domain.WebSession{}).Where("id = ?", s.ID).Updates(updates).Error; err != nil {
		return false, err
	}
	s.LastSeenAt = now
	if extended {
		s.ExpiresAt = now.Add(ttl)
	}
	return extended, nil
}

// Delete revokes a token (logout).
func (r *WebSessionRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&domain.WebSession{}).Error
}
