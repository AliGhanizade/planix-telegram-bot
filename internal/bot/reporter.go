package bot

import (
	"context"
	"fmt"
	"math"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// cbList handles list:* callbacks (task list navigation).
func (b *Bot) cbList(ctx context.Context, q *models.CallbackQuery) {
	// format: list:tasks:<filter>:<page>
	parts := ui.SplitCallback(q.Data)
	if len(parts) < 4 {
		b.answer(ctx, q, "")
		return
	}
	f := ui.ParseListFilter(parts[2])
	page := ui.AtoiOr(parts[3], 1)

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

// renderTaskList renders the user task list with filter and pagination.
// when messageID is greater than zero the same message is edited in place (ui sync).
func (b *Bot) renderTaskList(ctx context.Context, chatID int64, messageID int, userID uuid.UUID, f ui.ListFilter, page int) error {
	tasks, total, err := b.tasks.ListFiltered(ctx, userID, string(f), page, ui.PageSize)
	if err != nil {
		return err
	}

	pages := int(math.Ceil(float64(total) / float64(ui.PageSize)))
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}

	text := fmt.Sprintf("📋 %s — صفحه‌ی %d از %d\n\n", f.Label(), page, pages)
	if len(tasks) == 0 {
		text += f.EmptyText()
	} else {
		for i, t := range tasks {
			text += fmt.Sprintf("%d. %s\n", (page-1)*ui.PageSize+i+1, ui.FormatSmallInfo(&t))
		}
	}

	markup := ui.TaskListKeyboard(tasks, f, page, pages)
	return b.render(ctx, chatID, messageID, text, markup)
}

// renderTaskCard renders the full task card with management buttons.
func (b *Bot) renderTaskCard(ctx context.Context, chatID int64, messageID int, taskID uuid.UUID, o ui.TaskOrigin) error {
	task, err := b.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	text := "🗂 کارت تسک\n\n" + ui.FormatTask(task)
	return b.render(ctx, chatID, messageID, text, ui.TaskCardKeyboard(task, o))
}

// notifyOwner tells the owner when someone works on their delegated task.
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

// setTaskForOther delegates several tasks to another user and notifies both sides.
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
