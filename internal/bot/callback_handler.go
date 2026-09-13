package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// splitCallback دیتای کال‌بک را به اجزا می‌شکند.
func splitCallback(data string) []string {
	return strings.Split(data, ":")
}

// onCallback کال‌بک‌های دکمه‌های شیشه‌ای را بین هندلرها پخش می‌کند.
func (b *Bot) onCallback(ctx context.Context, q *models.CallbackQuery) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("panic in callback handler", zap.Any("panic", r), zap.String("data", q.Data))
		}
	}()

	data := q.Data
	switch {
	case strings.HasPrefix(data, "state:"):
		b.cbState(ctx, q)
	case strings.HasPrefix(data, "nav:"):
		b.cbNav(ctx, q)
	case strings.HasPrefix(data, "list:"):
		b.cbList(ctx, q)
	case strings.HasPrefix(data, "task:"):
		b.cbTask(ctx, q)
	case strings.HasPrefix(data, "user:"):
		b.cbUserPick(ctx, q)
	case strings.HasPrefix(data, "settings:"):
		b.cbSettings(ctx, q)
	default:
		b.answer(ctx, q, "")
	}
}

// cbTask کال‌بک‌های عملیات روی یک تسک (task:*) را پردازش می‌کند.
func (b *Bot) cbTask(ctx context.Context, q *models.CallbackQuery) {
	parts := splitCallback(q.Data)
	if len(parts) < 3 {
		b.answer(ctx, q, "")
		return
	}

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, "خطا! دوباره تلاش کن")
		return
	}
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = u.TelegramID
		messageID = 0
	}

	switch parts[1] {
	case "due":
		// task:due:<id>:<preset>[:مبدأ]
		if len(parts) < 5 {
			b.answer(ctx, q, "")
			return
		}
		b.cbTaskDue(ctx, q, parts[2], parts[3], chatID, messageID)

	case "prio":
		// task:prio:<id>:<level>[:مبدأ]
		if len(parts) < 5 {
			b.answer(ctx, q, "")
			return
		}
		b.cbTaskPriority(ctx, q, parts[2], parts[3], chatID, messageID)

	case "edit":
		// task:edit:<what>:<id>[:مبدأ]
		if len(parts) < 4 {
			b.answer(ctx, q, "")
			return
		}
		b.cbTaskEdit(ctx, q, u, parts[2], parts[3], chatID, messageID)

	case "delete":
		if parts[2] == "yes" {
			// task:delete:yes:<id>[:مبدأ]
			if len(parts) < 4 {
				b.answer(ctx, q, "")
				return
			}
			b.cbTaskDeleteYes(ctx, q, u, parts[3], chatID, messageID)
			return
		}
		// task:delete:<id>[:مبدأ]
		b.cbTaskDeleteAsk(ctx, q, parts[2], chatID, messageID)

	default:
		// عملیات‌های ساده: task:<action>:<id>[:مبدأ]
		b.cbTaskAction(ctx, q, u, parts[1], parts[2], chatID, messageID)
	}
}

// cbTaskAction عملیات‌های ساده روی تسک: info، refresh، done و reopen.
func (b *Bot) cbTaskAction(ctx context.Context, q *models.CallbackQuery, u *domain.User, action, rawID string, chatID int64, messageID int) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, "شناسه‌ی تسک نامعتبر است")
		return
	}
	origin, err := parseTaskOrigin(splitCallback(q.Data), 3)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	switch action {
	case "info":
		b.answer(ctx, q, "")
		if err := b.renderTaskCard(ctx, chatID, messageID, id, origin); err != nil {
			b.log.Error("render task card failed", zap.Error(err))
			b.answerAlert(ctx, q, "تسک پیدا نشد")
		}

	case "refresh":
		b.answer(ctx, q, "بروزرسانی شد 🔄")
		b.refreshTaskView(ctx, u.ID, chatID, messageID, id, origin)

	case "done":
		task, err := b.tasks.Complete(ctx, id)
		if err != nil {
			b.log.Error("complete task failed", zap.Error(err))
			b.answerAlert(ctx, q, "خطا در انجام تسک")
			return
		}
		b.answer(ctx, q, "تسک انجام شد ✅")
		b.notifyOwner(ctx, task, q.From.FirstName)
		b.refreshTaskView(ctx, u.ID, chatID, messageID, id, origin)

	case "reopen":
		task, err := b.tasks.Reopen(ctx, id)
		if err != nil {
			b.log.Error("reopen task failed", zap.Error(err))
			b.answerAlert(ctx, q, "خطا در بازگشایی تسک")
			return
		}
		b.answer(ctx, q, "تسک دوباره باز شد 🔄")
		b.notifyOwner(ctx, task, q.From.FirstName)
		b.refreshTaskView(ctx, u.ID, chatID, messageID, id, origin)

	default:
		b.answer(ctx, q, "")
	}
}

