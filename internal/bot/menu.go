package bot

import (
	"context"
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// متن‌های ثابت بات.
const (
	menuText = "🗂 منوی پلنیکس\n\nیک گزینه را انتخاب کن:"

	assignPrompt = "📋 فرمت ارسال\n\n" +
		"خط اول: یوزرنیم\n" +
		"بقیه: هر تسک در یک خط\n\n" +
		"@username\n" +
		"طراحی API\n" +
		"بررسی Pull Request"

	searchPrompt  = "🔍 متن یا کلمه‌ای از عنوان تسک را بفرست:"
	newTaskPrompt = "➕ عنوان تسک را بفرست. مثال: مطالعه گولنگ"

	statusPickPrompt = "📊 برای چه کسی می‌خوای وضعیت تسک‌هاشو ببینی؟"
)

// welcomeMessage پیام خوش‌آمد /start را می‌سازد.
func welcomeMessage(me *models.User) string {
	name := "پلنیکس"
	if me != nil && me.FirstName != "" {
		name = me.FirstName
	}
	return fmt.Sprintf("سلام! من %s هستم ✨\n\n"+
		"کارهایت را ثبت کن، برنامه‌ی امروزت را ببین و تسک‌ها را به دیگران واگذار کن.\n"+
		"از دکمه‌های رنگی پایین استفاده کن یا همین‌جا بنویس:\n\n"+
		"➕ تسک جدید — ثبت سریع تسک\n"+
		"📋 برنامه امروز — کارهای باز امروز\n"+
		"🔍 جستجو — بین تسک‌هایت بگرد\n"+
		"👥 واگذاری تسک — سپردن کار به دیگران\n"+
		"📊 وضعیت وظایف — پیگیری کارهای واگذارشده", name)
}

// helpMessage متن راهنمای کامل بات.
const helpMessage = "ℹ️ راهنمای پلنیکس\n\n" +
	"➕ تسک جدید — ثبت سریع تسک برای خودت\n" +
	"📋 برنامه امروز — فهرست کارهای باز با فیلتر و صفحه‌بندی\n" +
	"🔍 جستجو — جستجوی عنوان بین همه‌ی تسک‌هایت\n" +
	"👥 واگذاری تسک — ثبت چند تسک برای یک نفر در یک پیام\n" +
	"📊 وضعیت وظایف دیگران — ببین هر نفر کدام تسک‌های تو را انجام داده\n" +
	"👤 پروفایل — آمار تسک‌های تو\n" +
	"⚙️ تنظیمات — خاموش/روشن کردن گزارش روزانه\n\n" +
	"از کارت هر تسک می‌توانی تیک بزنی، عنوان و توضیحات و موعد و اولویت را عوض کنی یا حذفش کنی."

// showMenu منوی اصلی را در پیام موجود نشان می‌دهد یا پیام تازه می‌فرستد.
func (b *Bot) showMenu(ctx context.Context, chatID int64, messageID int) error {
	return b.render(ctx, chatID, messageID, menuText, MenuInlineKeyboard())
}

// cbNav کال‌بک‌های دکمه‌های nav:* (ناوبری منو) را پردازش می‌کند.
func (b *Bot) cbNav(ctx context.Context, q *models.CallbackQuery) {
	data := q.Data
	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		b.answer(ctx, q, "")
		return
	}

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.log.Error("upsert user failed", zap.Error(err))
		b.answer(ctx, q, "خطا! دوباره تلاش کن")
		return
	}

	switch data {
	case "nav:noop":
		b.answer(ctx, q, "")

	case "nav:menu":
		b.answer(ctx, q, "")
		if err := b.showMenu(ctx, chatID, messageID); err != nil {
			b.log.Error("show menu failed", zap.Error(err))
		}

	case "nav:today":
		b.answer(ctx, q, "")
		if err := b.renderTaskList(ctx, chatID, messageID, u.ID, filterPending, 1); err != nil {
			b.log.Error("render task list failed", zap.Error(err))
		}

	case "nav:new":
		b.answer(ctx, q, "")
		b.startNewTask(ctx, u, chatID, messageID)

	case "nav:search":
		b.answer(ctx, q, "")
		b.startSearch(ctx, u, chatID, messageID)

	case "nav:assign":
		b.answer(ctx, q, "")
		b.startAssign(ctx, u, chatID, messageID)

	case "nav:status":
		b.answer(ctx, q, "")
		b.sendStatusPick(ctx, u, chatID, messageID)

	case "nav:profile":
		b.answer(ctx, q, "")
		b.sendProfile(ctx, u, chatID, messageID)

	case "nav:help":
		b.answer(ctx, q, "")
		if err := b.render(ctx, chatID, messageID, helpMessage, MenuInlineKeyboard()); err != nil {
			b.log.Error("render help failed", zap.Error(err))
		}

	case "nav:settings":
		b.answer(ctx, q, "")
		b.showSettings(ctx, u, chatID, messageID)

	default:
		b.answer(ctx, q, "")
	}
}

