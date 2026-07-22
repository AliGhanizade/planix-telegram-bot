package bot

import (
	"context"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleMessage(ctx context.Context, m *tgbotapi.Message) error {
	helpMsg := "سلام، من پلنیکس هستم ✨\n\n" +
		"➕ تسک جدید — ثبت سریع تسک برای خودت\n" +
		"ℹ️ راهنما — همین پیام"

	u := &domain.User{TelegramID: m.From.ID, Username: m.From.UserName, FirstName: m.From.FirstName, LastName: m.From.LastName, LanguageCode: m.From.LanguageCode, LastSeenAt: ptr(time.Now())}
	if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
		return err
	}

	text := strings.TrimSpace(m.Text)
	switch text {
	case "/start", "ℹ️ راهنما":
		return b.reply(m.Chat.ID, helpMsg, MainKeyboard())
	case "➕ تسک جدید":
		return b.setStateAndReply(ctx, u.ID, "waiting_task_title", m.Chat.ID, "عنوان تسک را بفرست. مثال: مطالعه گولنگ", CancelStateInlineKeyboard())
	default:
		return b.checkState(ctx, u, text)
	}
}

func ptr(t time.Time) *time.Time { return &t }
