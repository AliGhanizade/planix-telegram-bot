package bot

import (
	"fmt"
	"strconv"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
)

// استایل‌های رنگی دکمه‌ها (Bot API 9.4+) — رنگ نهایی به تم کلاینت بستگی دارد.
const (
	StylePrimary = "primary" // آبی
	StyleSuccess = "success" // سبز
	StyleDanger  = "danger"  // قرمز
)

// kb دکمه‌ی کیبورد پاسخ (پایین صفحه) می‌سازد.
func kb(text, style string) models.KeyboardButton {
	return models.KeyboardButton{Text: text, Style: style}
}

// ib دکمه‌ی شیشه‌ای داخل پیام می‌سازد.
func ib(text, data, style string) models.InlineKeyboardButton {
	return models.InlineKeyboardButton{Text: text, CallbackData: data, Style: style}
}

// MainKeyboard کیبورد اصلی فارسی بات با دکمه‌های رنگی است.
func MainKeyboard() models.ReplyKeyboardMarkup {
	return models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{kb("➕ تسک جدید", StyleSuccess), kb("📋 برنامه امروز", StylePrimary)},
			{kb("🔍 جستجو", ""), kb("👥 واگذاری تسک", StylePrimary), kb("📊 وضعیت وظایف دیگران", "")},
			{kb("👤 پروفایل", ""), kb("⚙️ تنظیمات", ""), kb("ℹ️ راهنما", "")},
		},
		IsPersistent:          true,
		ResizeKeyboard:        true,
		InputFieldPlaceholder: "یکی از دکمه‌ها را بزن یا عنوان تسک را بنویس…",
	}
}

// MenuInlineKeyboard منوی داخل پیام برای ناوبری درجا.
func MenuInlineKeyboard() models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{ib("➕ تسک جدید", "nav:new", StyleSuccess), ib("📋 برنامه امروز", "nav:today", StylePrimary)},
			{ib("🔍 جستجو", "nav:search", ""), ib("👥 واگذاری تسک", "nav:assign", StylePrimary)},
			{ib("📊 وضعیت وظایف", "nav:status", ""), ib("👤 پروفایل", "nav:profile", "")},
			{ib("⚙️ تنظیمات", "nav:settings", ""), ib("ℹ️ راهنما", "nav:help", "")},
		},
	}
}

// BackKeyboard فقط دکمه‌ی بازگشت به منو.
func BackKeyboard() models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{ib("🏠 منوی اصلی", "nav:menu", StylePrimary)},
		},
	}
}

// CancelInlineKeyboard دکمه‌ی لغو جریان جاری.
func CancelInlineKeyboard() models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{ib("❌ لغو", "state:cancel", StyleDanger)},
		},
	}
}

// truncate عنوان را برای جا شدن در دکمه کوتاه می‌کند.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// taskOrigin مبدأ نمایش یک تسک را مشخص می‌کند: فهرست (با فیلتر و صفحه) یا کارت مستقل.
type taskOrigin struct {
	List   bool
	Filter string
	Page   int
}

// noOrigin مبدأ مستقل (کارت باز‌شده از جستجو و…).
var noOrigin = taskOrigin{}

// suffix پسوند مبدأ را برای دیتای کال‌بک می‌سازد.
func (o taskOrigin) suffix() string {
	if !o.List {
		return ":C"
	}
	return fmt.Sprintf(":L:%s:%d", o.Filter, o.Page)
}

// parseTaskOrigin پسوند مبدأ را از اجزای دیتای کال‌بک می‌خواند.
func parseTaskOrigin(parts []string, idx int) (taskOrigin, error) {
	if idx >= len(parts) {
		return noOrigin, nil
	}
	switch parts[idx] {
	case "C":
		return noOrigin, nil
	case "L":
		o := taskOrigin{List: true, Filter: "pending", Page: 1}
		if idx+1 < len(parts) {
			o.Filter = parts[idx+1]
		}
		if idx+2 < len(parts) {
			if p, err := strconv.Atoi(parts[idx+2]); err == nil {
				o.Page = p
			}
		}
		return o, nil
	default:
		return noOrigin, fmt.Errorf("unknown origin: %s", parts[idx])
	}
}

