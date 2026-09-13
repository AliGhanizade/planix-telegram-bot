// Package bot پیاده‌سازی بات تلگرام پلنیکس است: دریافت آپدیت‌ها،
// کیبوردهای رنگی، جریان‌های چندمرحله‌ای و رابط کاربری همگام.
package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Bot نگهدارنده‌ی کلاینت تلگرام و وابستگی‌های بات است.
type Bot struct {
	api   *tgbot.Bot
	users *repository.UserRepository
	tasks *service.TaskService
	db    *gorm.DB
	log   *zap.Logger
	owner string
	me    *models.User
}

// New کلاینت تلگرام را می‌سازد و هندلرها را ثبت می‌کند.
func New(token, owner string, db *gorm.DB, log *zap.Logger) (*Bot, error) {
	var b *Bot
	api, err := tgbot.New(token,
		tgbot.WithAllowedUpdates(tgbot.AllowedUpdates{"message", "callback_query"}),
		tgbot.WithErrorsHandler(func(err error) {
			log.Error("telegram api error", zap.Error(err))
		}),
		tgbot.WithDefaultHandler(func(ctx context.Context, tb *tgbot.Bot, u *models.Update) {
			b.onUpdate(ctx, u)
		}),
	)
	if err != nil {
		return nil, err
	}
	b = &Bot{api: api, users: repository.NewUser(db), tasks: service.NewTask(db), db: db, log: log, owner: owner}
	return b, nil
}

// Run فهرست دستورهای بات را ثبت و حلقه‌ی دریافت آپدیت‌ها را شروع می‌کند.
// تا لغو شدن کانتکست بلاک می‌ماند.
func (b *Bot) Run(ctx context.Context) {
	me, err := b.api.GetMe(ctx)
	if err != nil {
		b.log.Warn("get me failed", zap.Error(err))
	} else {
		b.me = me
		b.log.Info("bot started", zap.String("username", me.Username))
	}

	if _, err := b.api.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "start", Description: "شروع کار با پلنیکس"},
			{Command: "today", Description: "برنامه امروز"},
			{Command: "new", Description: "ثبت تسک جدید"},
			{Command: "search", Description: "جستجو در تسک‌ها"},
			{Command: "help", Description: "راهنما"},
		},
	}); err != nil {
		b.log.Warn("set my commands failed", zap.Error(err))
	}

	b.api.Start(ctx)
}

// onUpdate هر آپدیت را در برابر پنیک محافظت و به هندلر مناسب می‌سپارد.
func (b *Bot) onUpdate(ctx context.Context, u *models.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("panic while handling update", zap.Any("panic", r))
		}
	}()

	b.log.Debug("telegram update received", zap.Int64("update_id", u.ID))
	switch {
	case u.Message != nil && u.Message.Text != "":
		b.onText(ctx, u.Message)
	case u.CallbackQuery != nil:
		b.onCallback(ctx, u.CallbackQuery)
	}
}

// ProcessUpdate یک آپدیت خام (مثلاً از وب‌هوک) را پردازش می‌کند؛ پردازش ناهمگام است.
func (b *Bot) ProcessUpdate(ctx context.Context, u *models.Update) {
	go b.onUpdate(ctx, u)
}

// ---------- helper های ارسال و دریافت ----------

// send یک پیام با markup دلخواه می‌فرستد.
func (b *Bot) send(ctx context.Context, chatID int64, text string, markup models.ReplyMarkup) (*models.Message, error) {
	return b.api.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
}

// sendWithKeyboard پیام با کیبورد اصلی (رنگی) می‌فرستد.
func (b *Bot) sendWithKeyboard(ctx context.Context, chatID int64, text string) (*models.Message, error) {
	return b.send(ctx, chatID, text, MainKeyboard())
}

// edit متن و markup یک پیام موجود را همان‌جا بروزرسانی می‌کند تا رابط کاربری سینک بماند.
func (b *Bot) edit(ctx context.Context, chatID int64, messageID int, text string, markup models.ReplyMarkup) error {
	_, err := b.api.EditMessageText(ctx, &tgbot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ReplyMarkup: markup,
	})
	if err != nil && strings.Contains(err.Error(), "message is not modified") {
		return nil
	}
	return err
}

