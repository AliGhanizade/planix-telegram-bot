package bot

import (
	"context"
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"go.uber.org/zap"
)

// stateFieldEdit نام وضعیت برای هر فیلد پروفایل.
func stateFieldEdit(field string) (string, bool) {
	switch field {
	case "firstname":
		return stateWaitingEditFirstName, true
	case "lastname":
		return stateWaitingEditLastName, true
	case "timezone":
		return stateWaitingEditTimezone, true
	}
	return "", false
}

// showProfileEdit منوی ویرایش اطلاعات کاربر را با مقادیر فعلی رندر می‌کند.
func (b *Bot) showProfileEdit(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	text := fmt.Sprintf("✏️ ویرایش اطلاعات\n\n"+
		"🪪 نام: %s\n"+
		"🏷 نام خانوادگی: %s\n"+
		"🌍 تایم‌زون: %s\n\n"+
		"روی فیلد موردنظر بزن و مقدار جدید را بفرست.",
		displayName(u.FirstName, "ندارد"),
		displayName(u.LastName, "ندارد"),
		displayName(u.Timezone, "Asia/Tehran"),
	)
	if err := b.render(ctx, chatID, messageID, text, ui.ProfileEditInlineKeyboard()); err != nil {
		b.log.Error("render profile edit failed", zap.Error(err))
	}
}

// startProfileFieldEdit شروع ویرایش یک فیلد مشخص پروفایل.
func (b *Bot) startProfileFieldEdit(ctx context.Context, u *domain.User, data string, chatID int64, messageID int) {
	// فرمت: settings:edit:<field>
	parts := ui.SplitCallback(data)
	if len(parts) < 3 {
		return
	}
	state, ok := stateFieldEdit(parts[2])
	if !ok {
		return
	}

	prompt := map[string]string{
		stateWaitingEditFirstName: "🪪 نام جدیدت را بفرست:",
		stateWaitingEditLastName:  "🏷 نام خانوادگی جدید را بفرست (برای حذف، «-» بفرست):",
		stateWaitingEditTimezone:  "🌍 تایم‌زون را بفرست؛ مثال: Asia/Tehran",
	}[state]

	if err := b.setSession(ctx, u.ID, state, sessionData{
		Back: &backRef{Kind: "profile", ChatID: chatID, MessageID: messageID},
	}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, prompt, ui.CancelInlineKeyboard()); err != nil {
		b.log.Error("render profile field prompt failed", zap.Error(err))
	}
}

// displayName مقدار را با جایگزین پیش‌فرض برمی‌گرداند.
func displayName(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
