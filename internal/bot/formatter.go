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

// FormatTask کارت کامل یک تسک را به فارسی می‌سازد.
func FormatTask(task *domain.Task) string {
	due := "ندارد"
	if task.DueAt != nil {
		due = task.DueAt.Format("01-02 15:04")
	}
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
		due,
		evidence,
		completed,
	)
}

// FormatTasks خلاصه‌ی چند تسک را در یک متن می‌سازد.
func FormatTasks(tasks []domain.Task) string {
	var result string
	for _, task := range tasks {
		result += fmt.Sprintf("📝 عنوان: %s\n⚡ اولویت: %s   📌 وضعیت: %s\n\n",
			task.Title,
			priorityLabel(task.Priority),
			statusLabel(task.Status),
		)
	}
	return result
}

// FormatSmallInfo یک خط خلاصه برای فهرست‌ها می‌سازد.
func FormatSmallInfo(task *domain.Task) string {
	return fmt.Sprintf("📝 %s\n%s | %s", task.Title, priorityLabel(task.Priority), statusLabel(task.Status))
}
