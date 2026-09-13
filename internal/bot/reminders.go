package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"go.uber.org/zap"
)

// SendDailyReports گزارش روزانه را برای همه‌ی کاربرانی که روشن گذاشته‌اند می‌فرستد.
// توسط زمان‌بند (cron) صدا زده می‌شود.
func (b *Bot) SendDailyReports(ctx context.Context) {
	users, err := b.users.ListActiveWithDailyReport(ctx)
	if err != nil {
		b.log.Error("list daily report users failed", zap.Error(err))
		return
	}

	sent := 0
	for _, u := range users {
		tasks, err := b.tasks.Today(ctx, u.ID)
		if err != nil {
			b.log.Error("list today tasks failed", zap.Error(err), zap.String("user_id", u.ID.String()))
			continue
		}

		text := "🌅 گزارش روزانه پلنیکس\n\n"
		if len(tasks) == 0 {
			text += "امروز تسک بازی نداری؛ روز خوبی داشته باشی ✨"
		} else {
			text += fmt.Sprintf("شما %d تسک باز دارید:\n\n", len(tasks))
			for _, t := range tasks {
				text += ui.FormatSmallInfo(&t) + "\n"
			}
		}

		if _, err := b.sendWithKeyboard(ctx, u.TelegramID, text); err != nil {
			b.log.Warn("send daily report failed", zap.Error(err), zap.Int64("telegram_id", u.TelegramID))
			continue
		}
		sent++
	}

	b.log.Info("daily reports sent", zap.Int("sent", sent), zap.Int("users", len(users)))
}

// SendDueReminders برای تسک‌هایی که تا مدت lead آینده موعدشان تمام می‌شود
// یادآوری می‌فرستد و هر تسک فقط یک بار یادآوری می‌شود.
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

		remaining := time.Until(*t.DueAt)
		text := fmt.Sprintf("⏳ یادآوری پلنیکس\n\nتسک «%s» تا %s دیگر موعدش تمام می‌شود.\n📅 موعد: %s",
			ui.Truncate(t.Title, 60),
			humanDuration(remaining),
			t.DueAt.Format("01-02 15:04"),
		)
		if _, err := b.sendWithKeyboard(ctx, assignee.TelegramID, text); err != nil {
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

// humanDuration مدت زمان را به فارسی خوانا می‌کند.
func humanDuration(d time.Duration) string {
	if d <= 0 {
		return "همین حالا"
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%d دقیقه", minutes)
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%d ساعت", hours)
	}
	return fmt.Sprintf("%d روز", int(d.Hours()/24))
}
