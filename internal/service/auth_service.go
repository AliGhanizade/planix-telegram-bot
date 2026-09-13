// Package service implements the business rules.
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// web panel auth constants.
const (
	// CodeTTL is how long a login code stays valid.
	CodeTTL = 10 * time.Minute
	// SessionTTL is the web session lifetime; it slides forward on activity.
	SessionTTL = 30 * 24 * time.Hour
	// SessionExtendWindow renews the session when expiry is closer than this.
	SessionExtendWindow = 7 * 24 * time.Hour
)

// AuthService issues and verifies login codes and manages web sessions.
type AuthService struct {
	db       *gorm.DB
	users    *repository.UserRepository
	codes    *repository.LoginCodeRepository
	sessions *repository.WebSessionRepository
	log      *zap.Logger
}

// NewAuthService builds the auth service.
func NewAuthService(db *gorm.DB, log *zap.Logger) *AuthService {
	return &AuthService{
		db:       db,
		users:    repository.NewUser(db),
		codes:    repository.NewLoginCode(db),
		sessions: repository.NewWebSession(db),
		log:      log,
	}
}

// IssueLoginCode issues a login code and voids previous ones.
// source says whether the code was issued in the bot or requested from the web.
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

// RequestLoginByIdentifier finds a user by telegram username or numeric id
// and issues a code (web initiated). The user must have started the bot.
func (s *AuthService) RequestLoginByIdentifier(ctx context.Context, identifier string) (*domain.User, *domain.LoginCode, error) {
	var user *domain.User
	var err error
	if id, parseErr := strconv.ParseInt(identifier, 10, 64); parseErr == nil {
		user, err = s.users.GetByTelegramID(ctx, id)
	} else {
		user, err = s.users.GetByUsername(ctx, identifier)
	}
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

// VerifyLoginCode verifies the code and issues a 30 day web session.
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

// ValidateSession checks the token, returns the user and extends the session
// close to expiry, extending it by another 30 days so active users stay logged in.
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

// Logout revokes a token.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if err := s.sessions.Delete(ctx, token); err != nil {
		return err
	}
	s.audit(ctx, nil, "logout", uuid.Nil, nil)
	return nil
}

// createSession generates a random 256 bit token and a new session.
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

// audit writes auth events into ActivityLog.
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

// generateCode produces a random 6 digit code.
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
