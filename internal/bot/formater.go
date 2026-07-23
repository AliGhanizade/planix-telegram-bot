package bot

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

func FormatTask(task *domain.Task) string {
	var due string
	if task.DueAt != nil {
		due = task.DueAt.Format("01-02 15:04")
	} else {
		due = "ندارد"
	}

	var completed string
	if task.CompletedAt != nil {
		completed = task.CompletedAt.Format("01-02 15:04")
	} else {
		completed = "-"
	}

	description := task.Description
	if description == "" {
		description = "ندارد"
	}

	priority := map[string]string{
		"low":    "🟢 کم",
		"normal": "🟡 معمولی",
		"high":   "🟠 زیاد",
		"urgent": "🔴 فوری",
	}[task.Priority]

	if priority == "" {
		priority = task.Priority
	}

	status := map[string]string{
		"pending":   "⏳ در انتظار",
		"completed": "✅ انجام شده",
	}[task.Status]

	if status == "" {
		status = task.Status
	}

	evidence := "❌ خیر"
	if task.RequiresEvidence {
		evidence = "✅ بله"
	}

	return fmt.Sprintf(`📝 عنوان: %s
📄 توضیحات: %s
⚡ اولویت: %s   📌 وضعیت:%s
📅 موعد: %s
📷 نیاز به مدرک: %s
✅ زمان انجام: %s

`,
		task.Title,
		description,
		priority,
		status,
		due,
		evidence,
		completed,
	)
}

func FormatTasks(tasks []domain.Task) string {
	var result string
	for _, task := range tasks {


		priority := map[string]string{
			"low":    "🟢 کم",
			"normal": "🟡 معمولی",
			"high":   "🟠 زیاد",
			"urgent": "🔴 فوری",
		}[task.Priority]

		if priority == "" {
			priority = task.Priority
		}

		status := map[string]string{
			"pending":   "⏳ در انتظار",
			"completed": "✅ انجام شده",
		}[task.Status]

		if status == "" {
			status = task.Status
		}

		result += fmt.Sprintf(`📝 عنوان: %s
⚡ اولویت: %s   📌 وضعیت:%s

`,
			task.Title,
			priority,
			status,
		)
	}
	return result
}


func FormatSmallInfo(task *domain.Task) string {
	var result string


		priority := map[string]string{
			"low":    "🟢 کم",
			"normal": "🟡 معمولی",
			"high":   "🟠 زیاد",
			"urgent": "🔴 فوری",
		}[task.Priority]

		if priority == "" {
			priority = task.Priority
		}

		status := map[string]string{
			"pending":   "⏳ در انتظار",
			"completed": "✅ انجام شده",
		}[task.Status]

		if status == "" {
			status = task.Status
		}

		result += fmt.Sprintf(`
		📝 عنوان: %s
⚡ اولویت: %s   📌 وضعیت:%s
`,
			task.Title,
			priority,
			status,
		)
	return result
}
