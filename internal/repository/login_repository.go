package repository

import (
	"context"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoginCodeRepository کوئری‌های کدهای یک‌بارمصرف ورود وب.
type LoginCodeRepository struct{ db *gorm.DB }

func NewLoginCode(db *gorm.DB) *LoginCodeRepository { return &LoginCodeRepository{db} }

// Create کد جدید ذخیره می‌کند و کدهای قبلیِ مصرف‌نشده‌ی همان کاربر را باطل می‌کند.
func (r *LoginCodeRepository) Create(ctx context.Context, c *domain.LoginCode) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", c.UserID).
		Delete(&domain.LoginCode{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(c).Error
}

// FindActive کدِ معتبر و مصرف‌نشده را برمی‌گرداند.
func (r *LoginCodeRepository) FindActive(ctx context.Context, code string) (*domain.LoginCode, error) {
	var v domain.LoginCode
	err := r.db.WithContext(ctx).
		Where("code = ? AND used_at IS NULL AND expires_at > ?", code, time.Now()).
		First(&v).Error
	return &v, err
}

// MarkUsed کد را مصرف‌شده علامت می‌زند.
func (r *LoginCodeRepository) MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.LoginCode{}).Where("id = ?", id).Update("used_at", at).Error
}

// WebSessionRepository کوئری‌های نشست‌های پنل وب.
type WebSessionRepository struct{ db *gorm.DB }

func NewWebSession(db *gorm.DB) *WebSessionRepository { return &WebSessionRepository{db} }

// Create نشست جدید ذخیره می‌کند.
func (r *WebSessionRepository) Create(ctx context.Context, s *domain.WebSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetByToken نشست فعال را با توکن برمی‌گرداند.
func (r *WebSessionRepository) GetByToken(ctx context.Context, token string) (*domain.WebSession, error) {
	var v domain.WebSession
	err := r.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&v).Error
	return &v, err
}

// Touch آخرین فعالیت نشست را ثبت و در صورت نزدیک بودن انقضا، تمدید می‌کند.
// خروجی دوم یعنی نشست تمدید شده است.
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

// Delete توکن را باطل می‌کند (خروج از حساب).
func (r *WebSessionRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&domain.WebSession{}).Error
}
