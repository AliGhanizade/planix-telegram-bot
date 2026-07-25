package bot

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MainKeyboard کیبورد اصلی فارسی بات است.
func MainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("👥 اعمال وظایف دیگران"), tgbotapi.NewKeyboardButton("➕ تسک جدید")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("✅ وضعیت وظایف دیگران"), tgbotapi.NewKeyboardButton("📋 برنامه امروز"), tgbotapi.NewKeyboardButton("📅 برنامه‌های من")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("پشتیبانی"), tgbotapi.NewKeyboardButton("پروفایل"), tgbotapi.NewKeyboardButton("ℹ️ راهنما")),
	)
}

// TasksInlineKeyboard برای هر تسک دکمه‌ی جزئیات و تیک زدن می‌گذارد.
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

// CancelStateInlineKeyboard دکمه‌ی لغو جریان جاری را می‌سازد.
func CancelStateInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو", "state:cancel"),
		),
	)
}

// TaskInlineKeyboard کارت تسک را با دکمه‌های مدیریتی کامل می‌سازد.
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

// SuggestFriendInlineKeyboard فهرست کاربرانی که به آن‌ها تسک داده‌ای را پیشنهاد می‌دهد.
func SuggestFriendInlineKeyboard(users []domain.User) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(users))
	for _, user := range users {
		if user.Username == "" {
			continue
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("@"+user.Username, "user:filter:fr:"+user.Username),
		))
	}
	if len(rows) == 0 {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("هنوز به کسی تسک نداده‌ای", "state:cancel"),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ لغو", "state:cancel"),
	))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
