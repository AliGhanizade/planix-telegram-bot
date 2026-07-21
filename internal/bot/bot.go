package bot

import (
	"context"
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
