package bot

import (
	"context"
	"fmt"
	"math"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// pageSize تعداد تسک در هر صفحه‌ی فهرست است.
const pageSize = 5

// listFilter فیلتر وضعیت فهرست تسک‌ها.
type listFilter string

const (
	filterPending   listFilter = "pending"
	filterCompleted listFilter = "completed"
	filterAll       listFilter = "all"
)

// label عنوان فارسی فیلتر.
func (f listFilter) label() string {
	switch f {
	case filterCompleted:
		return "انجام‌شده‌ها"
	case filterAll:
		return "همه‌ی تسک‌ها"
	default:
		return "برنامه امروز"
	}
}

// emptyText پیام حالت خالی فیلتر.
func (f listFilter) emptyText() string {
	switch f {
	case filterCompleted:
		return "هنوز تسکی را تمام نکرده‌ای؛ یک شروع خوب، همین حالاست 💪"
	case filterAll:
		return "هنوز تسکی نساخته‌ای. با ➕ تسک جدید شروع کن!"
	default:
		return "امروز تسک بازی نداری؛ وقت یک شروع تازه است ✨"
	}
}

// parseListFilter فیلتر را از دیتای کال‌بک می‌خواند.
func parseListFilter(s string) listFilter {
	switch s {
	case string(filterCompleted):
		return filterCompleted
	case string(filterAll):
		return filterAll
	default:
		return filterPending
	}
}

// cbList کال‌بک‌های list:* (ناوبری فهرست تسک‌ها) را پردازش می‌کند.
func (b *Bot) cbList(ctx context.Context, q *models.CallbackQuery) {
	// فرمت: list:tasks:<filter>:<page>
	parts := splitCallback(q.Data)
	if len(parts) < 4 {
		b.answer(ctx, q, "")
		return
	}
	f := parseListFilter(parts[2])
	page := atoiOr(parts[3], 1)

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, "خطا! دوباره تلاش کن")
		return
	}

	b.answer(ctx, q, "")
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = u.TelegramID
		messageID = 0
	}
	if err := b.renderTaskList(ctx, chatID, messageID, u.ID, f, page); err != nil {
		b.log.Error("render task list failed", zap.Error(err))
	}
}

// renderTaskList فهرست تسک‌های کاربر را با فیلتر و صفحه‌بندی رندر می‌کند.
// اگر messageID بزرگ‌تر از صفر باشد همان پیام درجا ویرایش می‌شود (سینک UI).
func (b *Bot) renderTaskList(ctx context.Context, chatID int64, messageID int, userID uuid.UUID, f listFilter, page int) error {
	tasks, total, err := b.tasks.ListFiltered(ctx, userID, string(f), page, pageSize)
	if err != nil {
		return err
	}

	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}

	text := fmt.Sprintf("📋 %s — صفحه‌ی %d از %d\n\n", f.label(), page, pages)
	if len(tasks) == 0 {
		text += f.emptyText()
	} else {
		for i, t := range tasks {
			text += fmt.Sprintf("%d. %s\n", (page-1)*pageSize+i+1, FormatSmallInfo(&t))
		}
	}

	markup := TaskListKeyboard(tasks, f, page, pages)
	return b.render(ctx, chatID, messageID, text, markup)
}

// renderTaskCard کارت کامل تسک را با دکمه‌های مدیریتی رندر می‌کند.
func (b *Bot) renderTaskCard(ctx context.Context, chatID int64, messageID int, taskID uuid.UUID, o taskOrigin) error {
	task, err := b.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	text := "🗂 کارت تسک\n\n" + FormatTask(task)
	return b.render(ctx, chatID, messageID, text, TaskCardKeyboard(task, o))
}

// notifyOwner وقتی مجری تسکِ واگذارشده را انجام می‌دهد، مالک را خبر می‌کند.
func (b *Bot) notifyOwner(ctx context.Context, task *domain.Task, actor string) {
	if task.OwnerID == task.AssigneeID {
		return
	}
	owner, err := b.users.GetByID(ctx, task.OwnerID)
	if err != nil {
		b.log.Warn("fetch owner failed", zap.Error(err))
		return
	}
	verb := "انجام داد"
	if task.Status != "completed" {
		verb = "بازگشایی کرد"
	}
	if _, err := b.sendWithKeyboard(ctx, owner.TelegramID, fmt.Sprintf("📣 %s تسک «%s» را %s.", actor, task.Title, verb)); err != nil {
		b.log.Warn("notify owner failed", zap.Error(err))
	}
}

// setTaskForOther چند تسک را به کاربر دیگر واگذار می‌کند و هر دو طرف را خبر می‌کند.
func (b *Bot) setTaskForOther(ctx context.Context, ownerID, assigneeID uuid.UUID, titles []string, priority string) error {
	for _, title := range titles {
		task := &domain.Task{OwnerID: ownerID, AssigneeID: assigneeID, Title: title, Priority: priority, Status: "pending"}
		if err := b.tasks.CreateForOtherUser(ctx, task); err != nil {
			return err
		}
	}
	assignee, err := b.users.GetByID(ctx, assigneeID)
	if err != nil {
		return err
	}
	owner, err := b.users.GetByID(ctx, ownerID)
	if err != nil {
		return err
	}
	if _, err := b.sendWithKeyboard(ctx, assignee.TelegramID,
		fmt.Sprintf("📣 گزارش پلنیکس\n%s %d تسک برای تو ثبت کرد:", owner.FirstName, len(titles))); err != nil {
		return err
	}
	_, _ = b.sendWithKeyboard(ctx, owner.TelegramID,
		fmt.Sprintf("ثبت %d تسک برای %s با موفقیت انجام شد ✅", len(titles), assignee.FirstName))
	return nil
}