// refreshTaskView بسته به مبدأ، فهرست یا کارت را دوباره رندر می‌کند تا UI سینک بماند.
func (b *Bot) refreshTaskView(ctx context.Context, userID uuid.UUID, chatID int64, messageID int, id uuid.UUID, origin taskOrigin) {
	if origin.List {
		if err := b.renderTaskList(ctx, chatID, messageID, userID, listFilter(origin.Filter), origin.Page); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}
		return
	}
	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}

// cbTaskEdit شروع جریان‌های ویرایش (عنوان، توضیحات) و باز کردن انتخابگرها (موعد، اولویت).
func (b *Bot) cbTaskEdit(ctx context.Context, q *models.CallbackQuery, u *domain.User, what, rawID string, chatID int64, messageID int) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, "شناسه‌ی تسک نامعتبر است")
		return
	}
	origin, err := parseTaskOrigin(splitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	switch what {
	case "title":
		if err := b.startEditSession(ctx, u.ID, stateWaitingTaskTitle, id, cardBackRef(chatID, messageID, origin)); err != nil {
			b.log.Error("start edit session failed", zap.Error(err))
			b.answer(ctx, q, "خطا! دوباره تلاش کن")
			return
		}
		b.answer(ctx, q, "عنوان جدید را ارسال کنید ✏️")
		if err := b.edit(ctx, chatID, messageID, "📝 لطفاً عنوان جدید تسک را در پیام بعدی بفرست.", CancelInlineKeyboard()); err != nil {
			b.log.Error("render edit title prompt failed", zap.Error(err))
		}

	case "desc":
		if err := b.startEditSession(ctx, u.ID, stateWaitingTaskDescription, id, cardBackRef(chatID, messageID, origin)); err != nil {
			b.log.Error("start edit session failed", zap.Error(err))
			b.answer(ctx, q, "خطا! دوباره تلاش کن")
			return
		}
		b.answer(ctx, q, "توضیحات جدید را ارسال کنید ✏️")
		if err := b.edit(ctx, chatID, messageID, "📄 لطفاً توضیحات جدید تسک را در پیام بعدی بفرست.", CancelInlineKeyboard()); err != nil {
			b.log.Error("render edit desc prompt failed", zap.Error(err))
		}

	case "due":
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			b.answerAlert(ctx, q, "تسک پیدا نشد")
			return
		}
		b.answer(ctx, q, "")
		text := fmt.Sprintf("📅 موعد تسک «%s» را انتخاب کن:", truncate(task.Title, 40))
		if err := b.edit(ctx, chatID, messageID, text, DuePickerKeyboard(task, origin)); err != nil {
			b.log.Error("render due picker failed", zap.Error(err))
		}

	case "priority":
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			b.answerAlert(ctx, q, "تسک پیدا نشد")
			return
		}
		b.answer(ctx, q, "")
		text := fmt.Sprintf("⚡ اولویت تسک «%s» را انتخاب کن:", truncate(task.Title, 40))
		if err := b.edit(ctx, chatID, messageID, text, PriorityPickerKeyboard(task, origin)); err != nil {
			b.log.Error("render priority picker failed", zap.Error(err))
		}

	default:
		b.answer(ctx, q, "")
	}
}

