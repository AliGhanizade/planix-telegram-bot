package ui

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

// PriorityLabels برچسب فارسی اولویت‌ها. برچسب فارسی اولویت‌ها.
var PriorityLabels = map[string]string{
	"low":    "🟢 کم",
	"normal": "🟡 معمولی",
	"high":   "🟠 زیاد",
	"urgent": "🔴 فوری",
}

// StatusLabels برچسب فارسی وضعیت‌ها. برچسب فارسی وضعیت‌ها.
var StatusLabels = map[string]string{
	"pending":   "⏳ در انتظار",
	"completed": "✅ انجام شده",
	"cancelled": "❌ لغو شده",
}

// statusEmoji نشان وضعیت برای فهرست‌ها.
func StatusEmoji(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "cancelled":
		return "❌"
	default:
		return "⏳"
	}
}

// PriorityLabel برچسب فارسی اولویت را برمی‌گرداند؛ در نبود برچسب، مقدار خام.
func PriorityLabel(priority string) string {
	if label, ok := PriorityLabels[priority]; ok {
		return label
	}
	return priority
}

// StatusLabel برچسب فارسی وضعیت را برمی‌گرداند.
func StatusLabel(status string) string {
	if label, ok := StatusLabels[status]; ok {
		return label
	}
	return status
}

// DueLabel موعد تسک را به شکل خوانا برمی‌گرداند.
func DueLabel(task *domain.Task) string {
	if task.DueAt == nil {
		return "ندارد"
	}
	return task.DueAt.Format("01-02 15:04")
}

// FormatTask کارت کامل یک تسک را به فارسی می‌سازد.
func FormatTask(task *domain.Task) string {
	completed := "-"
	if task.CompletedAt != nil {
		completed = task.CompletedAt.Format("01-02 15:04")
	}
	description := task.Description
	if description == "" {
		description = "ندارد"
	}
	evidence := "❌ خیر"
	if task.RequiresEvidence {
		evidence = "✅ بله"
	}

	return fmt.Sprintf(`📝 عنوان: %s
📄 توضیحات: %s
⚡ اولویت: %s   📌 وضعیت: %s
📅 موعد: %s
📷 نیاز به مدرک: %s
✅ زمان انجام: %s
`,
		task.Title,
		description,
		PriorityLabel(task.Priority),
		StatusLabel(task.Status),
		DueLabel(task),
		evidence,
		completed,
	)
}

// FormatSmallInfo یک خط خلاصه برای فهرست‌ها می‌سازد.
func FormatSmallInfo(task *domain.Task) string {
	line := fmt.Sprintf("%s %s — %s", StatusEmoji(task.Status), task.Title, PriorityLabel(task.Priority))
	if task.DueAt != nil {
		line += " — 📅 " + task.DueAt.Format("01-02 15:04")
	}
	return line
}
