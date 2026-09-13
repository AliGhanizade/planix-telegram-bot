package ui

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
)

// button styles (Bot API 9.4+) - final colors depend on the client theme.
const (
	StylePrimary = "primary" // blue
	StyleSuccess = "success" // green
	StyleDanger  = "danger"  // red
)

// kb builds a reply keyboard button.
func kb(text, style string) models.KeyboardButton {
	return models.KeyboardButton{Text: text, Style: style}
}

// ib builds an inline button.
func ib(text, data, style string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: text, CallbackData: data, Style: style}
}

// BackLabel is the shared back-to-menu label.
func BackLabel(l Lang) string {
	if l == En {
		return "🏠 Main menu"
	}
	return "🏠 منوی اصلی"
}

// MainKeyboard is the main colored reply keyboard, icons come from the
// button colors instead of emoji.
func MainKeyboard(l Lang) models.ReplyKeyboardMarkup {
	if l == En {
		return models.ReplyKeyboardMarkup{
			Keyboard: [][]models.KeyboardButton{
				{kb("New task", StyleSuccess), kb("Today", StylePrimary)},
				{kb("Search", StylePrimary), kb("Delegate", StyleSuccess), kb("Delegated status", StylePrimary)},
				{kb("Profile", ""), kb("Settings", ""), kb("Help", "")},
			},
			IsPersistent:          true,
			ResizeKeyboard:        true,
			InputFieldPlaceholder: "Tap a button or type a task title…",
		}
	}
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{kb("تسک جدید", StyleSuccess), kb("برنامه امروز", StylePrimary)},
			{kb("جستجو", StylePrimary), kb("واگذاری تسک", StyleSuccess), kb("وضعیت وظایف", StylePrimary)},
			{kb("پروفایل", ""), kb("تنظیمات", ""), kb("راهنما", "")},
		},
		IsPersistent:          true,
		ResizeKeyboard:        true,
		InputFieldPlaceholder: "یکی از دکمه‌ها را بزن یا عنوان تسک را بنویس…",
	}
}

// MenuInlineKeyboard is the inline menu for in place navigation.
func MenuInlineKeyboard(l Lang) models.InlineKeyboardMarkup {
	if l == En {
		return models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{ib("➕ New task", "nav:new", StyleSuccess), ib("📋 Today", "nav:today", StylePrimary)},
				{ib("🔍 Search", "nav:search", ""), ib("👥 Delegate", "nav:assign", StylePrimary)},
				{ib("📊 Delegated status", "nav:status", ""), ib("👤 Profile", "nav:profile", "")},
				{ib("⚙️ Settings", "nav:settings", ""), ib("ℹ️ Help", "nav:help", "")},
			},
		}
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{ib("➕ تسک جدید", "nav:new", StyleSuccess), ib("📋 برنامه امروز", "nav:today", StylePrimary)},
			{ib("🔍 جستجو", "nav:search", ""), ib("👥 واگذاری تسک", "nav:assign", StylePrimary)},
			{ib("📊 وضعیت وظایف", "nav:status", ""), ib("👤 پروفایل", "nav:profile", "")},
			{ib("⚙️ تنظیمات", "nav:settings", ""), ib("ℹ️ راهنما", "nav:help", "")},
		},
	}
}

// BackKeyboard is a single back-to-menu button.
func BackKeyboard(l Lang) models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{ib(BackLabel(l), "nav:menu", StylePrimary)},
		},
	}
}

// CancelInlineKeyboard cancels the current flow.
func CancelInlineKeyboard(l Lang) models.InlineKeyboardMarkup {
	label := "❌ لغو"
	if l == En {
		label = "❌ Cancel"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{ib(label, "state:cancel", StyleDanger)},
		},
	}
}

