package ui

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

// bot message texts.
const (
	MenuText = "🗂 منوی پلنیکس\n\nیک گزینه را انتخاب کن:"

	AssignPrompt = "📋 فرمت ارسال\n\n" +
		"خط اول: یوزرنیم\n" +
		"بقیه: هر تسک در یک خط\n\n" +
		"@username\n" +
		"طراحی API\n" +
		"بررسی Pull Request"

	SearchPrompt  = "🔍 متن یا کلمه‌ای از عنوان تسک را بفرست:"
	NewTaskPrompt = "➕ عنوان تسک را بفرست. مثال: مطالعه گولنگ"

	StatusPickPrompt = "📊 برای چه کسی می‌خوای وضعیت تسک‌هاشو ببینی؟"
)

// WelcomeMessage builds the /start welcome message.
func WelcomeMessage(me *models.User) string {
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

// HelpMessage is the full help text.
const HelpMessage = "ℹ️ راهنمای پلنیکس\n\n" +
	"➕ تسک جدید — ثبت سریع تسک برای خودت\n" +
	"📋 برنامه امروز — فهرست کارهای باز با فیلتر و صفحه‌بندی\n" +
	"🔍 جستجو — جستجوی عنوان بین همه‌ی تسک‌هایت\n" +
	"👥 واگذاری تسک — ثبت چند تسک برای یک نفر در یک پیام\n" +
	"📊 وضعیت وظایف دیگران — ببین هر نفر کدام تسک‌های تو را انجام داده\n" +
	"👤 پروفایل — آمار تسک‌های تو\n" +
	"⚙️ تنظیمات — گزارش روزانه، ویرایش اطلاعات و اتصال پنل وب\n\n" +
	"از کارت هر تسک می‌توانی تیک بزنی، عنوان و توضیحات و موعد و اولویت را عوض کنی یا حذفش کنی."

// WebCodeMessage builds the web login code message.
func WebCodeMessage(code string, minutes int) string {
	return fmt.Sprintf("🌐 کد ورود پنل وب پلنیکس\n\n"+
		"کد ورود: %s\n\n"+
		"این کد %d دقیقه اعتبار دارد و فقط یک‌بار قابل استفاده است.\n"+
		"اگر شما این درخواست را نداده بودید، پیام را نادیده بگیرید.", code, minutes)
}

// DailyReportLabel renders the on/off label for the daily report.
func DailyReportLabel(on bool) string {
	if on {
		return "روشن ✅"
	}
	return "خاموش ❌"
}
