package bot

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) sendToday(ctx context.Context, u *domain.User, chatID int64) error {
	tasks, err := b.tasks.Today(ctx, u.ID)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return b.reply(chatID, "امروز تسک بازی نداری؛ وقت یک شروع تازه است ✨", MainKeyboard())
	}
	text := "📋 تسک‌های باز تو:\n"
	text += FormatTasks(tasks)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = TasksInlineKeyboard(tasks)
	_, err = b.api.Send(msg)
	return err
}
