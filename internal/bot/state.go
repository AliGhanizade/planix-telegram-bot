package bot

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// نام وضعیت‌های مکالمه‌ی چندمرحله‌ای بات.
const (
	stateWaitingTaskTitle       = "waiting_task_title"
	stateWaitingTaskDescription = "waiting_task_description"
	stateWaitingTaskForOther    = "waiting_task_for_other"
	stateWaitingStatusForOther  = "waiting_task_status_for_other"
)

// taskPayload داده‌ی JSON ذخیره‌شده در نشست برای ویرایش یک تسک است.
type taskPayload struct {
	TaskID string `json:"task_id"`
}

// checkState پیام آزاد کاربر را بر اساس وضعیت جاری مکالمه پردازش می‌کند.
func (b *Bot) checkState(ctx context.Context, u *domain.User, text string) error {
	session, ok, err := b.findActiveSession(ctx, u.ID)
	if err != nil || !ok {
		return err
	}

	switch session.State {
	case stateWaitingTaskTitle:
		return b.handleTitleInput(ctx, u, session, text)
	case stateWaitingTaskDescription:
		return b.handleDescriptionInput(ctx, u, session, text)
	case stateWaitingTaskForOther:
		return b.handleAssignInput(ctx, u, session, text)
	case stateWaitingStatusForOther:
		return b.handleStatusInput(ctx, u, text)
	}
	return nil
}

// findActiveSession آخرین نشست معتبر کاربر را برمی‌گرداند.
func (b *Bot) findActiveSession(ctx context.Context, userID uuid.UUID) (domain.BotSession, bool, error) {
	var session domain.BotSession
	err := b.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("updated_at DESC").
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.BotSession{}, false, nil
	}
	if err != nil {
		return domain.BotSession{}, false, err
	}
	return session, true, nil
}

// handleTitleInput عنوان جدید می‌سازد یا عنوان تسک موجودی را ویرایش می‌کند.
func (b *Bot) handleTitleInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	var payload taskPayload
	if err := json.Unmarshal([]byte(session.Data), &payload); err == nil && payload.TaskID != "" {
		id, err := uuid.Parse(payload.TaskID)
		if err != nil {
			return err
		}
		if err := b.tasks.UpdateTitle(ctx, id, text); err != nil {
			return err
		}
		if err := b.deleteState(ctx, u.ID); err != nil {
			return err
		}
		return b.reply(u.TelegramID, "✏️ عنوان تسک بروزرسانی شد.", MainKeyboard())
	}

	if err := b.deleteState(ctx, u.ID); err != nil {
		return err
	}
	task := &domain.Task{OwnerID: u.ID, AssigneeID: u.ID, Title: text, Priority: "normal", Status: "pending"}
	if err := b.tasks.Create(ctx, task); err != nil {
		return err
	}
	return b.reply(u.TelegramID, "تسک ثبت شد ✅\nهر زمان انجامش دادی از لیست امروز تیکش بزن.", MainKeyboard())
}

// handleDescriptionInput توضیحات تسکِ در حال ویرایش را ذخیره می‌کند.
func (b *Bot) handleDescriptionInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	var payload taskPayload
	if err := json.Unmarshal([]byte(session.Data), &payload); err != nil || payload.TaskID == "" {
		return b.deleteState(ctx, u.ID)
	}
	id, err := uuid.Parse(payload.TaskID)
	if err != nil {
		return err
	}
	if err := b.tasks.UpdateDescription(ctx, id, text); err != nil {
		return err
	}
	if err := b.deleteState(ctx, u.ID); err != nil {
		return err
	}
	return b.reply(u.TelegramID, "📄 توضیحات تسک بروزرسانی شد.", MainKeyboard())
}

// handleAssignInput ورودی «واگذاری تسک» را تجزیه و تسک‌ها را ثبت می‌کند.
func (b *Bot) handleAssignInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return b.reply(u.TelegramID, "لطفا یوزرنیم را در خط اول و هر تسک را در یک خط جدا بفرست.", CancelStateInlineKeyboard())
	}
	username := strings.TrimPrefix(strings.TrimSpace(lines[0]), "@")
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		return b.reply(u.TelegramID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", CancelStateInlineKeyboard())
	}
	if target.ID == u.ID {
		return b.reply(u.TelegramID, "نمی‌تونی تسک رو به خودت واگذار کنی 🙂", CancelStateInlineKeyboard())
	}
	if err := b.setTaskForOther(ctx, u.ID, target.ID, lines[1:], "normal"); err != nil {
		return err
	}
	return b.deleteState(ctx, u.ID)
}

// handleStatusInput وضعیت تسک‌های واگذارشده به کاربر هدف را نشان می‌دهد.
func (b *Bot) handleStatusInput(ctx context.Context, u *domain.User, text string) error {
	username := strings.TrimPrefix(strings.TrimSpace(text), "@")
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		return b.reply(u.TelegramID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", CancelStateInlineKeyboard())
	}
	if err := b.deleteState(ctx, u.ID); err != nil {
		return err
	}
	return b.getTargetTask(ctx, u.ID, target.ID)
}

// deleteState نشست فعال کاربر را حذف می‌کند.
func (b *Bot) deleteState(ctx context.Context, userID uuid.UUID) error {
	return b.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.BotSession{}).Error
}
