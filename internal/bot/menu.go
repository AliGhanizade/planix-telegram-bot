package bot

import (
	"context"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// showMenu shows the main menu in the existing message or sends a new one.
func (b *Bot) showMenu(ctx context.Context, chatID int64, messageID int, l ui.Lang) error {
	return b.render(ctx, chatID, messageID, ui.MenuText(l), ui.MenuInlineKeyboard(l))
}

// cbNav handles nav:* callbacks (menu navigation).
func (b *Bot) cbNav(ctx context.Context, q *models.CallbackQuery) {
	data := q.Data
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		b.answer(ctx, q, "")
		return
	}

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.log.Error("upsert user failed", zap.Error(err))
		b.answer(ctx, q, ui.ErrGeneric(b.lang(u)))
		return
	}
	l := b.lang(u)

	switch data {
	case "nav:noop":
		b.answer(ctx, q, "")

	case "nav:menu":
		b.answer(ctx, q, "")
		if err := b.showMenu(ctx, chatID, messageID, l); err != nil {
			b.log.Error("show menu failed", zap.Error(err))
		}

	case "nav:today":
		b.answer(ctx, q, "")
		if err := b.renderTaskList(ctx, chatID, messageID, u.ID, ui.FilterPending, 1, l); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}

	case "nav:new":
		b.answer(ctx, q, "")
		b.startNewTask(ctx, u, chatID, messageID)

	case "nav:search":
		b.answer(ctx, q, "")
		b.startSearch(ctx, u, chatID, messageID)

	case "nav:assign":
		b.answer(ctx, q, "")
		b.startAssign(ctx, u, chatID, messageID)

	case "nav:status":
		b.answer(ctx, q, "")
		b.sendStatusPick(ctx, u, chatID, messageID)

	case "nav:profile":
		b.answer(ctx, q, "")
		b.sendProfile(ctx, u, chatID, messageID)

	case "nav:help":
		b.answer(ctx, q, "")
		if err := b.render(ctx, chatID, messageID, ui.HelpMessage(l), ui.MenuInlineKeyboard(l)); err != nil {
			b.log.Error("render help failed", zap.Error(err))
		}

	case "nav:settings":
		b.answer(ctx, q, "")
		b.showSettings(ctx, u, chatID, messageID)

	default:
		b.answer(ctx, q, "")
	}
}

// cbState cancels the current flow from the cancel button.
func (b *Bot) cbState(ctx context.Context, q *models.CallbackQuery) {
	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, ui.ErrGeneric(ui.Fa))
		return
	}
	l := b.lang(u)
	if err := b.clearSession(ctx, u.ID); err != nil {
		b.log.Error("clear session failed", zap.Error(err))
	}
	b.answer(ctx, q, ui.CancelledToast(l))

	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		return
	}
	if err := b.showMenu(ctx, chatID, messageID, l); err != nil {
		b.log.Error("show menu failed", zap.Error(err))
	}
}

// cbUserPick picks a target user from suggestions and shows delegated status.
func (b *Bot) cbUserPick(ctx context.Context, q *models.CallbackQuery) {
	const prefix = "user:filter:fr:"
	if len(q.Data) <= len(prefix) {
		b.answer(ctx, q, "")
		return
	}
	username := q.Data[len(prefix):]

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, ui.ErrGeneric(b.lang(u)))
		return
	}
	l := b.lang(u)
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		b.answerAlert(ctx, q, ui.UsernameNotFoundToast(l))
		return
	}
	_ = b.clearSession(ctx, u.ID)
	b.answer(ctx, q, "")

	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = u.TelegramID
	}
	if err := b.renderDelegatedStatus(ctx, chatID, messageID, u.ID, target.ID, l); err != nil {
		b.log.Error("render delegated status failed", zap.Error(err))
	}
}

