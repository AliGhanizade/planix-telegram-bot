package ui

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

// PriorityLabels maps priority codes to labels.
var PriorityLabels = map[string]string{
	"low":    "🟢 کم",
	"normal": "🟡 معمولی",
	"high":   "🟠 زیاد",
	"urgent": "🔴 فوری",
}

// StatusLabels maps status codes to labels.
var StatusLabels = map[string]string{
	"pending":   "⏳ در انتظار",
	"completed": "✅ انجام شده",
	"cancelled": "❌ لغو شده",
}

// statusEmoji is the status icon used in lists.
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

// PriorityLabel returns the display label of a priority.
func PriorityLabel(priority string) string {
	if label, ok := PriorityLabels[priority]; ok {
		return label
	}
	return priority
}

// StatusLabel returns the display label of a status.
func StatusLabel(status string) string {
	if label, ok := StatusLabels[status]; ok {
		return label
	}
	return status
}

// DueLabel formats the task due date.
func DueLabel(task *domain.Task) string {
	if task.DueAt == nil {
		return "ندارد"
	}
	return task.DueAt.Format("01-02 15:04")
}

// FormatTask builds the full task card.
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

// FormatSmallInfo builds a one line summary for lists.
func FormatSmallInfo(task *domain.Task) string {
	line := fmt.Sprintf("%s %s — %s", StatusEmoji(task.Status), task.Title, PriorityLabel(task.Priority))
	if task.DueAt != nil {
		line += " — 📅 " + task.DueAt.Format("01-02 15:04")
	}
	return line
}
