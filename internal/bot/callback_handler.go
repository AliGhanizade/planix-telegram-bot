package bot

import (
	"context"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// onCallback routes inline button callbacks to their handlers.
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

// cbTask handles task:* callbacks.
func (b *Bot) cbTask(ctx context.Context, q *models.CallbackQuery) {
	parts := ui.SplitCallback(q.Data)
	if len(parts) < 3 {
		b.answer(ctx, q, "")
		return
	}

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, ui.ErrGeneric(b.lang(u)))
		return
	}
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = u.TelegramID
		messageID = 0
	}

	l := b.lang(u)

	switch parts[1] {
	case "due":
		// task:due:<id>:<preset>[:origin]
		if len(parts) < 5 {
			b.answer(ctx, q, "")
			return
		}
		b.cbTaskDue(ctx, q, parts[2], parts[3], chatID, messageID, l)

	case "prio":
		// task:prio:<id>:<level>[:origin]
		if len(parts) < 5 {
			b.answer(ctx, q, "")
			return
		}
		b.cbTaskPriority(ctx, q, parts[2], parts[3], chatID, messageID, l)

	case "proof":
		// task:proof:<id>[:origin]
		b.cbTaskProof(ctx, q, u, parts[2], chatID, messageID, l)

	case "viewproof":
		// task:viewproof:<id>[:origin]
		b.cbTaskViewProof(ctx, q, parts[2], chatID, messageID, l)

	case "evidence":
		// task:evidence:<id>[:origin]
		b.cbTaskEvidenceToggle(ctx, q, parts[2], chatID, messageID, l)

	case "edit":
		// task:edit:<what>:<id>[:origin]
		if len(parts) < 4 {
			b.answer(ctx, q, "")
			return
		}
		b.cbTaskEdit(ctx, q, u, parts[2], parts[3], chatID, messageID, l)

	case "delete":
		if parts[2] == "yes" {
			// task:delete:yes:<id>[:origin]
			if len(parts) < 4 {
				b.answer(ctx, q, "")
				return
			}
			b.cbTaskDeleteYes(ctx, q, u, parts[3], chatID, messageID, l)
			return
		}
		// task:delete:<id>[:origin]
		b.cbTaskDeleteAsk(ctx, q, parts[2], chatID, messageID, l)

	default:
		// simple actions: task:<action>:<id>[:origin]
		b.cbTaskAction(ctx, q, u, parts[1], parts[2], chatID, messageID, l)
	}
}

// cbTaskAction handles simple task actions: info, refresh, done and reopen.
func (b *Bot) cbTaskAction(ctx context.Context, q *models.CallbackQuery, u *domain.User, action, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 3)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	switch action {
	case "info":
		b.answer(ctx, q, "")
		if err := b.renderTaskCard(ctx, chatID, messageID, id, origin, l); err != nil {
			b.log.Error("render task card failed", zap.Error(err))
			b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
		}

	case "refresh":
		b.answer(ctx, q, ui.RefreshedToast(l))
		b.refreshTaskView(ctx, u.ID, chatID, messageID, id, origin, l)

	case "done":
		task, err := b.tasks.Complete(ctx, id)
		if err != nil {
			b.log.Error("complete task failed", zap.Error(err))
			b.answerAlert(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.TaskDoneToast(l))
		b.notifyOwner(ctx, task, q.From.FirstName)
		b.refreshTaskView(ctx, u.ID, chatID, messageID, id, origin, l)

	case "reopen":
		task, err := b.tasks.Reopen(ctx, id)
		if err != nil {
			b.log.Error("reopen task failed", zap.Error(err))
			b.answerAlert(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.TaskReopenedToast(l))
		b.notifyOwner(ctx, task, q.From.FirstName)
		b.refreshTaskView(ctx, u.ID, chatID, messageID, id, origin, l)

	default:
		b.answer(ctx, q, "")
	}
}

// refreshTaskView re-renders the list or the card depending on origin to keep the ui in sync.
func (b *Bot) refreshTaskView(ctx context.Context, userID uuid.UUID, chatID int64, messageID int, id uuid.UUID, origin ui.TaskOrigin, l ui.Lang) {
	if origin.List {
		if err := b.renderTaskList(ctx, chatID, messageID, userID, ui.ListFilter(origin.Filter), origin.Page, l); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}
		return
	}
	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin, l); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}