// cbSettings handles the settings menu callbacks.
func (b *Bot) cbSettings(ctx context.Context, q *models.CallbackQuery) {
	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, ui.ErrGeneric(ui.Fa))
		return
	}
	l := b.lang(u)
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = u.TelegramID
		messageID = 0
	}

	switch q.Data {
	case "settings:profile":
		b.answer(ctx, q, "")
		b.showProfileEdit(ctx, u, chatID, messageID)

	case "settings:lang":
		newLang := string(l.Toggle())
		if err := b.profiles.SetLanguage(ctx, u.ID, newLang); err != nil {
			b.log.Error("set language failed", zap.Error(err))
			b.answer(ctx, q, ui.ErrGeneric(l))
			return
		}
		u.Lang = newLang
		b.answer(ctx, q, "")
		b.showSettings(ctx, u, chatID, messageID)

	case "settings:times":
		b.answer(ctx, q, "")
		b.showReportTimes(ctx, u, chatID, messageID)

	case "settings:addtime":
		b.answer(ctx, q, "")
		b.startAddReportTime(ctx, u, chatID, messageID)

	case "settings:web":
		b.answer(ctx, q, "")
		b.issueWebCodeFromBot(ctx, u, chatID, messageID)

	case "settings:edit:firstname", "settings:edit:lastname", "settings:edit:timezone":
		b.answer(ctx, q, "")
		b.startProfileFieldEdit(ctx, u, q.Data, chatID, messageID)

	case "settings:back":
		b.answer(ctx, q, "")
		b.showSettings(ctx, u, chatID, messageID)

	case "settings:daily":
		newValue := !u.DailyReport
		if err := b.users.SetDailyReport(ctx, u.ID, newValue); err != nil {
			b.log.Error("set daily report failed", zap.Error(err))
			b.answer(ctx, q, "خطا! دوباره تلاش کن")
			return
		}
		u.DailyReport = newValue
		if newValue {
			b.answer(ctx, q, ui.DailyReportLabel(true, l))
		} else {
			b.answer(ctx, q, ui.DailyReportLabel(false, l))
		}
		b.showSettings(ctx, u, chatID, messageID)
	}
}

// showSettings renders the settings screen.
func (b *Bot) showSettings(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	text := ui.SettingsText(ui.DailyReportLabel(u.DailyReport, l), l)
	if err := b.render(ctx, chatID, messageID, text, ui.SettingsInlineKeyboard(u.DailyReport, l)); err != nil {
		b.log.Error("render settings failed", zap.Error(err))
	}
}

// startNewTask starts the new task flow.
func (b *Bot) startNewTask(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingTaskTitle, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, ui.NewTaskPrompt(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u))); err != nil {
		b.log.Error("render new task prompt failed", zap.Error(err))
	}
}

// startSearch starts the search flow.
func (b *Bot) startSearch(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingSearch, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, ui.SearchPrompt(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u))); err != nil {
		b.log.Error("render search prompt failed", zap.Error(err))
	}
}

// startAssign starts the task assignment flow.
func (b *Bot) startAssign(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingTaskForOther, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, ui.AssignPrompt(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u))); err != nil {
		b.log.Error("render assign prompt failed", zap.Error(err))
	}
}

// sendStatusPick shows the user suggestions keyboard for status tracking.
func (b *Bot) sendStatusPick(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingStatusForOther, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	users, err := b.users.ListUsersAssignedTo(ctx, u.ID)
	if err != nil {
		b.log.Error("list assigned users failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, ui.StatusPickPrompt(b.lang(u)), ui.SuggestFriendInlineKeyboard(users, b.lang(u))); err != nil {
		b.log.Error("render status pick failed", zap.Error(err))
	}
}

// sendProfile renders the user profile with stats.
func (b *Bot) sendProfile(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	st, err := b.tasks.Stats(ctx, u.ID)
	if err != nil {
		b.log.Error("task stats failed", zap.Error(err))
	}

	l := b.lang(u)
	name := (u.FirstName + " " + u.LastName)
	if name == " " {
		name = ui.ProfileDefaultName(l)
	}
	username := ui.NoUsername(l)
	if u.Username != "" {
		username = "@" + u.Username
	}

	text := ui.ProfileText(name, username, st.Pending, st.Completed, st.Cancelled, st.CompletionRate(), l)
	if err := b.render(ctx, chatID, messageID, text, ui.MenuInlineKeyboard(l)); err != nil {
		b.log.Error("render profile failed", zap.Error(err))
	}
}

// renderDelegatedStatus shows delegated task status for a target user in one message.
func (b *Bot) renderDelegatedStatus(ctx context.Context, chatID int64, messageID int, ownerID, targetID uuid.UUID, l ui.Lang) error {
	tasks, err := b.tasks.ListByOwnerAndAssignee(ctx, ownerID, targetID)
	if err != nil {
		return err
	}
	target, err := b.users.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		text := ui.DelegatedEmpty(target.FirstName, l)
		return b.render(ctx, chatID, messageID, text, ui.BackKeyboard(l))
	}

	done := ""
	pending := ""
	for _, t := range tasks {
		if t.Status == "completed" {
			done += ui.FormatSmallInfo(&t, l) + "\n"
		} else {
			pending += ui.FormatSmallInfo(&t, l) + "\n"
		}
	}

	text := ui.DelegatedHeader(target.FirstName, l)
	if pending != "" {
		text += ui.DelegatedOpenTitle(l) + pending + "\n"
	}
	if done != "" {
		text += ui.DelegatedDoneTitle(l) + done
	}
	return b.render(ctx, chatID, messageID, text, ui.BackKeyboard(l))
}
