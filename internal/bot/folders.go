package bot

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// showFolders lists the folders of the user with inline navigation.
func (b *Bot) showFolders(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	views, err := b.folders.List(ctx, u.ID)
	if err != nil {
		b.log.Error("list folders failed", zap.Error(err))
		return
	}

	var rows [][]models.InlineKeyboardButton
	for _, v := range views {
		label := "📁 " + ui.Truncate(v.Folder.Name, 30)
		if v.Kind == "shared" {
			if l == ui.En {
				label = "📥 " + ui.Truncate(v.Folder.Name, 28)
			} else {
				label = "📥 " + ui.Truncate(v.Folder.Name, 28)
			}
		}
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: label, CallbackData: "fopen:" + v.Folder.ID.String()},
		})
	}
	if len(rows) == 0 {
		empty := "هنوز پوشه‌ای نداری"
		if l == ui.En {
			empty = "No folders yet"
		}
		rows = append(rows, []models.InlineKeyboardButton{{Text: empty, CallbackData: "nav:noop"}})
	}
	newBtn := "📁 پوشه جدید"
	if l == ui.En {
		newBtn = "📁 New folder"
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: newBtn, CallbackData: "fnew", Style: ui.StylePrimary},
	})
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: ui.BackLabel(l), CallbackData: "nav:menu", Style: ui.StylePrimary},
	})

	title := "📁 پوشه‌ها\n\nپوشه‌ای را باز کن تا تسک‌هایش را ببینی."
	if l == ui.En {
		title = "📁 Folders\n\nOpen a folder to see its tasks."
	}
	if err := b.render(ctx, chatID, messageID, title, models.InlineKeyboardMarkup{InlineKeyboard: rows}); err != nil {
		b.log.Error("render folders failed", zap.Error(err))
	}
}

// showFolderTasks lists the tasks inside one folder.
func (b *Bot) showFolderTasks(ctx context.Context, u *domain.User, chatID int64, messageID int, folderID uuid.UUID) {
	l := b.lang(u)
	folder, err := b.folders.GetForUser(ctx, u.ID, folderID)
	if err != nil {
		b.answerAlert(ctx, nil, ui.TaskNotFoundToast(l))
		return
	}
	tasks, err := b.folders.TasksInFolder(ctx, u.ID, folderID)
	if err != nil {
		b.log.Error("folder tasks failed", zap.Error(err))
		return
	}

	text := "📁 " + ui.Esc(folder.Name) + "\n\n"
	if len(tasks) == 0 {
		text += ui.FolderEmpty(l)
	} else {
		for _, t := range tasks {
			text += ui.FormatSmallInfo(&t, l) + "\n"
		}
	}

	var rows [][]models.InlineKeyboardButton
	rows = append(rows, ui.TasksInlineKeyboard(tasks, l).InlineKeyboard...)
	newBtn := "📁 پوشه جدید"
	if l == ui.En {
		newBtn = "📁 New folder"
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: newBtn, CallbackData: "fnew", Style: ui.StylePrimary},
	})
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: ui.BackLabel(l), CallbackData: "folders:open", Style: ui.StylePrimary},
	})

	if err := b.render(ctx, chatID, messageID, text, models.InlineKeyboardMarkup{InlineKeyboard: rows}); err != nil {
		b.log.Error("render folder tasks failed", zap.Error(err))
	}
}

// startNewFolder asks for the folder name.
func (b *Bot) startNewFolder(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	if err := b.setSession(ctx, u.ID, stateWaitingFolderName, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, ui.FolderNewPrompt(l), ui.CancelInlineKeyboard(l)); err != nil {
		b.log.Error("render folder prompt failed", zap.Error(err))
	}
}

// cbFolders routes folder:* callbacks.
func (b *Bot) cbFolders(ctx context.Context, q *models.CallbackQuery) {
	data := q.Data
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = q.From.ID
		messageID = 0
	}
	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, ui.ErrGeneric(b.lang(u)))
		return
	}

	switch {
	case data == "folders:open":
		b.answer(ctx, q, "")
		b.showFolders(ctx, u, chatID, messageID)

	case data == "fnew":
		b.answer(ctx, q, "")
		b.startNewFolder(ctx, u, chatID, messageID)

	case strings.HasPrefix(data, "fopen:"):
		b.answer(ctx, q, "")
		id, err := uuid.Parse(strings.TrimPrefix(data, "fopen:"))
		if err != nil {
			return
		}
		b.showFolderTasks(ctx, u, chatID, messageID, id)
	}
}

