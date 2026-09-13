package bot

import (
	"context"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"go.uber.org/zap"
)

// SendDailyReports sends the daily digest to users who kept it enabled.
// invoked by the cron scheduler.
func (b *Bot) SendDailyReports(ctx context.Context) {
	users, err := b.users.ListActiveWithDailyReport(ctx)
	if err != nil {
		b.log.Error("list daily report users failed", zap.Error(err))
		return
	}

	sent := 0
	for _, u := range users {
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

		if _, err := b.sendWithKeyboard(ctx, u.TelegramID, text, l); err != nil {
			b.log.Warn("send daily report failed", zap.Error(err), zap.Int64("telegram_id", u.TelegramID))
			continue
		}
		sent++
	}

	b.log.Info("daily reports sent", zap.Int("sent", sent), zap.Int("users", len(users)))
}

// SendDueReminders notifies assignees of tasks due within the lead window;
// each task is reminded only once.
func (b *Bot) SendDueReminders(ctx context.Context, lead time.Duration) {
	now := time.Now()
	tasks, err := b.tasks.DueSoon(ctx, now, now.Add(lead))
	if err != nil {
		b.log.Error("list due soon tasks failed", zap.Error(err))
		return
	}

	for _, t := range tasks {
		assignee, err := b.users.GetByID(ctx, t.AssigneeID)
		if err != nil {
			b.log.Warn("fetch assignee failed", zap.Error(err))
			continue
		}
		l := b.lang(assignee)

		remaining := time.Until(*t.DueAt)
		text := ui.ReminderText(
			ui.Truncate(t.Title, 60),
			ui.HumanDuration(remaining, l),
			t.DueAt.Format("01-02 15:04"),
			l,
		)
		if _, err := b.sendWithKeyboard(ctx, assignee.TelegramID, text, l); err != nil {
			b.log.Warn("send reminder failed", zap.Error(err))
			continue
		}
		if err := b.tasks.MarkReminded(ctx, t.ID, now); err != nil {
			b.log.Warn("mark reminded failed", zap.Error(err))
		}
	}

	if len(tasks) > 0 {
		b.log.Info("due reminders sent", zap.Int("count", len(tasks)))
	}
}
