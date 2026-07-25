package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// handleCallback دکمه‌های شیشه‌ای (callback) را پردازش می‌کند.
func (b *Bot) handleCallback(ctx context.Context, q *tgbotapi.CallbackQuery) error {
	data := q.Data

	switch {
	// ---------- اطلاعات و بروزرسانی ----------
	case strings.HasPrefix(data, "task:info:"):
		id, err := parseTaskAction(data, "task:info:")
		if err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, ""))
		return b.sendInfoTask(ctx, q.Message.Chat.ID, id)

	case strings.HasPrefix(data, "task:refresh:"):
		id, err := parseTaskAction(data, "task:refresh:")
		if err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "بروزرسانی شد 🔄"))
		return b.sendInfoTask(ctx, q.Message.Chat.ID, id)

	// ---------- انجام تسک ----------
	case strings.HasPrefix(data, "task:done:"):
		id, err := parseTaskAction(data, "task:done:")
		if err != nil {
			return err
		}
		task, err := b.tasks.Complete(ctx, id)
		if err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "تسک انجام شد ✅"))
		_ = b.reply(q.Message.Chat.ID, fmt.Sprintf("🎉 تسک «%s» با موفقیت انجام شد.", task.Title), MainKeyboard())
		if task.OwnerID != task.AssigneeID {
			if owner, err := b.users.GetByID(ctx, task.OwnerID); err == nil {
				_ = b.reply(owner.TelegramID, fmt.Sprintf("📣 %s تسک «%s» را انجام داد.", q.From.FirstName, task.Title), MainKeyboard())
			}
		}
		return nil

	// ---------- انجام نشد ----------
	case strings.HasPrefix(data, "task:cancel:"):
		id, err := parseTaskAction(data, "task:cancel:")
		if err != nil {
			return err
		}
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "ثبت شد ❌"))
		return b.reply(q.Message.Chat.ID, fmt.Sprintf("❌ تسک «%s» انجام نشد.", task.Title), MainKeyboard())

	// ---------- ویرایش عنوان و توضیحات ----------
	case strings.HasPrefix(data, "task:edit:title:"):
		id, err := parseTaskAction(data, "task:edit:title:")
		if err != nil {
			return err
		}
		if err := b.startEditSession(ctx, q.From.ID, stateWaitingTaskTitle, id); err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "عنوان جدید را ارسال کنید ✏️"))
		return b.reply(q.Message.Chat.ID, "📝 لطفاً عنوان جدید تسک را ارسال کنید.", CancelStateInlineKeyboard())

	case strings.HasPrefix(data, "task:edit:desc:"):
		id, err := parseTaskAction(data, "task:edit:desc:")
		if err != nil {
			return err
		}
		if err := b.startEditSession(ctx, q.From.ID, stateWaitingTaskDescription, id); err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "توضیحات جدید را ارسال کنید ✏️"))
		return b.reply(q.Message.Chat.ID, "📄 لطفاً توضیحات جدید تسک را ارسال کنید.", CancelStateInlineKeyboard())

	// ---------- قابلیت‌های در دست ساخت ----------
	case strings.HasPrefix(data, "task:edit:due:"), strings.HasPrefix(data, "task:edit:priority:"):
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "این قابلیت به‌زودی اضافه می‌شود ⏳"))
		return nil

	// ---------- حذف ----------
	case strings.HasPrefix(data, "task:delete:"):
		id, err := parseTaskAction(data, "task:delete:")
		if err != nil {
			return err
		}
		if err := b.tasks.Delete(ctx, id); err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "حذف شد 🗑"))
		return b.reply(q.Message.Chat.ID, "🗑 تسک با موفقیت حذف شد.", MainKeyboard())

	// ---------- لغو جریان جاری ----------
	case strings.HasPrefix(data, "state:cancel"):
		user, err := b.users.GetByTelegramID(ctx, q.From.ID)
		if err != nil {
			return err
		}
		if err := b.deleteState(ctx, user.ID); err != nil {
			return err
		}
		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "لغو شد ❌"))
		return b.reply(q.Message.Chat.ID, "❌ عملیات لغو شد.", MainKeyboard())

	// ---------- انتخاب کاربر هدف ----------
	case strings.HasPrefix(data, "user:filter:fr:"):
		username := strings.TrimPrefix(data, "user:filter:fr:")
		user, err := b.users.GetByTelegramID(ctx, q.From.ID)
		if err != nil {
			return err
		}
		target, err := b.users.GetByUsername(ctx, username)
		if err != nil {
			_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "یوزرنیم پیدا نشد ❌"))
			return b.reply(user.TelegramID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", MainKeyboard())
		}
		if err := b.deleteState(ctx, user.ID); err != nil {
			return err
		}
		return b.getTargetTask(ctx, user.ID, target.ID)
	}
	return nil
}

// parseTaskAction شناسه‌ی تسک را از دیتای دکمه بیرون می‌کشد.
func parseTaskAction(data, prefix string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimPrefix(data, prefix))
}

// startEditSession برای ویرایش یک تسک، نشست کاربر را با شناسه‌ی تسک آماده می‌کند.
func (b *Bot) startEditSession(ctx context.Context, telegramID int64, state string, taskID uuid.UUID) error {
	u := &domain.User{TelegramID: telegramID}
	if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
		return err
	}
	payload, err := json.Marshal(taskPayload{TaskID: taskID.String()})
	if err != nil {
		return err
	}
	return b.setState(ctx, u.ID, state, string(payload))
}
