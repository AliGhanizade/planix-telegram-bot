// Package service قواعد کسب‌وکار برنامه را پیاده‌سازی می‌کند.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ثابت‌های احراز هویت پنل وب.
const (
	// CodeTTL مدت اعتبار کد یک‌بارمصرف.
	CodeTTL = 10 * time.Minute
	// SessionTTL مدت اعتبار نشست وب؛ با فعالیت کاربر به‌صورت لغزان تمدید می‌شود.
	SessionTTL = 30 * 24 * time.Hour
	// SessionExtendWindow اگر تا این مدت به انقضای نشست مانده باشد، تمدید می‌شود.
	SessionExtendWindow = 7 * 24 * time.Hour
)

// AuthService صدور و تایید کدهای ورود و مدیریت نشست‌های پنل وب.
type AuthService struct {
	db       *gorm.DB
	users    *repository.UserRepository
	codes    *repository.LoginCodeRepository
	sessions *repository.WebSessionRepository
	log      *zap.Logger
}

// NewAuthService سرویس احراز هویت را می‌سازد.
func NewAuthService(db *gorm.DB, log *zap.Logger) *AuthService {
	return &AuthService{
		db:       db,
		users:    repository.NewUser(db),
		codes:    repository.NewLoginCode(db),
		sessions: repository.NewWebSession(db),
		log:      log,
	}
}

// IssueLoginCode برای کاربر کد ورود صادر می‌کند و کدهای قبلی او را باطل می‌کند.
// source مشخص می‌کند کد از داخل بات صادر شده یا از سمت پنل وب درخواست شده.
func (s *AuthService) IssueLoginCode(ctx context.Context, userID uuid.UUID, source string) (*domain.LoginCode, error) {
	code, err := generateCode()
	if err != nil {
		return nil, err
	}
	lc := &domain.LoginCode{UserID: userID, Code: code, Source: source, ExpiresAt: time.Now().Add(CodeTTL)}
	if err := s.codes.Create(ctx, lc); err != nil {
		return nil, err
	}
	s.audit(ctx, &userID, "login_code_issued", lc.ID, map[string]string{"source": source})
	return lc, nil
}

// RequestLoginByUsername از سمت پنل وب: با یوزرنیم تلگرام کاربر را پیدا و کد صادر می‌کند.
// کاربر باید قبلاً حداقل یک‌بار با بات استارت کرده باشد.
func (s *AuthService) RequestLoginByUsername(ctx context.Context, username string) (*domain.User, *domain.LoginCode, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found: %w", err)
	}
	if !user.IsActive {
		return nil, nil, fmt.Errorf("user is not active")
	}
	code, err := s.IssueLoginCode(ctx, user.ID, "web")
	if err != nil {
		return nil, nil, err
	}
	return user, code, nil
}

// VerifyLoginCode کد را تایید و نشست وب ۳۰ روزه صادر می‌کند.
func (s *AuthService) VerifyLoginCode(ctx context.Context, code, userAgent string) (*domain.User, *domain.WebSession, error) {
	lc, err := s.codes.FindActive(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid or expired code")
	}
	now := time.Now()
	if err := s.codes.MarkUsed(ctx, lc.ID, now); err != nil {
		return nil, nil, err
	}

	user, err := s.users.GetByID(ctx, lc.UserID)
	if err != nil {
		return nil, nil, err
	}
	session, err := s.createSession(ctx, user.ID, userAgent)
	if err != nil {
		return nil, nil, err
	}
	s.audit(ctx, &user.ID, "login_verified", user.ID, map[string]string{"source": lc.Source})
	return user, session, nil
}

// ValidateSession توکن را بررسی می‌کند؛ کاربر را برمی‌گرداند و در صورت نزدیک بودن
// انقضا، نشست را برای ۳۰ روز دیگر تمدید می‌کند تا کاربر فعال دائمی لاگین نمانَد بیرون.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*domain.User, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid session")
	}
	now := time.Now()
	if _, err := s.sessions.Touch(ctx, session, now, SessionTTL, SessionExtendWindow); err != nil {
		return nil, err
	}
	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user is not active")
	}
	return user, nil
}

// Logout توکن را باطل می‌کند.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if err := s.sessions.Delete(ctx, token); err != nil {
		return err
	}
	s.audit(ctx, nil, "logout", uuid.Nil, nil)
	return nil
}

// createSession توکن تصادفی ۲۵۶ بیتی و نشست جدید می‌سازد.
func (s *AuthService) createSession(ctx context.Context, userID uuid.UUID, userAgent string) (*domain.WebSession, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	session := &domain.WebSession{
		UserID:     userID,
		Token:      hex.EncodeToString(raw),
		ExpiresAt:  time.Now().Add(SessionTTL),
		LastSeenAt: time.Now(),
		UserAgent:  userAgent,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}
	s.audit(ctx, &userID, "session_created", session.ID, nil)
	return session, nil
}

// audit رویدادهای احراز هویت را در جدول ActivityLog ثبت می‌کند.
func (s *AuthService) audit(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, meta map[string]string) {
	raw := "{}"
	if meta != nil {
		if b, err := json.Marshal(meta); err == nil {
			raw = string(b)
		}
	}
	err := s.db.WithContext(ctx).Create(&domain.ActivityLog{
		UserID:     userID,
		EntityType: "auth",
		EntityID:   entityID,
		Action:     action,
		Metadata:   raw,
		OccurredAt: time.Now(),
	}).Error
	if err != nil {
		s.log.Warn("auth audit failed", zap.Error(err), zap.String("action", action))
	}
}

// generateCode کد ۶ رقمی تصادفی امن تولید می‌کند.
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