// cbState لغو جریان جاری از دکمه‌ی «لغو».
func (b *Bot) cbState(ctx context.Context, q *models.CallbackQuery) {
	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, "خطا! دوباره تلاش کن")
		return
	}
	if err := b.clearSession(ctx, u.ID); err != nil {
		b.log.Error("clear session failed", zap.Error(err))
	}
	b.answer(ctx, q, "لغو شد ❌")

	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		return
	}
	if err := b.showMenu(ctx, chatID, messageID); err != nil {
		b.log.Error("show menu failed", zap.Error(err))
	}
}

// cbUserPick انتخاب کاربر هدف از کیبورد پیشنهادی و نمایش وضعیت واگذاری‌ها.
func (b *Bot) cbUserPick(ctx context.Context, q *models.CallbackQuery) {
	const prefix = "user:filter:fr:"
	if len(q.Data) <= len(prefix) {
		b.answer(ctx, q, "")
		return
	}
	username := q.Data[len(prefix):]

	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, "خطا! دوباره تلاش کن")
		return
	}
	target, err := b.users.GetByUsername(ctx, username)
	if err != nil {
		b.answerAlert(ctx, q, "یوزرنیم پیدا نشد ❌")
		return
	}
	_ = b.clearSession(ctx, u.ID)
	b.answer(ctx, q, "")

	chatID, messageID, ok := cbOrigin(q)
	if !ok {
		chatID = u.TelegramID
	}
	if err := b.renderDelegatedStatus(ctx, chatID, messageID, u.ID, target.ID); err != nil {
		b.log.Error("render delegated status failed", zap.Error(err))
	}
}

// cbSettings تغییر تنظیمات کاربر.
func (b *Bot) cbSettings(ctx context.Context, q *models.CallbackQuery) {
	u, err := b.upsertUser(ctx, q.From)
	if err != nil {
		b.answer(ctx, q, "خطا! دوباره تلاش کن")
		return
	}

	switch q.Data {
	case "settings:daily":
		newValue := !u.DailyReport
		if err := b.users.SetDailyReport(ctx, u.ID, newValue); err != nil {
			b.log.Error("set daily report failed", zap.Error(err))
			b.answer(ctx, q, "خطا! دوباره تلاش کن")
			return
		}
		u.DailyReport = newValue
		if newValue {
			b.answer(ctx, q, "گزارش روزانه روشن شد ✅")
		} else {
			b.answer(ctx, q, "گزارش روزانه خاموش شد ❌")
		}
		chatID, messageID, ok := cbOrigin(q)
		if !ok {
			return
		}
		b.showSettings(ctx, u, chatID, messageID)
	}
}

// showSettings صفحه‌ی تنظیمات را رندر می‌کند.
func (b *Bot) showSettings(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	text := "⚙️ تنظیمات\n\nوضعیت فعلی:" +
		fmt.Sprintf("\nگزارش روزانه: %s", dailyReportLabel(u.DailyReport))
	if err := b.render(ctx, chatID, messageID, text, SettingsInlineKeyboard(u.DailyReport)); err != nil {
		b.log.Error("render settings failed", zap.Error(err))
	}
}

func dailyReportLabel(on bool) string {
	if on {
		return "روشن ✅"
	}
	return "خاموش ❌"
}