// TaskListKeyboard builds the task list with done/details buttons, filter tabs and pagination.
func TaskListKeyboard(tasks []domain.Task, f ListFilter, page, pages int, l Lang) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	back := "🏠 منوی اصلی"
	prev, next := "⬅️ قبلی", "بعدی ➡️"
	open, done, all := "⏳ باز", "✅ انجام‌شده", "🗂 همه"
	pageOf := "از"
	if l == En {
		back = "🏠 Main menu"
		prev, next = "⬅️ Prev", "Next ➡️"
		open, done, all = "⏳ Open", "✅ Done", "🗂 All"
		pageOf = "of"
	}

	for _, t := range tasks {
		o := TaskOrigin{List: true, Filter: string(f), Page: page}
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: " ℹ️ " + Truncate(t.Title, 36), CallbackData: TaskData("info", t.ID, o)},
			{Text: "✅", CallbackData: TaskData("done", t.ID, o), Style: StyleSuccess},
		})
	}

	// filter tabs - the active tab is highlighted
	rows = append(rows, []models.InlineKeyboardButton{
		filterButton(open, FilterPending, f, l),
		filterButton(done, FilterCompleted, f, l),
		filterButton(all, FilterAll, f, l),
	})

	if pages > 1 {
		var nav []models.InlineKeyboardButton
		if page > 1 {
			nav = append(nav, ib(prev, ListData(f, page-1), ""))
		}
		nav = append(nav, ib(fmt.Sprintf("📄 %d %s %d", page, pageOf, pages), "nav:noop", ""))
		if page < pages {
			nav = append(nav, ib(next, ListData(f, page+1), ""))
		}
		rows = append(rows, nav)
	}

	rows = append(rows, []models.InlineKeyboardButton{
		{Text: back, CallbackData: "nav:menu", Style: StylePrimary},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func filterButton(label string, f, active ListFilter, l Lang) models.InlineKeyboardButton {
	data := ListData(f, 1)
	if f == active {
		return ib("● "+label, data, StylePrimary)
	}
	return ib(label, data, "")
}

// TaskCardKeyboard builds the colored task management card.
func TaskCardKeyboard(task *domain.Task, o TaskOrigin, l Lang, hasProof bool) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	editTitle, editDesc := "📝 عنوان", "📄 توضیحات"
	editDue, editPrio := "📅 موعد", "⚡ اولویت"
	del, refresh := "🗑 حذف", "🔄 بروزرسانی"
	back := "↩️ بازگشت"
	doneBtn, reopen := "✅ انجام شد", "🔄 بازگشایی تسک"
	proof, viewProof := "📷 ارسال عکس", "📷 مشاهده عکس"
	if l == En {
		editTitle, editDesc = "📝 Title", "📄 Description"
		editDue, editPrio = "📅 Due date", "⚡ Priority"
		del, refresh = "🗑 Delete", "🔄 Refresh"
		back = "↩️ Back"
		doneBtn, reopen = "✅ Done", "🔄 Reopen task"
		proof, viewProof = "📷 Send photo", "📷 View photo"
	}

	if task.Status == "completed" {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: reopen, CallbackData: TaskData("reopen", task.ID, o), Style: StylePrimary},
		})
	} else {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: doneBtn, CallbackData: TaskData("done", task.ID, o), Style: StyleSuccess},
		})
	}

	rows = append(rows, []models.InlineKeyboardButton{
		ib(editTitle, TaskData("edit:title", task.ID, o), ""),
		ib(editDesc, TaskData("edit:desc", task.ID, o), ""),
	})
	rows = append(rows, []models.InlineKeyboardButton{
		ib(editDue, TaskData("edit:due", task.ID, o), ""),
		ib(editPrio, TaskData("edit:priority", task.ID, o), ""),
	})
	if hasProof {
		rows = append(rows, []models.InlineKeyboardButton{
			ib(proof, TaskData("proof", task.ID, o), ""),
			ib(viewProof, TaskData("viewproof", task.ID, o), StylePrimary),
		})
	} else {
		rows = append(rows, []models.InlineKeyboardButton{
			ib(proof, TaskData("proof", task.ID, o), ""),
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: del, CallbackData: TaskData("delete", task.ID, o), Style: StyleDanger},
		ib(refresh, TaskData("refresh", task.ID, o), ""),
	})
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: back, CallbackData: BackDataOf(o), Style: StylePrimary},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// DuePickerKeyboard offers quick due date presets.
func DuePickerKeyboard(task *domain.Task, o TaskOrigin, l Lang) models.InlineKeyboardMarkup {
	id := task.ID
	due := func(preset string) string {
		return "task:due:" + id.String() + ":" + preset + o.Suffix()
	}
	cancel := "↩️ انصراف"
	today, tomorrow, week, none := "🌅 امروز تا شب", "☀️ فردا صبح", "🗓 هفته‌ی بعد", "🗑 بدون موعد"
	if l == En {
		cancel = "↩️ Cancel"
		today, tomorrow, week, none = "🌅 Today evening", "☀️ Tomorrow morning", "🗓 Next week", "🗑 No due date"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: today, CallbackData: due("today"), Style: StylePrimary}},
			{{Text: tomorrow, CallbackData: due("tomorrow")}},
			{{Text: week, CallbackData: due("week")}},
			{{Text: none, CallbackData: due("none"), Style: StyleDanger}},
			{{Text: cancel, CallbackData: TaskData("refresh", id, o)}},
		},
	}
}

