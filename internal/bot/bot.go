// Package bot پیاده‌سازی بات تلگرام پلنیکس است: دریافت آپدیت‌ها،
// کیبوردها، جریان‌های چندمرحله‌ای و پاسخ‌دهی به کاربر.
package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/AliGhanizade/planix-telegram-bot/internal/repository"
	"github.com/AliGhanizade/planix-telegram-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Bot نگهدارنده‌ی کلاینت تلگرام و وابستگی‌های بات است.
type Bot struct {
	api   *tgbotapi.BotAPI
	users *repository.UserRepository
	tasks *service.TaskService
	db    *gorm.DB
	log   *zap.Logger
	owner string
}

// New کلاینت تلگرام را با توکن داده‌شده می‌سازد.
func New(token, owner string, db *gorm.DB, log *zap.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, users: repository.NewUser(db), tasks: service.NewTask(db), db: db, log: log, owner: owner}, nil
}

// HandleUpdate آپدیت دریافتی را به هندلر مناسب (پیام یا کال‌بک) می‌سپارد.
func (b *Bot) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message != nil {
		return b.handleMessage(ctx, update.Message)
	}
	if update.CallbackQuery != nil {
		return b.handleCallback(ctx, update.CallbackQuery)
	}
	return nil
}

// reply پیام متنی می‌فرستد؛ اگر markup داده نشده باشد کیبورد اصلی می‌گذارد.
func (b *Bot) reply(chatID int64, text string, markup any) error {
	m := tgbotapi.NewMessage(chatID, text)
	if markup != nil {
		m.ReplyMarkup = markup
	} else {
		m.ReplyMarkup = MainKeyboard()
	}
	_, err := b.api.Send(m)
	return err
}

// setState وضعیت مکالمه‌ی کاربر و داده‌ی JSON آن را برای ۳۰ دقیقه ذخیره می‌کند.
func (b *Bot) setState(ctx context.Context, userID uuid.UUID, state, data string) error {
	s := domain.BotSession{UserID: userID, State: state, Data: data, ExpiresAt: time.Now().Add(30 * time.Minute)}
	return b.db.WithContext(ctx).Where("user_id = ?", userID).Assign(s).FirstOrCreate(&s).Error
}

// setStateAndReply وضعیت را ذخیره و پیام راهنمای مرحله‌ی بعد را می‌فرستد.
func (b *Bot) setStateAndReply(ctx context.Context, userID uuid.UUID, state, data string, chatID int64, text string, markup any) error {
	if err := b.setState(ctx, userID, state, data); err != nil {
		return err
	}
	return b.reply(chatID, text, markup)
}

// setTaskForOther چند تسک را به کاربر دیگر واگذار می‌کند و هر دو طرف را خبر می‌کند.
func (b *Bot) setTaskForOther(ctx context.Context, ownerID, assigneeID uuid.UUID, titles []string, priority string) error {
	for _, title := range titles {
		task := &domain.Task{OwnerID: ownerID, AssigneeID: assigneeID, Title: title, Priority: priority, Status: "pending"}
		if err := b.tasks.CreateForOtherUser(ctx, task); err != nil {
			return err
		}
	}
	assignee, err := b.users.GetByID(ctx, assigneeID)
	if err != nil {
		return err
	}
	owner, err := b.users.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}
	if err := b.reply(assignee.TelegramID, fmt.Sprintf("📣 گزارش پلنیکس\n%s %d تسک برای تو ثبت کرد:", owner.FirstName, len(titles)), MainKeyboard()); err != nil {
		return err
	}
	_ = b.reply(owner.TelegramID, fmt.Sprintf("ثبت %d تسک برای %s با موفقیت انجام شد ✅", len(titles), assignee.FirstName), MainKeyboard())
	return nil
}

// getTargetTask خلاصه‌ی وضعیت تسک‌های واگذارشده به کاربر هدف را برای مالک می‌فرستد.
func (b *Bot) getTargetTask(ctx context.Context, ownerID, targetID uuid.UUID) error {
	tasks, err := b.tasks.ListByOwnerAndAssignee(ctx, ownerID, targetID)
	if err != nil {
		return err
	}
	target, err := b.users.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	owner, err := b.users.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return b.reply(owner.TelegramID, fmt.Sprintf("شما تا کنون به %s تسکی نداده‌اید.", target.FirstName), MainKeyboard())
	}
	done := fmt.Sprintf("✅ وضعیت تسک‌های واگذارشده به %s:\n\n", target.FirstName)
	pending := fmt.Sprintf("⏳ وضعیت تسک‌های واگذارشده به %s:\n\n", target.FirstName)
	for _, t := range tasks {
		if t.Status == "completed" {
			done += FormatSmallInfo(&t) + "\n"
		} else {
			pending += FormatSmallInfo(&t) + "\n"
		}
	}
	if err := b.reply(owner.TelegramID, pending, MainKeyboard()); err != nil {
		return err
	}
	return b.reply(owner.TelegramID, done, MainKeyboard())
}

// sendInfoTask کارت کامل تسک را با دکمه‌های مدیریتی می‌فرستد.
func (b *Bot) sendInfoTask(ctx context.Context, chatID int64, taskID uuid.UUID) error {
	task, err := b.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	return b.reply(chatID, FormatTask(task), TaskInlineKeyboard(task))
}