// cbTaskEdit starts edit flows (title, description) and opens pickers (due, priority).
func (b *Bot) cbTaskEdit(ctx context.Context, q *models.CallbackQuery, u *domain.User, what, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	switch what {
	case "title":
		if err := b.startEditSession(ctx, u.ID, stateWaitingTaskTitle, id, cardBackRef(chatID, messageID, origin)); err != nil {
			b.log.Error("start edit session failed", zap.Error(err))
			b.answer(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.EditTitleToast(l))
		if err := b.edit(ctx, chatID, messageID, ui.EditTitlePrompt(l), ui.CancelInlineKeyboard(l)); err != nil {
			b.log.Error("render edit title prompt failed", zap.Error(err))
		}

	case "desc":
		if err := b.startEditSession(ctx, u.ID, stateWaitingTaskDescription, id, cardBackRef(chatID, messageID, origin)); err != nil {
			b.log.Error("start edit session failed", zap.Error(err))
			b.answer(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.EditDescToast(l))
		if err := b.edit(ctx, chatID, messageID, ui.EditDescPrompt(l), ui.CancelInlineKeyboard(l)); err != nil {
			b.log.Error("render edit desc prompt failed", zap.Error(err))
		}

	case "due":
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
			return
		}
		b.answer(ctx, q, "")
		text := ui.DuePickerText(ui.Truncate(task.Title, 40), l)
		if err := b.edit(ctx, chatID, messageID, text, ui.DuePickerKeyboard(task, origin, l)); err != nil {
			b.log.Error("render due picker failed", zap.Error(err))
		}

	case "priority":
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
			return
		}
		b.answer(ctx, q, "")
		text := ui.PriorityPickerText(ui.Truncate(task.Title, 40), l)
		if err := b.edit(ctx, chatID, messageID, text, ui.PriorityPickerKeyboard(task, origin, l)); err != nil {
			b.log.Error("render priority picker failed", zap.Error(err))
		}

	default:
		b.answer(ctx, q, "")
	}
}

// cbTaskDeleteAsk shows the two step delete confirmation.
func (b *Bot) cbTaskDeleteAsk(ctx context.Context, q *models.CallbackQuery, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 3)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	task, err := b.tasks.GetByID(ctx, id)
	if err != nil {
		b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
		return
	}
	b.answer(ctx, q, "")
	text := ui.DeleteConfirmText(ui.Truncate(task.Title, 60), l)
	if err := b.edit(ctx, chatID, messageID, text, ui.DeleteConfirmKeyboard(task, origin, l)); err != nil {
		b.log.Error("render delete confirm failed", zap.Error(err))
	}
}

// cbTaskDeleteYes deletes the task after confirmation.
func (b *Bot) cbTaskDeleteYes(ctx context.Context, q *models.CallbackQuery, u *domain.User, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	if err := b.tasks.Delete(ctx, id); err != nil {
		b.log.Error("delete task failed", zap.Error(err))
		b.answerAlert(ctx, q, ui.ErrGeneric(l))
		return
	}
	b.answer(ctx, q, ui.DeletedToast(l))

	if origin.List {
		if err := b.renderTaskList(ctx, chatID, messageID, u.ID, ui.ListFilter(origin.Filter), origin.Page, l); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}
		return
	}
	if err := b.edit(ctx, chatID, messageID, ui.DeletedMessage(l), ui.BackKeyboard(l)); err != nil {
		b.log.Error("render deleted notice failed", zap.Error(err))
	}
}