// PriorityPickerKeyboard lets the user pick a task priority.
func PriorityPickerKeyboard(task *domain.Task, o TaskOrigin, l Lang) models.InlineKeyboardMarkup {
	id := task.ID
	prio := func(level string) string {
		return "task:prio:" + id.String() + ":" + level + o.Suffix()
	}
	cancel := "↩️ انصراف"
	low, normal, high, urgent := "🟢 کم", "🟡 معمولی", "🟠 زیاد", "🔴 فوری"
	if l == En {
		cancel = "↩️ Cancel"
		low, normal, high, urgent = "🟢 Low", "🟡 Normal", "🟠 High", "🔴 Urgent"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: low, CallbackData: prio("low")}},
			{{Text: normal, CallbackData: prio("normal"), Style: StylePrimary}},
			{{Text: high, CallbackData: prio("high"), Style: StyleSuccess}},
			{{Text: urgent, CallbackData: prio("urgent"), Style: StyleDanger}},
			{{Text: cancel, CallbackData: TaskData("refresh", id, o)}},
		},
	}
}

// DeleteConfirmKeyboard is the two step delete confirmation.
func DeleteConfirmKeyboard(task *domain.Task, o TaskOrigin, l Lang) models.InlineKeyboardMarkup {
	id := task.ID
	yes, no := "🗑 بله، حذف کن", "انصراف"
	if l == En {
		yes, no = "🗑 Yes, delete", "Cancel"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: yes, CallbackData: TaskData("delete:yes", id, o), Style: StyleDanger},
				{Text: no, CallbackData: TaskData("refresh", id, o)},
			},
		},
	}
}

// SuggestFriendInlineKeyboard suggests users you assigned tasks to.
func SuggestFriendInlineKeyboard(users []domain.User, l Lang) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton
	for _, user := range users {
		if user.Username == "" {
			continue
		}
		rows = append(rows, []models.InlineKeyboardButton{
			ib("👤 @"+user.Username, "user:filter:fr:"+user.Username, StylePrimary),
		})
	}
	if len(rows) == 0 {
		empty := "هنوز به کسی تسک نداده‌ای"
		if l == En {
			empty = "You haven't delegated any tasks yet"
		}
		rows = append(rows, []models.InlineKeyboardButton{
			ib(empty, "nav:noop", ""),
		})
	}
	cancel := "❌ لغو"
	if l == En {
		cancel = "❌ Cancel"
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: cancel, CallbackData: "state:cancel", Style: StyleDanger},
	})
	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// SearchResultsKeyboard lists search results with a details button.
func SearchResultsKeyboard(tasks []domain.Task, l Lang) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton
	for _, t := range tasks {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: " ℹ️ " + Truncate(t.Title, 40), CallbackData: TaskData("info", t.ID, NoOrigin)},
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: BackLabel(l), CallbackData: "nav:menu", Style: StylePrimary},
	})
	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// SettingsInlineKeyboard is the user settings menu.
func SettingsInlineKeyboard(dailyReport bool, l Lang) models.InlineKeyboardMarkup {
	daily := DailyReportLabel(dailyReport, l)
	style := ""
	if dailyReport {
		style = StyleSuccess
	}
	edit, web := "✏️ تغییر اطلاعات", "🌐 اتصال به پنل وب"
	if l == En {
		edit, web = "✏️ Edit info", "🌐 Link web panel"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: daily, CallbackData: "settings:daily", Style: style}},
			{{Text: edit, CallbackData: "settings:profile", Style: StylePrimary}},
			{{Text: web, CallbackData: "settings:web"}},
			{{Text: LanguageLabel(l), CallbackData: "settings:lang", Style: StylePrimary}},
			{{Text: BackLabel(l), CallbackData: "nav:menu", Style: StylePrimary}},
		},
	}
}

// ProfileEditInlineKeyboard picks a profile field to edit.
func ProfileEditInlineKeyboard(l Lang) models.InlineKeyboardMarkup {
	first, last, tz, back := "🪪 نام", "🏷 نام خانوادگی", "🌍 تایم‌زون", "↩️ بازگشت"
	if l == En {
		first, last, tz, back = "🪪 First name", "🏷 Last name", "🌍 Timezone", "↩️ Back"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: first, CallbackData: "settings:edit:firstname", Style: StylePrimary}},
			{{Text: last, CallbackData: "settings:edit:lastname", Style: StylePrimary}},
			{{Text: tz, CallbackData: "settings:edit:timezone", Style: StylePrimary}},
			{{Text: back, CallbackData: "settings:back", Style: StyleDanger}},
		},
	}
}

// TodayInlineKeyboard is the shortcut shown after creating a task.
func TodayInlineKeyboard(l Lang) models.InlineKeyboardMarkup {
	label := "📋 برنامه امروز"
	if l == En {
		label = "📋 Today's plan"
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: label, CallbackData: "nav:today", Style: StylePrimary}},
		},
	}
}