// handleFolderNameInput creates a folder from the user text.
func (b *Bot) handleFolderNameInput(ctx context.Context, u *domain.User, text string, chatID int64) error {
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	f, err := b.folders.Create(ctx, u.ID, text, nil)
	if err != nil {
		_, rerr := b.send(ctx, chatID, ui.ValidationError(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u)))
		if rerr != nil {
			return err
		}
		return nil
	}
	_, err = b.send(ctx, chatID, ui.FolderCreated(f.Name, b.lang(u)), ui.MainKeyboard(b.lang(u)))
	return err
}

// showFolderPick lists folders to link or unlink the session task.
func (b *Bot) showFolderPick(ctx context.Context, u *domain.User, taskID uuid.UUID, chatID int64, messageID int) {
	l := b.lang(u)
	views, err := b.folders.List(ctx, u.ID)
	if err != nil {
		b.log.Error("list folders failed", zap.Error(err))
		return
	}
	linked := map[uuid.UUID]bool{}
	links, _ := b.folders.LinksOfTask(ctx, u.ID, taskID)
	for _, id := range links {
		linked[id] = true
	}

	refs := make([]ui.FolderRef, 0, len(views))
	for _, v := range views {
		if v.Kind != "own" {
			continue
		}
		refs = append(refs, ui.FolderRef{ID: v.Folder.ID.String(), Name: v.Folder.Name, Linked: linked[v.Folder.ID]})
	}

	if err := b.setSession(ctx, u.ID, stateWaitingFolderPick, sessionData{TaskID: taskID.String()}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	hint := "روی پوشه بزن تا تسک داخلش قرار بگیرد یا ازش خارج شود."
	if l == ui.En {
		hint = "Tap a folder to put the task inside or take it out."
	}
	text := "📁 پوشه‌ها\n\n" + hint
	if err := b.render(ctx, chatID, messageID, text, ui.FolderPickKeyboard(refs, l)); err != nil {
		b.log.Error("render folder pick failed", zap.Error(err))
	}
}

// cbFolderPick toggles the session task membership in a folder.
func (b *Bot) cbFolderPick(ctx context.Context, q *models.CallbackQuery, folderIDStr string) {
	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, ui.ErrGeneric(ui.Fa))
		return
	}
	l := b.lang(u)
	session, ok, err := b.activeSession(ctx, u.ID)
	if err != nil || !ok || session.State != stateWaitingFolderPick {
		b.answer(ctx, q, "")
		return
	}
	var data sessionData
	_ = json.Unmarshal([]byte(session.Data), &data)
	taskID, err := uuid.Parse(data.TaskID)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	links, _ := b.folders.LinksOfTask(ctx, u.ID, taskID)
	already := false
	for _, id := range links {
		if id == folderID {
			already = true
			break
		}
	}
	if already {
		if err := b.folders.Unlink(ctx, u.ID, taskID, folderID); err != nil {
			b.answer(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.FolderUnlinkedToast(l))
	} else {
		if err := b.folders.Link(ctx, u.ID, taskID, folderID); err != nil {
			b.answerAlert(ctx, q, err.Error())
			return
		}
		b.answer(ctx, q, ui.FolderLinkedToast(l))
	}

	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		return
	}
	b.showFolderPick(ctx, u, taskID, chatID, messageID)
}

// showReportTimes renders the daily report times editor.
func (b *Bot) showReportTimes(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	times := parseReportTimes(u.ReportTimes)
	if len(times) == 0 {
		times = []string{"21:00"}
	}

	text := ui.ReportTimesTitle(strings.Join(times, " ، "), l)
	var rows [][]models.InlineKeyboardButton
	for _, t := range times {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: "🗑 " + t, CallbackData: "rtime:" + t, Style: ui.StyleDanger},
		})
	}
	add := "➕ افزودن ساعت"
	if l == ui.En {
		add = "➕ Add time"
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: add, CallbackData: "settings:addtime", Style: ui.StylePrimary},
	})
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: ui.BackLabel(l), CallbackData: "nav:menu", Style: ui.StylePrimary},
	})

	if err := b.render(ctx, chatID, messageID, text, models.InlineKeyboardMarkup{InlineKeyboard: rows}); err != nil {
		b.log.Error("render report times failed", zap.Error(err))
	}
}