// taskData دیتای کال‌بک یک عملیات تسک را با مبدأ می‌سازد.
func taskData(action string, id interface{ String() string }, o taskOrigin) string {
	return "task:" + action + ":" + id.String() + o.suffix()
}

// listData دیتای کال‌بک فهرست تسک‌ها را می‌سازد.
func listData(f listFilter, page int) string {
	return fmt.Sprintf("list:tasks:%s:%d", f, page)
}

// backDataOf دکمه‌ی بازگشتِ مناسب برای یک مبدأ را برمی‌گرداند.
func backDataOf(o taskOrigin) string {
	if o.List {
		return listData(listFilter(o.Filter), o.Page)
	}
	return "nav:menu"
}

// TaskListKeyboard فهرست تسک‌ها با دکمه‌ی انجام/جزئیات، تب‌های فیلتر و صفحه‌بندی.
func TaskListKeyboard(tasks []domain.Task, f listFilter, page, pages int) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	for _, t := range tasks {
		o := taskOrigin{List: true, Filter: string(f), Page: page}
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: " ℹ️ " + truncate(t.Title, 36), CallbackData: taskData("info", t.ID, o)},
			{Text: "✅", CallbackData: taskData("done", t.ID, o), Style: StyleSuccess},
		})
	}

	// تب‌های فیلتر — تب فعال آبی می‌شود
	rows = append(rows, []models.InlineKeyboardButton{
		filterButton("⏳ باز", filterPending, f),
		filterButton("✅ انجام‌شده", filterCompleted, f),
		filterButton("🗂 همه", filterAll, f),
	})

	// صفحه‌بندی
	if pages > 1 {
		var nav []models.InlineKeyboardButton
		if page > 1 {
			nav = append(nav, ib("⬅️ قبلی", listData(f, page-1), ""))
		}
		nav = append(nav, ib(fmt.Sprintf("📄 %d از %d", page, pages), "nav:noop", ""))
		if page < pages {
			nav = append(nav, ib("بعدی ➡️", listData(f, page+1), ""))
		}
		rows = append(rows, nav)
	}

	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "🏠 منوی اصلی", CallbackData: "nav:menu", Style: StylePrimary},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func filterButton(label string, f, active listFilter) models.InlineKeyboardButton {
	data := listData(f, 1)
	if f == active {
		return ib("● "+label, data, StylePrimary)
	}
	return ib(label, data, "")
}

// TaskCardKeyboard کارت تسک را با دکمه‌های مدیریتی رنگی می‌سازد.
func TaskCardKeyboard(task *domain.Task, o taskOrigin) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton

	if task.Status == "completed" {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: "🔄 بازگشایی تسک", CallbackData: taskData("reopen", task.ID, o), Style: StylePrimary},
		})
	} else {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: "✅ انجام شد", CallbackData: taskData("done", task.ID, o), Style: StyleSuccess},
		})
	}

	rows = append(rows, []models.InlineKeyboardButton{
		ib("📝 عنوان", taskData("edit:title", task.ID, o), ""),
		ib("📄 توضیحات", taskData("edit:desc", task.ID, o), ""),
	})
	rows = append(rows, []models.InlineKeyboardButton{
		ib("📅 موعد", taskData("edit:due", task.ID, o), ""),
		ib("⚡ اولویت", taskData("edit:priority", task.ID, o), ""),
	})
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "🗑 حذف", CallbackData: taskData("delete", task.ID, o), Style: StyleDanger},
		ib("🔄 بروزرسانی", taskData("refresh", task.ID, o), ""),
	})
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "↩️ بازگشت", CallbackData: backDataOf(o), Style: StylePrimary},
	})

	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// DuePickerKeyboard انتخاب سریع موعد تسک.
