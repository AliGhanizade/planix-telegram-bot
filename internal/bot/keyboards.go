package bot

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("👥 اعمال وظایف دیگران"), tgbotapi.NewKeyboardButton("➕ تسک جدید")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("✅ وضعیت وظایف دیگران"), tgbotapi.NewKeyboardButton("📋 برنامه امروز"), tgbotapi.NewKeyboardButton("📅 برنامه‌های من")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("پشتیبانی"), tgbotapi.NewKeyboardButton("پروفایل"), tgbotapi.NewKeyboardButton("ℹ️ راهنما")),
	)
}

func TasksInlineKeyboard(tasks []domain.Task) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(tasks)*2)
	for _, task := range tasks {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(" ℹ️ "+task.Title, "task:info:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData(task.Title+" ✅ ", "task:done:"+task.ID.String()),
		))

	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func CancelStateInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو", "state:cancel"),
		),
	)
}

func TaskInlineKeyboard(task *domain.Task) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ انجام شد", "task:done:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData("❌ انجام نشد", "task:cancel:"+task.ID.String()),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 عنوان", "task:edit:title:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData("📄 توضیحات", "task:edit:desc:"+task.ID.String()),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 موعد", "task:edit:due:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData("⚡ اولویت", "task:edit:priority:"+task.ID.String()),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 حذف", "task:delete:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData("🔄 بروزرسانی", "task:refresh:"+task.ID.String()),
		),
	)
}

func SuggestFriendInlineKeyboard(users []domain.User) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(users)*2)
	for _, user := range users {

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(user.Username, "user:filter:fr:"+user.Username),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ لغو", "state:cancel"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func FilterInlineKeyboard(tasks []domain.Task) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(tasks)*2)
	for _, task := range tasks {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("All Task", "task:filter:not:"+task.ID.String()),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌", "task:filter:pendding:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData("✅", "task:filter:completed:"+task.ID.String()),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌", "task:filter:not:"+task.ID.String()),
			tgbotapi.NewInlineKeyboardButtonData("✅", "task:filter:done:"+task.ID.String()),
		))

	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
