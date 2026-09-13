// Package bot implements the planix telegram bot: update handling,
// colored keyboards, multi step flows and a synced user interface.
package bot

import (
	"context"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Bot holds the telegram client and bot dependencies.
type Bot struct {
	api      *tgbot.Bot
	users    *repository.UserRepository
	tasks    *service.TaskService
	auth     *service.AuthService
	profiles *service.UserService
	db       *gorm.DB
	log      *zap.Logger
	owner    string
	me       *models.User
}

// New creates the telegram client and registers handlers; auth
// and profile services are injected so bot and web share one service layer.
func New(token, owner string, db *gorm.DB, log *zap.Logger, auth *service.AuthService, profiles *service.UserService) (*Bot, error) {
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
	b = &Bot{api: api, users: repository.NewUser(db), tasks: service.NewTask(db), auth: auth, profiles: profiles, db: db, log: log, owner: owner}
	return b, nil
}

// Tasks exposes the task service (shared with httpapi).
func (b *Bot) Tasks() *service.TaskService { return b.tasks }

// Run registers bot commands and starts the update loop.
// blocks until the context is cancelled.
func (b *Bot) Run(ctx context.Context) {
	me, err := b.api.GetMe(ctx)
	if err != nil {
		b.log.Warn("get me failed", zap.Error(err))
	} else {
		b.me = me
		b.log.Info("bot started", zap.String("username", me.Username))
	}

	faCommands := []models.BotCommand{
		{Command: "start", Description: "شروع کار با پلنیکس"},
		{Command: "today", Description: "برنامه امروز"},
		{Command: "new", Description: "ثبت تسک جدید"},
		{Command: "search", Description: "جستجو در تسک‌ها"},
		{Command: "help", Description: "راهنما"},
	}
	enCommands := []models.BotCommand{
		{Command: "start", Description: "Get started with Planix"},
		{Command: "today", Description: "Today's plan"},
		{Command: "new", Description: "Create a new task"},
		{Command: "search", Description: "Search your tasks"},
		{Command: "help", Description: "Help"},
	}
	if _, err := b.api.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{Commands: faCommands}); err != nil {
		b.log.Warn("set my commands failed", zap.Error(err))
	}
	if _, err := b.api.SetMyCommands(ctx, &tgbot.SetMyCommandsParams{Commands: enCommands, LanguageCode: "en"}); err != nil {
		b.log.Warn("set english commands failed", zap.Error(err))
	}

	b.api.Start(ctx)
}

// onUpdate guards against panics and routes each update.
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

// ProcessUpdate handles a raw update (e.g. from the webhook) asynchronously.
func (b *Bot) ProcessUpdate(ctx context.Context, u *models.Update) {
	go b.onUpdate(ctx, u)
}

// ---------- send helpers ----------

// send sends a message with the given markup.
func (b *Bot) send(ctx context.Context, chatID int64, text string, markup models.ReplyMarkup) (*models.Message, error) {
	return b.api.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
}

// sendWithKeyboard sends a message with the main colored keyboard.
func (b *Bot) sendWithKeyboard(ctx context.Context, chatID int64, text string, l ui.Lang) (*models.Message, error) {
	return b.send(ctx, chatID, text, ui.MainKeyboard(l))
}

// lang returns the display language of a user.
func (b *Bot) lang(u *domain.User) ui.Lang { return ui.Normalize(u.Lang) }

// edit updates an existing message in place to keep the ui in sync.
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

// render edits the existing message or sends a fresh one.
func (b *Bot) render(ctx context.Context, chatID int64, messageID int, text string, markup models.ReplyMarkup) error {
	if messageID > 0 {
		return b.edit(ctx, chatID, messageID, text, markup)
	}
	_, err := b.send(ctx, chatID, text, markup)
	return err
}

// answer shows a short toast on the pressed button.
func (b *Bot) answer(ctx context.Context, q *models.CallbackQuery, text string) {
	_, _ = b.api.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: q.ID,
		Text:            text,
	})
}

// answerAlert shows an alert style toast on the button.
func (b *Bot) answerAlert(ctx context.Context, q *models.CallbackQuery, text string) {
	_, _ = b.api.AnswerCallbackQuery(ctx, &tgbot.AnswerCallbackQueryParams{
		CallbackQueryID: q.ID,
		Text:            text,
		ShowAlert:       true,
	})
}

// upsertUser creates or updates the telegram user in the database.
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

// cbOrigin returns the chat and message a callback came from.
func cbOrigin(q *models.CallbackQuery) (chatID int64, messageID int, ok bool) {
	if q.Message.Message != nil {
		return q.Message.Message.Chat.ID, q.Message.Message.ID, true
	}
	if q.Message.InaccessibleMessage != nil {
		return q.Message.InaccessibleMessage.Chat.ID, q.Message.InaccessibleMessage.MessageID, false
	}
	return 0, 0, false
}

// onText processes user text messages.
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
	l := b.lang(u)
	switch text {
	case "/start":
		b.sendWithKeyboard(ctx, m.Chat.ID, ui.WelcomeMessage(b.me, l), l)
	case "/today", "📋 برنامه امروز", "📅 برنامه‌های من", "📋 Today":
		if err := b.renderTaskList(ctx, m.Chat.ID, 0, u.ID, ui.FilterPending, 1, l); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}
	case "/new", "➕ تسک جدید", "➕ New task":
		b.startNewTask(ctx, u, m.Chat.ID, 0)
	case "/search", "🔍 جستجو", "🔍 Search":
		b.startSearch(ctx, u, m.Chat.ID, 0)
	case "/help", "ℹ️ راهنما", "ℹ️ Help":
		b.sendWithKeyboard(ctx, m.Chat.ID, ui.HelpMessage(l), l)
	case "/profile", "👤 پروفایل", "👤 Profile":
		b.sendProfile(ctx, u, m.Chat.ID, 0)
	case "👥 واگذاری تسک", "👥 اعمال وظایف دیگران", "👥 Delegate":
		b.startAssign(ctx, u, m.Chat.ID, 0)
	case "📊 وضعیت وظایف دیگران", "✅ وضعیت وظایف دیگران", "📊 Delegated status":
		b.sendStatusPick(ctx, u, m.Chat.ID, 0)
	case "⚙️ تنظیمات", "⚙️ Settings":
		b.showSettings(ctx, u, m.Chat.ID, 0)
	case "🛟 پشتیبانی", "پشتیبانی", "🛟 Support":
		b.sendWithKeyboard(ctx, m.Chat.ID, ui.SupportMessage(b.owner, l), l)
	case "❌ بازگشت", " ❌ بازگشت", "❌ Back":
		_ = b.clearSession(ctx, u.ID)
		b.sendWithKeyboard(ctx, m.Chat.ID, ui.MenuText(l), l)
	default:
		if err := b.checkState(ctx, u, text, m.Chat.ID); err != nil {
			b.log.Error("handle state failed", zap.Error(err))
		}
	}
}

func ptr(t time.Time) *time.Time { return &t }