func DuePickerKeyboard(task *domain.Task, o taskOrigin) models.InlineKeyboardMarkup {
	id := task.ID
	due := func(preset string) string {
		return "task:due:" + id.String() + ":" + preset + o.suffix()
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "🌅 امروز تا شب", CallbackData: due("today"), Style: StylePrimary}},
			{{Text: "☀️ فردا صبح", CallbackData: due("tomorrow")}},
			{{Text: "🗓 هفته‌ی بعد", CallbackData: due("week")}},
			{{Text: "🗑 بدون موعد", CallbackData: due("none"), Style: StyleDanger}},
			{{Text: "↩️ انصراف", CallbackData: taskData("refresh", id, o)}},
		},
	}
}

// PriorityPickerKeyboard انتخاب اولویت تسک.
func PriorityPickerKeyboard(task *domain.Task, o taskOrigin) models.InlineKeyboardMarkup {
	id := task.ID
	prio := func(level string) string {
		return "task:prio:" + id.String() + ":" + level + o.suffix()
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "🟢 کم", CallbackData: prio("low")}},
			{{Text: "🟡 معمولی", CallbackData: prio("normal"), Style: StylePrimary}},
			{{Text: "🟠 زیاد", CallbackData: prio("high"), Style: StyleSuccess}},
			{{Text: "🔴 فوری", CallbackData: prio("urgent"), Style: StyleDanger}},
			{{Text: "↩️ انصراف", CallbackData: taskData("refresh", id, o)}},
		},
	}
}

// DeleteConfirmKeyboard تایید دو مرحله‌ای حذف تسک.
func DeleteConfirmKeyboard(task *domain.Task, o taskOrigin) models.InlineKeyboardMarkup {
	id := task.ID
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🗑 بله، حذف کن", CallbackData: taskData("delete:yes", id, o), Style: StyleDanger},
				{Text: "انصراف", CallbackData: taskData("refresh", id, o)},
			},
		},
	}
}

// SuggestFriendInlineKeyboard فهرست کاربرانی که به آن‌ها تسک داده‌ای را پیشنهاد می‌دهد.
func SuggestFriendInlineKeyboard(users []domain.User) models.InlineKeyboardMarkup {
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
		rows = append(rows, []models.InlineKeyboardButton{
			ib("هنوز به کسی تسک نداده‌ای", "nav:noop", ""),
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "❌ لغو", CallbackData: "state:cancel", Style: StyleDanger},
	})
	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// SearchResultsKeyboard نتیجه‌های جستجو با دکمه‌ی جزئیات هر تسک.
func SearchResultsKeyboard(tasks []domain.Task) models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton
	for _, t := range tasks {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: " ℹ️ " + truncate(t.Title, 40), CallbackData: taskData("info", t.ID, noOrigin)},
		})
	}
	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "🏠 منوی اصلی", CallbackData: "nav:menu", Style: StylePrimary},
	})
	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// SettingsInlineKeyboard تنظیمات کاربر با وضعیت فعلی.
func SettingsInlineKeyboard(dailyReport bool) models.InlineKeyboardMarkup {
	label := "گزارش روزانه: خاموش ❌"
	style := ""
	if dailyReport {
		label = "گزارش روزانه: روشن ✅"
		style = StyleSuccess
	}
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: label, CallbackData: "settings:daily", Style: style}},
			{{Text: "🏠 منوی اصلی", CallbackData: "nav:menu", Style: StylePrimary}},
		},
	}
}

// TodayInlineKeyboard دکمه‌ی میانبر بعد از ثبت تسک جدید.
func TodayInlineKeyboard() models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "📋 برنامه امروز", CallbackData: "nav:today", Style: StylePrimary}},
		},
	}
}
