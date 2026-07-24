package bot

import (
	"context"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func (b *Bot) handleMessage(ctx context.Context, m *tgbotapi.Message) error {
	helpMsg := "سلام، من پلنیکس هستم ✨\n\n" +
		"➕ ** ثبت تسک جدید: ** ثبت سریع تسک برای خودتان\n" +
		"📋 ** برنامه امروز:** مشاهده و مدیریت کارهای امروز شما\n" +
		"👥 **اعمال وظایف دیگران:** واگذاری چند تسک به یک فرد دیگر\n" +
		"✅ **وضعیت وظایف دیگران:** مشاهده تسک‌هایی که به دیگران محول کرده‌اید"
	statusOtherMsg := "برای چه کسی میخوای وضعیت تسک‌هاشو ببینی؟\n" +
		"لطفا یوزرنیمش را بفرست. مثال:\n@ali\nUsername"
	promptMsg := "📋 *فرمت ارسال*\n\n" +
		"خط اول: یوزرنیم\n" +
		"بقیه: هر تسک در یک خط\n\n" +
		"`@username\n" +
		"طراحی API\n" +
		"بررسی Pull Request`"

	u := &domain.User{TelegramID: m.From.ID, Username: m.From.UserName, FirstName: m.From.FirstName, LastName: m.From.LastName, LanguageCode: m.From.LanguageCode, LastSeenAt: ptr(time.Now())}
	if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
		return err
	}
	b.log.Debug("bot in tel os work msg", zap.String("text", m.Text), zap.String("username", m.From.UserName), zap.String("first_name", m.From.FirstName), zap.String("last_name", m.From.LastName))
	text := strings.TrimSpace(m.Text)
	switch text {
	case "/start", "ℹ️ راهنما":
		return b.reply(m.Chat.ID, helpMsg, MainKeyboard())
	case "➕ تسک جدید":
		return b.setStateAndReply(ctx, u.ID, "waiting_task_title", m.Chat.ID, "عنوان تسک را بفرست. مثال: مطالعه گولنگ", CancelStateInlineKeyboard())
	case "📋 برنامه امروز", "📅 برنامه‌های من":
		return b.sendToday(ctx, u, m.Chat.ID)
	case "👥 اعمال وظایف دیگران":
		return b.setStateAndReply(ctx, u.ID, "waiting_task_for_other", m.Chat.ID, promptMsg, CancelStateInlineKeyboard())
	case "✅ وضعیت وظایف دیگران":
		b.setState(ctx, u.ID, "waiting_task_status_for_other", m.Chat.ID, nil)
		
		users, err := b.users.ListUsersAssignedTo(ctx, u.ID)
		if err != nil {
			return err
		}
		return b.reply(m.Chat.ID, statusOtherMsg, SuggestFriendInlineKeyboard(users))
	case " ❌ بازگشت":
		return b.deleteState(ctx, u.ID)
	default:
		err := b.checkState(ctx, u, text)
		if err != nil {
			return err
		}
		return nil
	}
}

func ptr(t time.Time) *time.Time { return &t }
