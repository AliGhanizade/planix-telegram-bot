package bot

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

// SendLoginCode کد ورود پنل وب را در چت خصوصی کاربر می‌فرستد.
// توسط httpapi برای جریان لاگین «وب‌محور» صدا زده می‌شود.
func (b *Bot) SendLoginCode(ctx context.Context, telegramID int64, code string, ttlMinutes int) error {
	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "📋 کپی کد", CopyText: &models.CopyTextButton{Text: code}}},
		},
	}
	_, err := b.send(ctx, telegramID, ui.WebCodeMessage(code, ttlMinutes), kb)
	return err
}

// issueWebCodeFromBot جریان «اتصال پنل وب» از داخل بات: کد صادر و در همین چت نشان داده می‌شود.
func (b *Bot) issueWebCodeFromBot(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	lc, err := b.auth.IssueLoginCode(ctx, u.ID, "bot")
	if err != nil {
		b.log.Error("issue login code failed", zap.Error(err))
		b.render(ctx, chatID, messageID, "خطا در صدور کد؛ دوباره تلاش کن.", ui.BackKeyboard())
		return
	}

	kb := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "📋 کپی کد", CopyText: &models.CopyTextButton{Text: lc.Code}}},
			{{Text: "🔄 کد جدید", CallbackData: "settings:web", Style: ui.StylePrimary}},
			{{Text: "🏠 منوی اصلی", CallbackData: "nav:menu"}},
		},
	}
	text := ui.WebCodeMessage(lc.Code, 10)
	if err := b.render(ctx, chatID, messageID, text, kb); err != nil {
		b.log.Error("render web code failed", zap.Error(err))
	}
}
