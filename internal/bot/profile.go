package bot

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"go.uber.org/zap"
)

// stateFieldEdit maps a profile field to its state name.
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

// showProfileEdit renders the profile edit menu with current values.
func (b *Bot) showProfileEdit(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	text := ui.ProfileEditText(
		displayName(u.FirstName, ui.NoUsername(l)),
		displayName(u.LastName, ui.NoUsername(l)),
		displayName(u.Timezone, "Asia/Tehran"),
		l,
	)
	if err := b.render(ctx, chatID, messageID, text, ui.ProfileEditInlineKeyboard(l)); err != nil {
		b.log.Error("render profile edit failed", zap.Error(err))
	}
}

// startProfileFieldEdit starts editing one profile field.
func (b *Bot) startProfileFieldEdit(ctx context.Context, u *domain.User, data string, chatID int64, messageID int) {
	// format: settings:edit:<field>
	parts := ui.SplitCallback(data)
	if len(parts) < 3 {
		return
	}
	state, ok := stateFieldEdit(parts[2])
	if !ok {
		return
	}

	l := b.lang(u)
	prompt := map[string]string{
		stateWaitingEditFirstName: ui.EditFirstNamePrompt(l),
		stateWaitingEditLastName:  ui.EditLastNamePrompt(l),
		stateWaitingEditTimezone:  ui.EditTimezonePrompt(l),
	}[state]

	if err := b.setSession(ctx, u.ID, state, sessionData{
		Back: &backRef{Kind: "profile", ChatID: chatID, MessageID: messageID},
	}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, prompt, ui.CancelInlineKeyboard(l)); err != nil {
		b.log.Error("render profile field prompt failed", zap.Error(err))
	}
}

// displayName returns the value or a fallback.
func displayName(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
