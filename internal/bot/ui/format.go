package ui

import (
	"fmt"
	"html"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

// Esc escapes user text for safe html parse mode.
func Esc(s string) string { return html.EscapeString(s) }

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

// ProofLine shows the proof status on the task card.
func ProofLine(has bool, l Lang) string {
	if l == En {
		if has {
			return "📷 <b>Proof:</b> yes ✅"
		}
		return "📷 <b>Proof:</b> no ❌"
	}
	if has {
		return "📷 <b>مدرک:</b> دارد ✅"
	}
	return "📷 <b>مدرک:</b> ندارد ❌"
}

// PeopleLine shows who owns and who executes a delegated task.
func PeopleLine(owner, assignee string, l Lang) string {
	if l == En {
		return fmt.Sprintf("👤 <b>Owner:</b> %s   🛠 <b>Assignee:</b> %s", owner, assignee)
	}
	return fmt.Sprintf("👤 <b>مالک:</b> %s   🛠 <b>مجری:</b> %s", owner, assignee)
}

// FormatTask builds the full task card in html.
func FormatTask(task *domain.Task, l Lang) string {
	title := html.EscapeString(task.Title)
	description := html.EscapeString(task.Description)
	completed := "-"
	if task.CompletedAt != nil {
		completed = task.CompletedAt.Format("01-02 15:04")
	}
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

	needProof := "❌"
	if task.RequiresEvidence {
		needProof = "✅"
	}
	if l == En {
		return fmt.Sprintf(`📝 <b>Title:</b> <b>%s</b>
📄 <b>Description:</b> %s
⚡ <b>Priority:</b> %s   📌 <b>Status:</b> %s
📅 <b>Due:</b> %s
📷 <b>Photo proof:</b> %s   🖼 <b>Requires proof:</b> %s
✅ <b>Completed at:</b> %s
`,
			title,
			description,
			PriorityLabel(task.Priority, l),
			StatusLabel(task.Status, l),
			DueLabel(task, l),
			evidence,
			needProof,
			completed,
		)
	}
	return fmt.Sprintf(`📝 <b>عنوان:</b> <b>%s</b>
📄 <b>توضیحات:</b> %s
⚡ <b>اولویت:</b> %s   📌 <b>وضعیت:</b> %s
📅 <b>موعد:</b> %s
📷 <b>عکس مدرک:</b> %s   🖼 <b>نیاز به مدرک:</b> %s
✅ <b>زمان انجام:</b> %s
`,
		title,
		description,
		PriorityLabel(task.Priority, l),
		StatusLabel(task.Status, l),
		DueLabel(task, l),
		evidence,
		needProof,
		completed,
	)
}

// FormatSmallInfo builds a one line summary for lists.
func FormatSmallInfo(task *domain.Task, l Lang) string {
	line := fmt.Sprintf("%s <b>%s</b> — %s", StatusEmoji(task.Status), html.EscapeString(task.Title), PriorityLabel(task.Priority, l))
	if task.DueAt != nil {
		line += " — 📅 " + task.DueAt.Format("01-02 15:04")
	}
	return line
}