// cbTaskDue saves the due date picked from the quick picker.
func (b *Bot) cbTaskDue(ctx context.Context, q *models.CallbackQuery, rawID, preset string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	if preset == "none" {
		if err := b.tasks.UpdateDueAt(ctx, id, nil); err != nil {
			b.log.Error("clear due date failed", zap.Error(err))
			b.answerAlert(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.DueClearedToast(l))
	} else {
		task, err := b.tasks.GetByID(ctx, id)
		if err != nil {
			b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
			return
		}
		assignee, err := b.users.GetByID(ctx, task.AssigneeID)
		if err != nil {
			b.answerAlert(ctx, q, ui.ErrGeneric(l))
			return
		}
		due, ok := dueFromPreset(preset, time.Now(), userLocation(assignee.Timezone))
		if !ok {
			b.answer(ctx, q, "")
			return
		}
		if err := b.tasks.UpdateDueAt(ctx, id, &due); err != nil {
			b.log.Error("update due date failed", zap.Error(err))
			b.answerAlert(ctx, q, ui.ErrGeneric(l))
			return
		}
		b.answer(ctx, q, ui.DueSetToast(due.Format("01-02 15:04"), l))
	}

	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin, l); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}

// cbTaskPriority saves the picked priority.
func (b *Bot) cbTaskPriority(ctx context.Context, q *models.CallbackQuery, rawID, level string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 4)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}

	if err := b.tasks.UpdatePriority(ctx, id, level); err != nil {
		b.log.Error("update priority failed", zap.Error(err))
		b.answerAlert(ctx, q, ui.ErrGeneric(l))
		return
	}
	b.answer(ctx, q, ui.PrioritySetToast(ui.PriorityLabel(level, l), l))

	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin, l); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}

// cbTaskProof starts the photo proof flow for a task.
func (b *Bot) cbTaskProof(ctx context.Context, q *models.CallbackQuery, u *domain.User, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 3)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}
	if err := b.setSession(ctx, u.ID, stateWaitingTaskProof, sessionData{
		TaskID: id.String(),
		Back:   cardBackRef(chatID, messageID, origin),
	}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
		b.answer(ctx, q, ui.ErrGeneric(l))
		return
	}
	b.answer(ctx, q, "")
	if err := b.edit(ctx, chatID, messageID, ui.PhotoProofPrompt(l), ui.CancelInlineKeyboard(l)); err != nil {
		b.log.Error("render proof prompt failed", zap.Error(err))
	}
}

// cbTaskViewProof re-sends the stored proof photo of a task.
func (b *Bot) cbTaskViewProof(ctx context.Context, q *models.CallbackQuery, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	evidence, err := b.tasks.LatestEvidence(ctx, id)
	if err != nil || evidence.TelegramFileID == "" {
		b.answerAlert(ctx, q, ui.ProofMissingToast(l))
		return
	}
	task, err := b.tasks.GetByID(ctx, id)
	if err != nil {
		b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
		return
	}
	b.answer(ctx, q, "")
	_, err = b.api.SendPhoto(ctx, &tgbot.SendPhotoParams{
		ChatID:  chatID,
		Photo:   &models.InputFileString{Data: evidence.TelegramFileID},
		Caption: ui.PhotoCaption(task.Title, l),
	})
	if err != nil {
		b.log.Error("send proof photo failed", zap.Error(err))
	}
}

// cbTaskEvidenceToggle turns the proof requirement on or off.
func (b *Bot) cbTaskEvidenceToggle(ctx context.Context, q *models.CallbackQuery, rawID string, chatID int64, messageID int, l ui.Lang) {
	id, err := uuid.Parse(rawID)
	if err != nil {
		b.answerAlert(ctx, q, ui.InvalidTaskIDToast(l))
		return
	}
	origin, err := ui.ParseTaskOrigin(ui.SplitCallback(q.Data), 3)
	if err != nil {
		b.answer(ctx, q, "")
		return
	}
	task, err := b.tasks.GetByID(ctx, id)
	if err != nil {
		b.answerAlert(ctx, q, ui.TaskNotFoundToast(l))
		return
	}
	newValue := !task.RequiresEvidence
	if err := b.tasks.SetEvidenceRequired(ctx, id, newValue); err != nil {
		b.log.Error("set evidence failed", zap.Error(err))
		b.answer(ctx, q, ui.ErrGeneric(l))
		return
	}
	b.answer(ctx, q, ui.EvidenceToast(newValue, l))
	if err := b.renderTaskCard(ctx, chatID, messageID, id, origin, l); err != nil {
		b.log.Error("render task card failed", zap.Error(err))
	}
}
