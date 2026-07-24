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

type Bot struct {
	api   *tgbotapi.BotAPI
	users *repository.UserRepository
	tasks *service.TaskService
	db    *gorm.DB
	log   *zap.Logger
}

func New(token string, db *gorm.DB, log *zap.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Bot{api: api, users: repository.NewUser(db), tasks: service.NewTask(db), db: db, log: log}, nil
}


func (b *Bot) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message != nil {
		return b.handleMessage(ctx, update.Message)
	}
	if update.CallbackQuery != nil {
		return b.handleCallback(ctx, update.CallbackQuery)
	}
	return nil
}

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
func (b *Bot) setState(ctx context.Context, userID uuid.UUID, state string, chatID int64, k any) error {
	s := domain.BotSession{UserID: userID, State: state, Data: "{}", ExpiresAt: time.Now().Add(2 * time.Minute)}
	if err := b.db.WithContext(ctx).Where("user_id = ?", userID).Assign(s).FirstOrCreate(&s).Error; err != nil {
		return err
	}
	return nil
}

func (b *Bot) setStateAndReply(ctx context.Context, userID uuid.UUID, state string, chatID int64, text string, k any) error {
	s := domain.BotSession{UserID: userID, State: state, Data: "{}", ExpiresAt: time.Now().Add(2 * time.Minute)}
	if err := b.db.WithContext(ctx).Where("user_id = ?", userID).Assign(s).FirstOrCreate(&s).Error; err != nil {
		return err
	}
	return b.reply(chatID, text, k)
}


func (b *Bot) setTaskForOther(ctx context.Context, ownerID uuid.UUID, assigneeID uuid.UUID, titles []string, priority string) error {
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
	message := fmt.Sprintf("📣 گزارش پلنیکس\n%s یک تسک برای تو ثبت کرد:", owner.FirstName)
	if err := b.reply(assignee.TelegramID, message, MainKeyboard()); err != nil {
		return err
	}
	// notify owner that task was assigned
	_ = b.reply(owner.TelegramID, fmt.Sprintf("تسک برای %s با موفقیت ثبت شد:", assignee.FirstName), MainKeyboard())
	return nil
}

// getTargetTask sends the owner a summary of tasks they assigned to a target user
func (b *Bot) getTargetTask(ctx context.Context, ownerID uuid.UUID, targetID uuid.UUID) error {
	// fetch tasks where owner is ownerID and assignee is targetID
	tasks, err := b.tasks.ListByOwnerAndAssignee(ctx, ownerID, targetID)
	if err != nil {
		return err
	}
	owner, err := b.users.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}
	target, err := b.users.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return b.reply(owner.TelegramID, fmt.Sprintf("شما تا کنون به %s تسکی ندادید.", target.FirstName), MainKeyboard())
	}
	compeletText := fmt.Sprintf("📋 وضعیت تسک‌های واگذار شده به %s:\n انجام شده \n", target.FirstName)
	pendingText := fmt.Sprintf("📋 وضعیت تسک‌های واگذار شده به %s:\n انجام نشده \n", target.FirstName)

	for _, t := range tasks {
		if t.Status == "completed" {
			compeletText += FormatSmallInfo(&t) + "\n"
		} else {
			pendingText += FormatSmallInfo(&t) + "\n"
		}
	}
	b.reply(owner.TelegramID, pendingText, MainKeyboard())

	return b.reply(owner.TelegramID, compeletText, MainKeyboard())
}

func (b *Bot) sendInfoTask(ctx context.Context, chatID int64, taskID uuid.UUID) error {
	task, err := b.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	return b.reply(
		chatID,
		FormatTask(task),
		TaskInlineKeyboard(task),
	)
}
