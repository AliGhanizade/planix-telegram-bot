package bot

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

// SendLoginCode dms the web login code to a user.
// called by httpapi for the web initiated login flow.
func (b *Bot) SendLoginCode(ctx context.Context, telegramID int64, code string, ttlMinutes int, l ui.Lang) error {
	copyLabel := "📋 کپی کد"
	if l == ui.En {
		copyLabel = "📋 Copy code"
	}
	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: copyLabel, CopyText: &models.CopyTextButton{Text: code}}},
		},
	}
	_, err := b.send(ctx, telegramID, ui.WebCodeMessage(code, ttlMinutes, l), kb)
	return err
}

// issueWebCodeFromBot is the in-chat web linking flow: issue a code and show it here.
func (b *Bot) issueWebCodeFromBot(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	lc, err := b.auth.IssueLoginCode(ctx, u.ID, "bot")
	if err != nil {
		b.log.Error("issue login code failed", zap.Error(err))
		b.render(ctx, chatID, messageID, ui.IssueCodeError(l), ui.BackKeyboard(l))
		return
	}

	copyLabel, newCode := "📋 کپی کد", "🔄 کد جدید"
	if l == ui.En {
		copyLabel, newCode = "📋 Copy code", "🔄 New code"
	}
	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: copyLabel, CopyText: &models.CopyTextButton{Text: lc.Code}}},
			{{Text: newCode, CallbackData: "settings:web", Style: ui.StylePrimary}},
			{{Text: ui.BackLabel(l), CallbackData: "nav:menu"}},
		},
	}
	text := ui.WebCodeMessage(lc.Code, 10, l)
	if err := b.render(ctx, chatID, messageID, text, kb); err != nil {
		b.log.Error("render web code failed", zap.Error(err))
	}
}

// NotifyAssignment dms a user about a task just delegated to them from the web panel.
func (b *Bot) NotifyAssignment(ctx context.Context, task *domain.Task, ownerName string) {
	assignee, err := b.users.GetByID(ctx, task.AssigneeID)
	if err != nil {
		return
	}
	l := b.lang(assignee)
	text := ui.AssignedNotify(ownerName, 1, l) + "\n" + ui.FormatSmallInfo(task, l)
	_, _ = b.sendWithKeyboard(ctx, assignee.TelegramID, text, l)
}