// render یا پیام موجود را ویرایش می‌کند یا پیام تازه می‌فرستد.
func (b *Bot) render(ctx context.Context, chatID int64, messageID int, text string, markup models.ReplyMarkup) error {
	if messageID > 0 {
		return b.edit(ctx, chatID, messageID, text, markup)
	}
	_, err := b.send(ctx, chatID, text, markup)
	return err
}

// answer روی دکمه‌ی فشرده‌شده یک اعلان کوتاه نشان می‌دهد.
func (b *Bot) answer(ctx context.Context, q *models.CallbackQuery, text string) {
	_, _ = b.api.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: q.ID,
		Text:            text,
	})
}

// answerAlert روی دکمه یک اعلان هشداری نشان می‌دهد.
func (b *Bot) answerAlert(ctx context.Context, q *models.CallbackQuery, text string) {
	_, _ = b.api.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: q.ID,
		Text:            text,
		ShowAlert:       true,
	})
}

// upsertUser کاربر تلگرام را در دیتابیس ثبت یا بروزرسانی می‌کند.
func (b *Bot) upsertUser(ctx context.Context, from models.User) (*domain.User, error) {
	u := &domain.User{
		TelegramID:   from.ID,
		Username:     from.Username,
		FirstName:    from.FirstName,
		LastName:     from.LastName,
		LanguageCode: from.LanguageCode,
		LastSeenAt:   ptr(time.Now()),
	}
	if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// cbOrigin چت و پیامِ مبدأ یک کال‌بک را برمی‌گرداند.
func cbOrigin(q *models.CallbackQuery) (chatID int64, messageID int, ok bool) {
	if q.Message.Message != nil {
		return q.Message.Message.Chat.ID, q.Message.Message.ID, true
	}
	if q.Message.InaccessibleMessage != nil {
		return q.Message.InaccessibleMessage.Chat.ID, q.Message.InaccessibleMessage.MessageID, false
	}
	return 0, 0, false
}

// onText پیام‌های متنی کاربر را پردازش می‌کند.
func (b *Bot) onText(ctx context.Context, m *models.Message) {
	if m.From == nil || m.From.IsBot {
		return
	}
	u, err := b.upsertUser(ctx, *m.From)
	if err != nil {
		b.log.Error("upsert user failed", zap.Error(err))
		return
	}

	text := strings.TrimSpace(m.Text)
	switch text {
	case "/start":
		b.sendWithKeyboard(ctx, m.Chat.ID, welcomeMessage(b.me))
	case "/today", "📋 برنامه امروز", "📅 برنامه‌های من":
		if err := b.renderTaskList(ctx, m.Chat.ID, 0, u.ID, filterPending, 1); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}
	case "/new", "➕ تسک جدید":
		b.startNewTask(ctx, u, m.Chat.ID, 0)
	case "/search", "🔍 جستجو":
		b.startSearch(ctx, u, m.Chat.ID, 0)
	case "/help", "ℹ️ راهنما":
		b.sendWithKeyboard(ctx, m.Chat.ID, helpMessage)
	case "/profile", "👤 پروفایل":
		b.sendProfile(ctx, u, m.Chat.ID, 0)
	case "👥 واگذاری تسک", "👥 اعمال وظایف دیگران":
		b.startAssign(ctx, u, m.Chat.ID, 0)
	case "📊 وضعیت وظایف دیگران", "✅ وضعیت وظایف دیگران":
		b.sendStatusPick(ctx, u, m.Chat.ID, 0)
	case "⚙️ تنظیمات":
		b.showSettings(ctx, u, m.Chat.ID, 0)
	case "🛟 پشتیبانی", "پشتیبانی":
		b.sendWithKeyboard(ctx, m.Chat.ID, fmt.Sprintf("🛟 برای ارتباط با پشتیبانی به @%s پیام بده.", b.owner))
	case "❌ بازگشت", " ❌ بازگشت":
		_ = b.clearSession(ctx, u.ID)
		b.sendWithKeyboard(ctx, m.Chat.ID, menuText)
	default:
		if err := b.checkState(ctx, u, text, m.Chat.ID); err != nil {
			b.log.Error("handle state failed", zap.Error(err))
		}
	}
}

func ptr(t time.Time) *time.Time { return &t }
