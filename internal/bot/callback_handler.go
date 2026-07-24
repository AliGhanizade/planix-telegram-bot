package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func (b *Bot) handleCallback(ctx context.Context, q *tgbotapi.CallbackQuery) error {
	data := q.Data

	switch {

	// ===========================
	// اطلاعات تسک
	// ===========================
	case strings.HasPrefix(data, "task:info:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:info:"))
		if err != nil {
			return err
		}

		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, ""))

		return b.sendInfoTask(ctx, q.Message.Chat.ID, id)

	// ===========================
	// بروزرسانی اطلاعات
	// ===========================
	case strings.HasPrefix(data, "task:refresh:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:refresh:"))
		if err != nil {
			return err
		}

		_, _ = b.api.Request(tgbotapi.NewCallback(q.ID, "بروزرسانی شد 🔄"))

		return b.sendInfoTask(ctx, q.Message.Chat.ID, id)

	// ===========================
	// انجام تسک
	// ===========================
	case strings.HasPrefix(data, "task:done:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:done:"))
		if err != nil {
			return err
		}

		task, err := b.tasks.Complete(ctx, id)
		if err != nil {
			return err
		}

		_, _ = b.api.Request(
			tgbotapi.NewCallback(q.ID, "تسک انجام شد ✅"),
		)

		_ = b.reply(
			q.Message.Chat.ID,
			fmt.Sprintf("🎉 تسک «%s» با موفقیت انجام شد.", task.Title),
			MainKeyboard(),
		)

		if task.OwnerID != task.AssigneeID {
			owner, err := b.users.GetByID(ctx, task.OwnerID)
			if err == nil {
				_ = b.reply(
					owner.TelegramID,
					fmt.Sprintf(
						"📣 %s تسک «%s» را انجام داد.",
						q.From.FirstName,
						task.Title,
					),
					MainKeyboard(),
				)
			}
		}

		return nil

	// ===========================
	// انجام نشد
	// ===========================
	case strings.HasPrefix(data, "task:cancel:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:cancel:"))
		if err != nil {
			return err
		}

		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			return err
		}

		_, _ = b.api.Request(
			tgbotapi.NewCallback(q.ID, "ثبت شد ❌"),
		)
		return b.reply(
			q.Message.Chat.ID,
			fmt.Sprintf("❌ تسک «%s» انجام نشد.", task.Title),
			MainKeyboard(),
		)

	// ===========================
	// ویرایش عنوان
	// ===========================
	case strings.HasPrefix(data, "task:edit:title:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:edit:title:"))
		if err != nil {
			return err
		}

		u := &domain.User{
			TelegramID: q.From.ID,
		}

		if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
			return err
		}

		session := domain.BotSession{
			UserID:    u.ID,
			State:     "waiting_task_title",
			Data:      id.String(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}

		if err := b.db.WithContext(ctx).
			Where("user_id = ?", u.ID).
			Assign(session).
			FirstOrCreate(&session).Error; err != nil {
			return err
		}

		_, _ = b.api.Request(
			tgbotapi.NewCallback(q.ID, "عنوان جدید را ارسال کنید ✏️"),
		)

		return b.reply(
			q.Message.Chat.ID,
			"📝 لطفاً عنوان جدید تسک را ارسال کنید.",
			MainKeyboard(),
		)

	// ===========================
	// ویرایش توضیحات
	// ===========================
	case strings.HasPrefix(data, "task:edit:desc:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:edit:desc:"))
		if err != nil {
			return err
		}

		u := &domain.User{
			TelegramID: q.From.ID,
		}

		if err := b.users.UpsertTelegramUser(ctx, u); err != nil {
			return err
		}

		session := domain.BotSession{
			UserID:    u.ID,
			State:     "waiting_task_description",
			Data:      id.String(),
			ExpiresAt: time.Now().Add(30 * time.Minute),
		}

		if err := b.db.WithContext(ctx).
			Where("user_id = ?", u.ID).
			Assign(session).
			FirstOrCreate(&session).Error; err != nil {
			return err
		}

		_, _ = b.api.Request(
			tgbotapi.NewCallback(q.ID, "توضیحات جدید را ارسال کنید ✏️"),
		)

		return b.reply(
			q.Message.Chat.ID,
			"📄 لطفاً توضیحات جدید تسک را ارسال کنید.",
			MainKeyboard(),
		)

		// ===========================
		// حذف
		// ===========================
	case strings.HasPrefix(data, "task:delete:"):
		id, err := uuid.Parse(strings.TrimPrefix(data, "task:delete:"))
		if err != nil {
			return err
		}

		if err := b.tasks.Delete(ctx, id); err != nil {
			return err
		}

		_, _ = b.api.Request(
			tgbotapi.NewCallback(q.ID, "حذف شد 🗑"),
		)

		return b.reply(
			q.Message.Chat.ID,
			"🗑 تسک با موفقیت حذف شد.",
			MainKeyboard(),
		)
	case strings.HasPrefix(data, "state:cancel"):
		user, err := b.users.GetByTelegramID(ctx, q.From.ID)
		if err != nil {
			return err
		}
		b.deleteState(ctx, user.ID)
		_, _ = b.api.Request(
			tgbotapi.NewCallback(q.ID, "لغو شد ❌"),
		)
		return b.reply(
			q.Message.Chat.ID,
			"❌ عملیات لغو شد.",
			MainKeyboard(),
		)

	case strings.HasPrefix(data, "user:filter:fr:"):
		username := strings.TrimPrefix(data, "user:filter:fr:")
		user, err := b.users.GetByTelegramID(ctx, q.From.ID)

		targetUser, err := b.users.GetByUsername(ctx, username)
		if err != nil {
			return b.reply(user.TelegramID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", MainKeyboard())
		}

		b.deleteState(ctx,user.ID)
		return b.getTargetTask(ctx, user.ID, targetUser.ID)
	}
	return nil
}
