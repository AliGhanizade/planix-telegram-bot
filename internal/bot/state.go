package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/bot/ui"
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// نام وضعیت‌های مکالمه‌ی چندمرحله‌ای بات.
const (
	stateWaitingTaskTitle       = "waiting_task_title"
	stateWaitingTaskDescription = "waiting_task_description"
	stateWaitingTaskForOther    = "waiting_task_for_other"
	stateWaitingStatusForOther  = "waiting_task_status_for_other"
	stateWaitingSearch          = "waiting_search"

	stateWaitingEditFirstName = "waiting_edit_first_name"
	stateWaitingEditLastName  = "waiting_edit_last_name"
	stateWaitingEditTimezone  = "waiting_edit_timezone"
)

// backRef مقصد بازگشت بعد از پایان یک جریان؛ برای سینک رابط کاربری.
type backRef struct {
	Kind      string `json:"kind"` // card | list
	Filter    string `json:"filter,omitempty"`
	Page      int    `json:"page,omitempty"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
}

// sessionData داده‌ی JSON ذخیره‌شده در نشست کاربر.
type sessionData struct {
	TaskID string   `json:"task_id,omitempty"`
	Back   *backRef `json:"back,omitempty"`
}

// cardBackRef مبدأ کارت را به backRef تبدیل می‌کند.
func cardBackRef(chatID int64, messageID int, o ui.TaskOrigin) *backRef {
	ref := &backRef{Kind: "card", ChatID: chatID, MessageID: messageID}
	if o.List {
		ref.Filter = o.Filter
		ref.Page = o.Page
	}
	return ref
}

// origin مبدأ ذخیره‌شده را به ui.TaskOrigin تبدیل می‌کند.
func (r *backRef) origin() ui.TaskOrigin {
	if r == nil || r.Filter == "" {
		return ui.NoOrigin
	}
	return ui.TaskOrigin{List: true, Filter: r.Filter, Page: r.Page}
}

// setSession وضعیت مکالمه و داده‌ی آن را برای ۳۰ دقیقه ذخیره می‌کند.
func (b *Bot) setSession(ctx context.Context, userID uuid.UUID, state string, data sessionData) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s := domain.BotSession{UserID: userID, State: state, Data: string(raw), ExpiresAt: time.Now().Add(30 * time.Minute)}
	return b.db.WithContext(ctx).Where("user_id = ?", userID).Assign(s).FirstOrCreate(&s).Error
}

// activeSession آخرین نشست معتبر کاربر را برمی‌گرداند.
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

// clearSession نشست فعال کاربر را حذف می‌کند.
func (b *Bot) clearSession(ctx context.Context, userID uuid.UUID) error {
	return b.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.BotSession{}).Error
}

// checkState پیام آزاد کاربر را بر اساس وضعیت جاری مکالمه پردازش می‌کند.
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

// handleProfileEditInput مقدار جدید فیلد پروفایل را ذخیره و منوی ویرایش را دوباره رندر می‌کند.
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
		_, rerr := b.send(ctx, u.TelegramID, "⚠️ "+err.Error(), ui.CancelInlineKeyboard())
		if rerr != nil {
			return err
		}
		// نشست باز بماند تا کاربر مقدار درست بفرستد.
		return nil
	}

	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}

	// سینک: منوی ویرایش در همان پیام قبلی بروزرسانی می‌شود.
	if data.Back != nil {
		b.showProfileEdit(ctx, u, data.Back.ChatID, data.Back.MessageID)
		return nil
	}
	b.showProfileEdit(ctx, u, u.TelegramID, 0)
	return nil
}

// handleTitleInput عنوان جدید می‌سازد یا عنوان تسک موجودی را ویرایش می‌کند.
func (b *Bot) handleTitleInput(ctx context.Context, u *domain.User, session domain.BotSession, text string) error {
	var data sessionData
	_ = json.Unmarshal([]byte(session.Data), &data)

	// حالت ویرایش: نشست حاوی شناسه‌ی تسک است.
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
			return b.renderTaskCard(ctx, data.Back.ChatID, data.Back.MessageID, id, data.Back.origin())
		}
		return nil
	}

	// حالت ایجاد تسک جدید.
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	task := &domain.Task{OwnerID: u.ID, AssigneeID: u.ID, Title: text, Priority: "normal", Status: "pending"}
	if err := b.tasks.Create(ctx, task); err != nil {
		return err
	}
	_, err := b.send(ctx, u.TelegramID,
		fmt.Sprintf("✅ تسک «%s» ثبت شد.\nهر زمان انجامش دادی از برنامه‌ی امروز تیکش بزن.", task.Title),
		ui.TodayInlineKeyboard())
	return err
}

// handleDescriptionInput توضیحات تسکِ در حال ویرایش را ذخیره می‌کند.
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
		return b.renderTaskCard(ctx, data.Back.ChatID, data.Back.MessageID, id, data.Back.origin())
	}
	return nil
}

// handleSearchInput جستجوی کاربر را اجرا و نتایج را نشان می‌دهد.
func (b *Bot) handleSearchInput(ctx context.Context, u *domain.User, session domain.BotSession, text string, chatID int64) error {
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	tasks, err := b.tasks.Search(ctx, u.ID, text, 10)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("🔍 نتایج جستجو برای «%s»:\n\n", text)
	if len(tasks) == 0 {
		header += "چیزی پیدا نشد 🤷"
	} else {
		for _, t := range tasks {
			header += ui.FormatSmallInfo(&t) + "\n"
		}
	}
	_, err = b.send(ctx, chatID, header, ui.SearchResultsKeyboard(tasks))
	return err
}

// handleAssignInput ورودی «واگذاری تسک» را تجزیه و تسک‌ها را ثبت می‌کند.
func (b *Bot) handleAssignInput(ctx context.Context, u *domain.User, text string, chatID int64) error {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		_, err := b.send(ctx, chatID, "لطفا یوزرنیم را در خط اول و هر تسک را در یک خط جدا بفرست.", ui.CancelInlineKeyboard())
		return err
	}
	username := strings.TrimPrefix(strings.TrimSpace(lines[0]), "@")
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		_, err := b.send(ctx, chatID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", ui.CancelInlineKeyboard())
		return err
	}
	if target.ID == u.ID {
		_, err := b.send(ctx, chatID, "نمی‌تونی تسک رو به خودت واگذار کنی 🙂", ui.CancelInlineKeyboard())
		return err
	}
	if err := b.setTaskForOther(ctx, u.ID, target.ID, lines[1:], "normal"); err != nil {
		return err
	}
	return b.clearSession(ctx, u.ID)
}

// handleStatusInput وضعیت تسک‌های واگذارشده به کاربر هدف را نشان می‌دهد.
func (b *Bot) handleStatusInput(ctx context.Context, u *domain.User, text string, chatID int64) error {
	username := strings.TrimPrefix(strings.TrimSpace(text), "@")
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		_, err := b.send(ctx, chatID, "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن.", ui.CancelInlineKeyboard())
		return err
	}
	if err := b.clearSession(ctx, u.ID); err != nil {
		return err
	}
	return b.renderDelegatedStatus(ctx, chatID, 0, u.ID, target.ID)
}

// startEditSession برای ویرایش یک تسک، نشست کاربر را با شناسه‌ی تسک و مقصد بازگشت آماده می‌کند.
func (b *Bot) startEditSession(ctx context.Context, userID uuid.UUID, state string, taskID uuid.UUID, back *backRef) error {
	return b.setSession(ctx, userID, state, sessionData{TaskID: taskID.String(), Back: back})
}
