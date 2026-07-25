package bot

import (
	"context"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// assignPrompt فرمت پیام واگذاری تسک را به کاربر توضیح می‌دهد.
const assignPrompt = "📋 فرمت ارسال\n\n" +
	"خط اول: یوزرنیم\n" +
	"بقیه: هر تسک در یک خط\n\n" +
	"@username\n" +
	"طراحی API\n" +
	"بررسی Pull Request"

// handleMessage پیام‌های متنی کاربر را پردازش می‌کند.
func (b *Bot) handleMessage(ctx context.Context, m *tgbotapi.Message) error {
	helpMsg := "سلام، من پلنیکس هستم ✨\n\n" +
		"➕ تسک جدید — ثبت سریع تسک برای خودت\n" +
		"📋 برنامه امروز — مشاهده و مدیریت کارهای امروز\n" +
		"👥 اعمال وظایف دیگران — واگذاری چند تسک به یک نفر\n" +
		"✅ وضعیت وظایف دیگران — پیگیری تسک‌هایی که محول کرده‌ای\n" +
		"ℹ️ راهنما — همین پیام"

	u := &domain.User{TelegramID: m.From.ID, Username: m.From.UserName, FirstName: m.From.FirstName, LastName: m.From.LastName, LanguageCode: m.From.LanguageCode, LastSeenAt: ptr(time.Now())}
	if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
		return err
	}
	b.log.Debug("message received", zap.String("text", m.Text), zap.String("username", m.From.UserName))

	text := strings.TrimSpace(m.Text)
	switch text {
	case "/start", "ℹ️ راهنما":
		return b.reply(m.Chat.ID, helpMsg, MainKeyboard())
	case "➕ تسک جدید":
		return b.setStateAndReply(ctx, u.ID, stateWaitingTaskTitle, "{}", m.Chat.ID, "عنوان تسک را بفرست. مثال: مطالعه گولنگ", CancelStateInlineKeyboard())
	case "📋 برنامه امروز", "📅 برنامه‌های من":
		return b.sendToday(ctx, u, m.Chat.ID)
	case "👥 اعمال وظایف دیگران":
		return b.setStateAndReply(ctx, u.ID, stateWaitingTaskForOther, "{}", m.Chat.ID, assignPrompt, CancelStateInlineKeyboard())
	case "✅ وضعیت وظایف دیگران":
		if err := b.setState(ctx, u.ID, stateWaitingStatusForOther, "{}"); err != nil {
			return err
		}
		users, err := b.users.ListUsersAssignedTo(ctx, u.ID)
		if err != nil {
			return err
		}
		return b.reply(m.Chat.ID, "برای چه کسی می‌خوای وضعیت تسک‌هاشو ببینی؟", SuggestFriendInlineKeyboard(users))
	case "❌ بازگشت", " ❌ بازگشت":
		return b.deleteState(ctx, u.ID)
	default:
		return b.checkState(ctx, u, text)
	}
}

func ptr(t time.Time) *time.Time { return &t }
