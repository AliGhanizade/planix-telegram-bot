package bot

import (
	"context"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
)

func (b *Bot) checkState(ctx context.Context, u *domain.User, text string) error {
	var session domain.BotSession
	if err := b.db.WithContext(ctx).Where("user_id = ? AND state = ? AND expires_at > ?", u.ID, "waiting_task_title", time.Now()).First(&session).Error; err != nil {
		return nil
	}
	task := &domain.Task{OwnerID: u.ID, AssigneeID: u.ID, Title: text, Priority: "normal", Status: "pending"}
	if err := b.tasks.Create(ctx, task); err != nil {
		return err
	}
	b.db.WithContext(ctx).Delete(&session)
	return b.reply(u.TelegramID, "تسک ثبت شد ✅\nهر زمان انجامش دادی از لیست امروز تیکش بزن.", MainKeyboard())
}

func (b *Bot) deleteState(ctx context.Context, userID uuid.UUID) error {
	return b.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.BotSession{}).Error
}