// startAddReportTime asks for a new report time.
func (b *Bot) startAddReportTime(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	l := b.lang(u)
	prompt := "⏳ ساعت جدید را بفرست. مثال: 14:30 (به وقت تهران، حداکثر ۳ ساعت)"
	if l == ui.En {
		prompt = "⏳ Send the new time. Example: 14:30 (Tehran time, up to 3 times)"
	}
	if err := b.setSession(ctx, u.ID, stateWaitingReportTime, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, prompt, ui.CancelInlineKeyboard(l)); err != nil {
		b.log.Error("render add time prompt failed", zap.Error(err))
	}
}

// handleReportTimeInput parses and stores a new report time.
func (b *Bot) handleReportTimeInput(ctx context.Context, u *domain.User, text string, chatID int64) error {
	l := b.lang(u)
	value := strings.TrimSpace(text)
	if _, err := time.Parse("15:04", value); err != nil {
		_, rerr := b.send(ctx, chatID, ui.InvalidTime(l), ui.CancelInlineKeyboard(l))
		if rerr != nil {
			return err
		}
		return nil
	}
	times := parseReportTimes(u.ReportTimes)
	for _, t := range times {
		if t == value {
			_, _ = b.send(ctx, chatID, ui.InvalidTime(l), ui.MainKeyboard(l))
			_ = b.clearSession(ctx, u.ID)
			return nil
		}
	}
	times = append(times, value)
	if _, err := b.profiles.SetReportTimes(ctx, u.ID, times); err != nil {
		_, _ = b.send(ctx, chatID, err.Error(), ui.CancelInlineKeyboard(l))
		return nil
	}
	_ = b.clearSession(ctx, u.ID)
	b.showReportTimes(ctx, u, chatID, 0)
	return nil
}

// parseReportTimes splits the stored times.
func parseReportTimes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{"21:00"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// removeReportTime drops one time from the user settings.
func (b *Bot) removeReportTime(ctx context.Context, q *models.CallbackQuery, u *domain.User, value string, chatID int64, messageID int) {
	l := b.lang(u)
	times := parseReportTimes(u.ReportTimes)
	kept := make([]string, 0, len(times))
	for _, t := range times {
		if t != value {
			kept = append(kept, t)
		}
	}
	if _, err := b.profiles.SetReportTimes(ctx, u.ID, kept); err != nil {
		b.log.Error("set report times failed", zap.Error(err))
		b.answer(ctx, q, ui.ErrGeneric(l))
		return
	}
	b.answer(ctx, q, "🗑 حذف شد")
	b.showReportTimes(ctx, u, chatID, messageID)
}

// SendDailyReportsAtMinute sends the daily report to every user whose
// chosen time matches the current tehran time.
func (b *Bot) SendDailyReportsAtMinute(ctx context.Context, hhmm string) {
	users, err := b.users.ListActiveWithDailyReport(ctx)
	if err != nil {
		b.log.Error("list daily report users failed", zap.Error(err))
		return
	}

	sent := 0
	for _, u := range users {
		if !containsTime(parseReportTimes(u.ReportTimes), hhmm) {
			continue
		}
		l := b.lang(&u)
		tasks, err := b.tasks.Today(ctx, u.ID)
		if err != nil {
			b.log.Error("list today tasks failed", zap.Error(err), zap.String("user_id", u.ID.String()))
			continue
		}

		text := ui.DailyReportHeader(l)
		if len(tasks) == 0 {
			text += ui.DailyReportEmpty(l)
		} else {
			text += ui.DailyReportCount(len(tasks), l)
			for _, t := range tasks {
				text += ui.FormatSmallInfo(&t, l) + "\n"
			}
		}

		if _, err := b.send(ctx, u.TelegramID, text, ui.TasksInlineKeyboard(tasks, l)); err != nil {
			b.log.Warn("send daily report failed", zap.Error(err), zap.Int64("telegram_id", u.TelegramID))
			continue
		}
		sent++
	}
	if sent > 0 {
		b.log.Info("daily reports sent", zap.Int("sent", sent), zap.String("at", hhmm))
	}
}

func containsTime(times []string, value string) bool {
	for _, t := range times {
		if t == value {
			return true
		}
	}
	return false
}

// TehranHHMM returns the current tehran wall clock as HH:MM.
func TehranHHMM() string {
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		loc = time.UTC
	}
	return time.Now().In(loc).Format("15:04")
}
