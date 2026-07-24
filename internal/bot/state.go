package bot

import (
	"context"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
)

func (b *Bot) checkState(ctx context.Context, u *domain.User, text string) error {
	var session domain.BotSession

	switch {
	case b.db.WithContext(ctx).Where("user_id = ? AND state = ? AND expires_at > ?", u.ID, "waiting_task_title", time.Now()).First(&session).Error == nil:
		task := &domain.Task{OwnerID: u.ID, AssigneeID: u.ID, Title: text, Priority: "normal", Status: "pending"}
		if err := b.tasks.Create(ctx, task); err != nil {
			return err
		}
		b.db.WithContext(ctx).Delete(&session)
		return b.reply(u.TelegramID, "تسک ثبت شد ✅\nهر زمان انجامش دادی از لیست امروز تیکش بزن.", MainKeyboard())

	case b.db.WithContext(ctx).Where("user_id = ? AND state = ? AND expires_at > ?", u.ID, "waiting_task_for_other", time.Now()).First(&session).Error == nil:
		lines := strings.Split(text, "\n")
		if len(lines) < 2 {
			return b.reply(u.TelegramID, "لطفا یوزرنیم را در خط اول و هر تسک را در یک خط جدا بفرست.", CancelStateInlineKeyboard())
		}

		username := strings.TrimPrefix(strings.TrimSpace(lines[0]), "@")
		targetUser, err := b.users.GetByUsername(ctx, username)
		if err != nil {
			return b.reply(u.TelegramID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", CancelStateInlineKeyboard())
		}
		if targetUser.ID == u.ID {
			return b.reply(u.TelegramID, "نمی‌تونی تسک رو به خودت واگذار کنی 🙂", CancelStateInlineKeyboard())
		}

		if err := b.setTaskForOther(ctx, u.ID, targetUser.ID, lines[1:], "normal"); err != nil {
			return err
		}
		return b.deleteState(ctx, u.ID)

	case b.db.WithContext(ctx).Where("user_id = ? AND state = ? AND expires_at > ?", u.ID, "waiting_task_status_for_other", time.Now()).First(&session).Error == nil:
		username := strings.TrimPrefix(text, "@")

		targetUser, err := b.users.GetByUsername(ctx, username)
		if err != nil {
			return b.reply(u.TelegramID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", CancelStateInlineKeyboard())
		}

		b.db.WithContext(ctx).Delete(&session)
		return b.getTargetTask(ctx, u.ID, targetUser.ID)
	default:
		return nil
	}
}

func (b *Bot) deleteState(ctx context.Context, userID uuid.UUID) error {
	return b.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.BotSession{}).Error
}