// startNewTask شروع جریان ثبت تسک جدید.
func (b *Bot) startNewTask(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingTaskTitle, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, newTaskPrompt, CancelInlineKeyboard()); err != nil {
		b.log.Error("render new task prompt failed", zap.Error(err))
	}
}

// startSearch شروع جریان جستجو.
func (b *Bot) startSearch(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingSearch, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, searchPrompt, CancelInlineKeyboard()); err != nil {
		b.log.Error("render search prompt failed", zap.Error(err))
	}
}

// startAssign شروع جریان واگذاری تسک.
func (b *Bot) startAssign(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingTaskForOther, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, assignPrompt, CancelInlineKeyboard()); err != nil {
		b.log.Error("render assign prompt failed", zap.Error(err))
	}
}

// sendStatusPick کیبورد پیشنهاد کاربران برای پیگیری وضعیت را نشان می‌دهد.
func (b *Bot) sendStatusPick(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	if err := b.setSession(ctx, u.ID, stateWaitingStatusForOther, sessionData{}); err != nil {
		b.log.Error("set session failed", zap.Error(err))
	}
	users, err := b.users.ListUsersAssignedTo(ctx, u.ID)
	if err != nil {
		b.log.Error("list assigned users failed", zap.Error(err))
	}
	if err := b.render(ctx, chatID, messageID, statusPickPrompt, SuggestFriendInlineKeyboard(users)); err != nil {
		b.log.Error("render status pick failed", zap.Error(err))
	}
}

// sendProfile پروفایل و آمار کاربر را رندر می‌کند.
func (b *Bot) sendProfile(ctx context.Context, u *domain.User, chatID int64, messageID int) {
	st, err := b.tasks.Stats(ctx, u.ID)
	if err != nil {
		b.log.Error("task stats failed", zap.Error(err))
	}

	name := (u.FirstName + " " + u.LastName)
	if name == " " {
		name = "کاربر پلنیکس"
	}
	username := "ندارد"
	if u.Username != "" {
		username = "@" + u.Username
	}

	text := "👤 پروفایل\n\n" +
		fmt.Sprintf("🪪 %s\n", name) +
		fmt.Sprintf("🔗 یوزرنیم: %s\n", username) +
		fmt.Sprintf("⏳ تسک‌های باز: %d\n", st.Pending) +
		fmt.Sprintf("✅ انجام‌شده: %d\n", st.Completed) +
		fmt.Sprintf("❌ لغو‌شده: %d\n", st.Cancelled) +
		fmt.Sprintf("📈 پیشرفت: %.0f%%", st.CompletionRate())

	if err := b.render(ctx, chatID, messageID, text, MenuInlineKeyboard()); err != nil {
		b.log.Error("render profile failed", zap.Error(err))
	}
}

// renderDelegatedStatus وضعیت تسک‌های واگذارشده به کاربر هدف را در یک پیام نشان می‌دهد.
func (b *Bot) renderDelegatedStatus(ctx context.Context, chatID int64, messageID int, ownerID, targetID uuid.UUID) error {
	tasks, err := b.tasks.ListByOwnerAndAssignee(ctx, ownerID, targetID)
	if err != nil {
		return err
	}
	target, err := b.users.GetByID(ctx, targetID)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		text := fmt.Sprintf("📭 تا کنون به %s تسکی واگذار نکرده‌ای.", target.FirstName)
		return b.render(ctx, chatID, messageID, text, BackKeyboard())
	}

	done := ""
	pending := ""
	for _, t := range tasks {
		if t.Status == "completed" {
			done += FormatSmallInfo(&t) + "\n"
		} else {
			pending += FormatSmallInfo(&t) + "\n"
		}
	}

	text := fmt.Sprintf("📊 وضعیت تسک‌های واگذارشده به %s:\n\n", target.FirstName)
	if pending != "" {
		text += "⏳ باز:\n" + pending + "\n"
	}
	if done != "" {
		text += "✅ انجام‌شده:\n" + done
	}
	return b.render(ctx, chatID, messageID, text, BackKeyboard())
}
