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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// maxNameLen حداکثر طول مجاز نام و نام خانوادگی.
const maxNameLen = 64

// UserService پروفایل و تنظیمات کاربران.
type UserService struct {
	db    *gorm.DB
	users *repository.UserRepository
	tasks *repository.TaskRepository
	log   *zap.Logger
}

// NewUserService سرویس کاربر را می‌سازد.
func NewUserService(db *gorm.DB, log *zap.Logger) *UserService {
	return &UserService{
		db:    db,
		users: repository.NewUser(db),
		tasks: repository.NewTask(db),
		log:   log,
	}
}

// Profile اطلاعات و آمار کاربر را برمی‌گرداند.
func (s *UserService) Profile(ctx context.Context, userID uuid.UUID) (*domain.User, *TaskStats, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	st := &TaskStats{}
	if st.Pending, err = s.tasks.CountByAssigneeAndStatus(ctx, userID, "pending"); err != nil {
		return nil, nil, err
	}
	if st.Completed, err = s.tasks.CountByAssigneeAndStatus(ctx, userID, "completed"); err != nil {
		return nil, nil, err
	}
	if st.Cancelled, err = s.tasks.CountByAssigneeAndStatus(ctx, userID, "cancelled"); err != nil {
		return nil, nil, err
	}
	return u, st, nil
}

// UpdateProfile فیلدهای ارسالی را بروزرسانی می‌کند؛ nil یعنی تغییر نکن.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, firstName, lastName, timezone *string) (*domain.User, error) {
	updates := map[string]any{}

	if firstName != nil {
		name := strings.TrimSpace(*firstName)
		if name == "" {
			return nil, fmt.Errorf("first name cannot be empty")
		}
		if len([]rune(name)) > maxNameLen {
			return nil, fmt.Errorf("first name is too long")
		}
		updates["first_name"] = name
	}
	if lastName != nil {
		name := strings.TrimSpace(*lastName)
		if len([]rune(name)) > maxNameLen {
			return nil, fmt.Errorf("last name is too long")
		}
		updates["last_name"] = name
	}
	if timezone != nil {
		tz := strings.TrimSpace(*timezone)
		if _, err := time.LoadLocation(tz); err != nil {
			return nil, fmt.Errorf("invalid timezone: %s", tz)
		}
		updates["timezone"] = tz
	}

	if len(updates) == 0 {
		return s.users.GetByID(ctx, userID)
	}
	if err := s.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return nil, err
	}

	changed := make([]string, 0, len(updates))
	for k := range updates {
		changed = append(changed, k)
	}
	s.audit(ctx, userID, "profile_updated", changed)
	return s.users.GetByID(ctx, userID)
}

// SetDailyReport خاموش/روشن کردن گزارش روزانه.
func (s *UserService) SetDailyReport(ctx context.Context, userID uuid.UUID, enabled bool) error {
	if err := s.users.SetDailyReport(ctx, userID, enabled); err != nil {
		return err
	}
	s.audit(ctx, userID, "daily_report_changed", []string{strconvBool(enabled)})
	return nil
}

// audit تغییرات پروفایل را در ActivityLog ثبت می‌کند.
func (s *UserService) audit(ctx context.Context, userID uuid.UUID, action string, changed []string) {
	raw := "{}"
	if b, err := json.Marshal(changed); err == nil {
		raw = string(b)
	}
	err := s.db.WithContext(ctx).Create(&domain.ActivityLog{
		UserID:     &userID,
		EntityType: "user",
		EntityID:   userID,
		Action:     action,
		Metadata:   raw,
		OccurredAt: time.Now(),
	}).Error
	if err != nil {
		s.log.Warn("user audit failed", zap.Error(err), zap.String("action", action))
	}
}

// strconvBool بولین را به رشته‌ی کوتاه برای متادیتا تبدیل می‌کند.
func strconvBool(b bool) string {
	if b {
		return "on"
	}
	return "off"
}
