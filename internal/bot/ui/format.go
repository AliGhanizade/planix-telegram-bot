package ui

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

// PriorityLabels maps priority codes to per language labels.
var PriorityLabels = map[string]map[Lang]string{
	"low":    {Fa: "🟢 کم", En: "🟢 Low"},
	"normal": {Fa: "🟡 معمولی", En: "🟡 Normal"},
	"high":   {Fa: "🟠 زیاد", En: "🟠 High"},
	"urgent": {Fa: "🔴 فوری", En: "🔴 Urgent"},
}

// StatusLabels maps status codes to per language labels.
var StatusLabels = map[string]map[Lang]string{
	"pending":   {Fa: "⏳ در انتظار", En: "⏳ Pending"},
	"completed": {Fa: "✅ انجام شده", En: "✅ Completed"},
	"cancelled": {Fa: "❌ لغو شده", En: "❌ Cancelled"},
}

// StatusEmoji is the status icon used in lists.
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
func PriorityLabel(priority string, l Lang) string {
	if labels, ok := PriorityLabels[priority]; ok {
		return labels[l]
	}
	return priority
}

// StatusLabel returns the display label of a status.
func StatusLabel(status string, l Lang) string {
	if labels, ok := StatusLabels[status]; ok {
		return labels[l]
	}
	return status
}

// DueLabel formats the task due date.
func DueLabel(task *domain.Task, l Lang) string {
	if task.DueAt == nil {
		if l == En {
			return "none"
		}
		return "ندارد"
	}
	return task.DueAt.Format("01-02 15:04")
}

// FormatTask builds the full task card.
func FormatTask(task *domain.Task, l Lang) string {
	completed := "-"
	if task.CompletedAt != nil {
		completed = task.CompletedAt.Format("01-02 15:04")
	}
	description := task.Description
	if description == "" {
		if l == En {
			description = "none"
		} else {
			description = "ندارد"
		}
	}
	evidence := "❌"
	if task.RequiresEvidence {
		evidence = "✅"
	}

	if l == En {
		return fmt.Sprintf(`📝 Title: %s
📄 Description: %s
⚡ Priority: %s   📌 Status: %s
📅 Due: %s
📷 Needs proof: %s
✅ Completed at: %s
`,
			task.Title,
			description,
			PriorityLabel(task.Priority, l),
			StatusLabel(task.Status, l),
			DueLabel(task, l),
			evidence,
			completed,
		)
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
		PriorityLabel(task.Priority, l),
		StatusLabel(task.Status, l),
		DueLabel(task, l),
		evidence,
		completed,
	)
}

// FormatSmallInfo builds a one line summary for lists.
func FormatSmallInfo(task *domain.Task, l Lang) string {
	line := fmt.Sprintf("%s %s — %s", StatusEmoji(task.Status), task.Title, PriorityLabel(task.Priority, l))
	if task.DueAt != nil {
		line += " — 📅 " + task.DueAt.Format("01-02 15:04")
	}
	return line
}
