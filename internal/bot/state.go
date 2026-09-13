package bot

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// conversation state names of the bot.
const (
	stateWaitingTaskTitle       = "waiting_task_title"
	stateWaitingTaskDescription = "waiting_task_description"
	stateWaitingTaskForOther    = "waiting_task_for_other"
	stateWaitingStatusForOther  = "waiting_task_status_for_other"
	stateWaitingSearch          = "waiting_search"

	stateWaitingEditFirstName = "waiting_edit_first_name"
	stateWaitingEditLastName  = "waiting_edit_last_name"
	stateWaitingEditTimezone  = "waiting_edit_timezone"

	stateWaitingTaskProof = "waiting_task_proof"
)

// backRef is where to re-render after a flow ends; keeps the ui in sync.
type backRef struct {
	Kind      string `json:"kind"` // card | list
	Filter    string `json:"filter,omitempty"`
	Page      int    `json:"page,omitempty"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
}

// sessionData is the json payload stored in a user session.
type sessionData struct {
	TaskID string   `json:"task_id,omitempty"`
	Back   *backRef `json:"back,omitempty"`
}

// cardBackRef converts a card origin into a backRef.
func cardBackRef(chatID int64, messageID int, o ui.TaskOrigin) *backRef {
	ref := &backRef{Kind: "card", ChatID: chatID, MessageID: messageID}
	if o.List {
		ref.Filter = o.Filter
		ref.Page = o.Page
	}
	return ref
}

// origin converts the stored ref into a ui.TaskOrigin.
func (r *backRef) origin() ui.TaskOrigin {
	if r == nil || r.Filter == "" {
		return ui.NoOrigin
	}
	return ui.TaskOrigin{List: true, Filter: r.Filter, Page: r.Page}
}

// setSession stores the conversation state and its data for 30 minutes.
func (b *Bot) setSession(ctx context.Context, userID uuid.UUID, state string, data sessionData) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s := domain.BotSession{UserID: userID, State: state, Data: string(raw), ExpiresAt: time.Now().Add(30 * time.Minute)}
	return b.db.WithContext(ctx).Where("user_id = ?", userID).Assign(s).FirstOrCreate(&s).Error
}

// activeSession returns the user's latest valid session.
func (b *Bot) activeSession(ctx context.Context, userID uuid.UUID) (domain.BotSession, bool, error) {
	var session domain.BotSession
	err := b.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("updated_at DESC").
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.BotSession{}, false, nil
	}
	if err != nil {
		return domain.BotSession{}, false, err
	}
	return session, true, nil
}

// clearSession removes the active session of a user.
func (b *Bot) clearSession(ctx context.Context, userID uuid.UUID) error {
	return b.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.BotSession{}).Error
}

// checkState processes free text based on the current conversation state.
func (b *Bot) checkState(ctx context.Context, u *domain.User, text string, chatID int64) error {
	session, ok, err := b.activeSession(ctx, u.ID)
	if err != nil || !ok {
		return err
	}

	switch session.State {
	case stateWaitingTaskTitle:
		return b.handleTitleInput(ctx, u, session, text)
	case stateWaitingTaskDescription:
		return b.handleDescriptionInput(ctx, u, session, text)
	case stateWaitingTaskForOther:
		return b.handleAssignInput(ctx, u, text, chatID)
	case stateWaitingStatusForOther:
		return b.handleStatusInput(ctx, u, text, chatID)
	case stateWaitingSearch:
		return b.handleSearchInput(ctx, u, session, text, chatID)
	case stateWaitingEditFirstName, stateWaitingEditLastName, stateWaitingEditTimezone:
		return b.handleProfileEditInput(ctx, u, session, text)
	}
	return nil
}

// onPhoto processes photo messages, currently as task proof.
func (b *Bot) onPhoto(ctx context.Context, m *models.Message) {
	if m.From == nil || m.From.IsBot {
		return
	}
	u, err := b.upsertUser(ctx, *m.From)
	if err != nil {
		b.log.Error("upsert user failed", zap.Error(err))
		return
	}
	session, ok, err := b.activeSession(ctx, u.ID)
	if err != nil || !ok || session.State != stateWaitingTaskProof {
		_, _ = b.send(ctx, m.Chat.ID, ui.NoTaskForPhoto(b.lang(u)), ui.MainKeyboard(b.lang(u)))
		return
	}
	if len(m.Photo) == 0 {
		return
	}
	// the last photo size is the largest one
	fileID := m.Photo[len(m.Photo)-1].FileID

	var data sessionData
	_ = json.Unmarshal([]byte(session.Data), &data)
	taskID, err := uuid.Parse(data.TaskID)
	if err != nil {
		_ = b.clearSession(ctx, u.ID)
		return
	}
	if err := b.tasks.SavePhotoEvidence(ctx, taskID, u.ID, fileID); err != nil {
		b.log.Error("save photo evidence failed", zap.Error(err))
		return
	}
	if err := b.clearSession(ctx, u.ID); err != nil {
		return
	}
	_, _ = b.send(ctx, m.Chat.ID, ui.PhotoSaved(b.lang(u)), ui.MainKeyboard(b.lang(u)))
	if data.Back != nil {
		b.renderTaskCard(ctx, data.Back.ChatID, data.Back.MessageID, taskID, data.Back.origin(), b.lang(u))
	}
}

// handleProfileEditInput saves the new profile field value and re-renders the edit menu.
func (b *Bot) handleProfileEditInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	var data sessionData
	_ = json.Unmarshal([]byte(session.Data), &data)

	var firstName, lastName, timezone *string
	value := strings.TrimSpace(text)
	switch session.State {
	case stateWaitingEditFirstName:
		firstName = &value
	case stateWaitingEditLastName:
		if value == "-" {
			empty := ""
			lastName = &empty
		} else {
			lastName = &value
		}
	case stateWaitingEditTimezone:
		timezone = &value
	}

	if _, err := b.profiles.UpdateProfile(ctx, u.ID, firstName, lastName, timezone); err != nil {
		_, rerr := b.send(ctx, u.TelegramID, ui.ValidationError(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u)))
		if rerr != nil {
			return err
		}
		// keep the session open so the user can send a valid value.
		return nil
	}

	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}

	// sync: the edit menu re-renders on the same message.
	if data.Back != nil {
		b.showProfileEdit(ctx, u, data.Back.ChatID, data.Back.MessageID)
		return nil
	}
	b.showProfileEdit(ctx, u, u.TelegramID, 0)
	return nil
}

// handleTitleInput creates a task or edits the title of an existing one.
func (b *Bot) handleTitleInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	var data sessionData
	_ = json.Unmarshal([]byte(session.Data), &data)

	// edit mode: the session holds a task id.
	if data.TaskID != "" {
		id, err := uuid.Parse(data.TaskID)
		if err != nil {
			return err
		}
		if err := b.tasks.UpdateTitle(ctx, id, text); err != nil {
			return err
		}
		if err := b.clearSession(ctx, u.ID); err != nil {
			return err
		}
		if data.Back != nil {
			return b.renderTaskCard(ctx, data.Back.ChatID, data.Back.MessageID, id, data.Back.origin(), b.lang(u))
		}
		return nil
	}

	// create new task mode.
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	task := &domain.Task{OwnerID: u.ID, AssigneeID: u.ID, Title: text, Priority: "normal", Status: "pending"}
	if err := b.tasks.Create(ctx, task); err != nil {
		return err
	}
	_, err := b.send(ctx, u.TelegramID,
		ui.TaskSaved(task.Title, b.lang(u)),
		ui.TodayInlineKeyboard(b.lang(u)))
	return err
}

// handleDescriptionInput saves the description of the task being edited.
func (b *Bot) handleDescriptionInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	var data sessionData
	if err := json.Unmarshal([]byte(session.Data), &data); err != nil || data.TaskID == "" {
		return b.clearSession(ctx, u.ID)
	}
	id, err := uuid.Parse(data.TaskID)
	if err != nil {
		return err
	}
	if err := b.tasks.UpdateDescription(ctx, id, text); err != nil {
		return err
	}
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	if data.Back != nil {
		return b.renderTaskCard(ctx, data.Back.ChatID, data.Back.MessageID, id, data.Back.origin(), b.lang(u))
	}
	return nil
}

// handleSearchInput runs the search and shows the results.
func (b *Bot) handleSearchInput(ctx context.Context, u *domain.User, session domain.BotSession, text string, chatID int64) error {
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	tasks, err := b.tasks.Search(ctx, u.ID, text, 10)
	if err != nil {
		return err
	}

	header := ui.SearchResultsHeader(text, b.lang(u))
	if len(tasks) == 0 {
		header += ui.SearchEmpty(b.lang(u))
	} else {
		for _, t := range tasks {
			header += ui.FormatSmallInfo(&t, b.lang(u)) + "\n"
		}
	}
	_, err = b.send(ctx, chatID, header, ui.SearchResultsKeyboard(tasks, b.lang(u)))
	return err
}

// handleAssignInput parses the assignment input and saves the tasks.
func (b *Bot) handleAssignInput(ctx context.Context, u *domain.User, text string, chatID int64) error {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		_, err := b.send(ctx, chatID, ui.AssignFormatError(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u)))
		return err
	}
	username := strings.TrimPrefix(strings.TrimSpace(lines[0]), "@")
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		_, err := b.send(ctx, chatID, ui.UsernameNotFound(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u)))
		return err
	}
	if target.ID == u.ID {
		_, err := b.send(ctx, chatID, ui.SelfAssignError(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u)))
		return err
	}
	if err := b.setTaskForOther(ctx, u.ID, target.ID, lines[1:], "normal", b.lang(u)); err != nil {
		return err
	}
	return b.clearSession(ctx, u.ID)
}

// handleStatusInput shows the status of tasks delegated to the target user.
func (b *Bot) handleStatusInput(ctx context.Context, u *domain.User, text string, chatID int64) error {
	username := strings.TrimPrefix(strings.TrimSpace(text), "@")
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		_, err := b.send(ctx, chatID, ui.UsernameNotFound(b.lang(u)), ui.CancelInlineKeyboard(b.lang(u)))
		return err
	}
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	return b.renderDelegatedStatus(ctx, chatID, 0, u.ID, target.ID, b.lang(u))
}

// startEditSession prepares a session for editing a task with its back target.
func (b *Bot) startEditSession(ctx context.Context, userID uuid.UUID, state string, taskID uuid.UUID, back *backRef) error {
	return b.setSession(ctx, userID, state, sessionData{TaskID: taskID.String(), Back: back})
}