// cbTaskDeleteAsk نمایش تایید دو مرحله‌ای حذف.
func (b *Bot) cbTaskDeleteAsk(ctx context.Context, q *models.CallbackQuery, rawID string, chatID int64, messageID int) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, "شناسه‌ی تسک نامعتبر است")
		return
	}
	origin, err := parseTaskOrigin(splitCallback(q.Data), 3)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	task, err := b.tasks.GetByID(ctx, id)
	if err != nil {
		b.answerAlert(ctx, q, "تسک پیدا نشد")
		return
	}
	b.answer(ctx, q, "")
	text := fmt.Sprintf("🗑 مطمئنی می‌خوای تسک «%s» را حذف کنی؟\nاین عمل قابل بازگشت نیست.", truncate(task.Title, 60))
	if err := b.edit(ctx, chatID, messageID, text, DeleteConfirmKeyboard(task, origin)); err != nil {
		b.log.Error("render delete confirm failed", zap.Error(err))
	}
}

// cbTaskDeleteYes حذف نهایی تسک بعد از تایید.
func (b *Bot) cbTaskDeleteYes(ctx context.Context, q *models.CallbackQuery, u *domain.User, rawID string, chatID int64, messageID int) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, "شناسه‌ی تسک نامعتبر است")
		return
	}
	origin, err := parseTaskOrigin(splitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	if err := b.tasks.Delete(ctx, id); err != nil {
		b.log.Error("delete task failed", zap.Error(err))
		b.answerAlert(ctx, q, "خطا در حذف تسک")
		return
	}
	b.answer(ctx, q, "حذف شد 🗑")

	if origin.List {
		if err := b.renderTaskList(ctx, chatID, messageID, u.ID, listFilter(origin.Filter), origin.Page); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}
		return
	}
	if err := b.edit(ctx, chatID, messageID, "🗑 تسک حذف شد.", BackKeyboard()); err != nil {
		b.log.Error("render deleted notice failed", zap.Error(err))
	}
}

// cbTaskDue ثبت موعد انتخابی از انتخابگر سریع.
func (b *Bot) cbTaskDue(ctx context.Context, q *models.CallbackQuery, rawID, preset string, chatID int64, messageID int) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, "شناسه‌ی تسک نامعتبر است")
		return
	}
	origin, err := parseTaskOrigin(splitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	if preset == "none" {
		if err := b.tasks.UpdateDueAt(ctx, id, nil); err != nil {
			b.log.Error("clear due date failed", zap.Error(err))
			b.answerAlert(ctx, q, "خطا در ثبت موعد")
			return
		}
		b.answer(ctx, q, "موعد حذف شد 🗓")
	} else {
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			b.answerAlert(ctx, q, "تسک پیدا نشد")
			return
		}
		assignee, err := b.users.GetByID(ctx, task.AssigneeID)
		if err != nil {
			b.answerAlert(ctx, q, "خطا در ثبت موعد")
			return
		}
		due, ok := dueFromPreset(preset, time.Now(), userLocation(assignee.Timezone))
		if !ok {
			b.answer(ctx, q, "")
			return
		}
		if err := b.tasks.UpdateDueAt(ctx, id, &due); err != nil {
			b.log.Error("update due date failed", zap.Error(err))
			b.answerAlert(ctx, q, "خطا در ثبت موعد")
			return
		}
		b.answer(ctx, q, "موعد ثبت شد 📅 "+due.Format("01-02 15:04"))
	}

	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}

// cbTaskPriority ثبت اولویت انتخابی.
func (b *Bot) cbTaskPriority(ctx context.Context, q *models.CallbackQuery, rawID, level string, chatID int64, messageID int) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, "شناسه‌ی تسک نامعتبر است")
		return
	}
	origin, err := parseTaskOrigin(splitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	if err := b.tasks.UpdatePriority(ctx, id, level); err != nil {
		b.log.Error("update priority failed", zap.Error(err))
		b.answerAlert(ctx, q, "خطا در ثبت اولویت")
		return
	}
	b.answer(ctx, q, "اولویت ثبت شد ⚡ "+priorityLabel(level))

	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}

// atoiOr صفحات را از دیتای کال‌بک می‌خواند.
func atoiOr(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return fallback
}
