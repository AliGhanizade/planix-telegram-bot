package bot

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

// priorityLabels برچسب فارسی اولویت‌ها.
var priorityLabels = map[string]string{
	"low":    "🟢 کم",
	"normal": "🟡 معمولی",
	"high":   "🟠 زیاد",
	"urgent": "🔴 فوری",
}

// statusLabels برچسب فارسی وضعیت‌ها.
var statusLabels = map[string]string{
	"pending":   "⏳ در انتظار",
	"completed": "✅ انجام شده",
	"cancelled": "❌ لغو شده",
}

// statusEmoji نشان وضعیت برای فهرست‌ها.
func statusEmoji(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "cancelled":
		return "❌"
	default:
		return "⏳"
	}
}

// priorityLabel برچسب فارسی اولویت را برمی‌گرداند؛ در نبود برچسب، مقدار خام.
func priorityLabel(priority string) string {
	if label, ok := priorityLabels[priority]; ok {
		return label
	}
	return priority
}

// statusLabel برچسب فارسی وضعیت را برمی‌گرداند.
func statusLabel(status string) string {
	if label, ok := statusLabels[status]; ok {
		return label
	}
	return status
}

// dueLabel موعد تسک را به شکل خوانا برمی‌گرداند.
func dueLabel(task *domain.Task) string {
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
		priorityLabel(task.Priority),
		statusLabel(task.Status),
		dueLabel(task),
		evidence,
		completed,
	)
}

// FormatSmallInfo یک خط خلاصه برای فهرست‌ها می‌سازد.
func FormatSmallInfo(task *domain.Task) string {
	line := fmt.Sprintf("%s %s — %s", statusEmoji(task.Status), task.Title, priorityLabel(task.Priority))
	if task.DueAt != nil {
		line += " — 📅 " + task.DueAt.Format("01-02 15:04")
	}
	return line
}
