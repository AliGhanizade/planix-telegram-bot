package ui

import (
	"fmt"
	"time"

	"github.com/go-telegram/bot/models"
)

// All bot texts in persian and english. The user language picks the branch.

// WelcomeMessage builds the /start welcome message.
func WelcomeMessage(me *models.User, l Lang) string {
	name := "Planix"
	if l == Fa {
		name = "پلنیکس"
	}
	if me != nil && me.FirstName != "" {
		name = me.FirstName
	}
	if l == En {
		return fmt.Sprintf("Hi! I'm %s ✨\n\n"+
			"Save your tasks, see today's plan and delegate work to others.\n"+
			"Use the colored buttons below or just type:\n\n"+
			"➕ New task — quick task entry\n"+
			"📋 Today — open tasks for today\n"+
			"🔍 Search — search your tasks\n"+
			"👥 Delegate — hand work to someone else\n"+
			"📊 Delegated status — track what you assigned", name)
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

// MenuText is the inline menu header.
func MenuText(l Lang) string {
	if l == En {
		return "🗂 Planix menu\n\nPick an option:"
	}
	return "🗂 منوی پلنیکس\n\nیک گزینه را انتخاب کن:"
}

// HelpMessage is the full help text.
func HelpMessage(l Lang) string {
	if l == En {
		return "ℹ️ Planix guide\n\n" +
			"➕ New task — quick task entry\n" +
			"📋 Today — open tasks with filters and pages\n" +
			"🔍 Search — search task titles\n" +
			"👥 Delegate — assign several tasks in one message\n" +
			"📊 Delegated status — see what each person completed\n" +
			"👤 Profile — your task stats\n" +
			"⚙️ Settings — daily report, profile info, web panel link\n\n" +
			"From the task card you can tick it, edit title, description, due date and priority, or delete it."
	}
	return "ℹ️ راهنمای پلنیکس\n\n" +
		"➕ تسک جدید — ثبت سریع تسک برای خودت\n" +
		"📋 برنامه امروز — فهرست کارهای باز با فیلتر و صفحه‌بندی\n" +
		"🔍 جستجو — جستجوی عنوان بین همه‌ی تسک‌هایت\n" +
		"👥 واگذاری تسک — ثبت چند تسک برای یک نفر در یک پیام\n" +
		"📊 وضعیت وظایف دیگران — ببین هر نفر کدام تسک‌های تو را انجام داده\n" +
		"👤 پروفایل — آمار تسک‌های تو\n" +
		"⚙️ تنظیمات — گزارش روزانه، ویرایش اطلاعات و اتصال پنل وب\n\n" +
		"از کارت هر تسک می‌توانی تیک بزنی، عنوان و توضیحات و موعد و اولویت را عوض کنی یا حذفش کنی."
}

// AssignPrompt explains the delegation input format.
func AssignPrompt(l Lang) string {
	if l == En {
		return "📋 Input format\n\n" +
			"First line: username\n" +
			"Then: one task per line\n\n" +
			"@username\n" +
			"Design the API\n" +
			"Review the pull request"
	}
	return "📋 فرمت ارسال\n\n" +
		"خط اول: یوزرنیم\n" +
		"بقیه: هر تسک در یک خط\n\n" +
		"@username\n" +
		"طراحی API\n" +
		"بررسی Pull Request"
}

// SearchPrompt asks for a search query.
func SearchPrompt(l Lang) string {
	if l == En {
		return "🔍 Send a word from the task title:"
	}
	return "🔍 متن یا کلمه‌ای از عنوان تسک را بفرست:"
}

// NewTaskPrompt asks for a task title.
func NewTaskPrompt(l Lang) string {
	if l == En {
		return "➕ Send the task title. Example: study Go"
	}
	return "➕ عنوان تسک را بفرست. مثال: مطالعه گولنگ"
}

// StatusPickPrompt asks which user to track.
func StatusPickPrompt(l Lang) string {
	if l == En {
		return "📊 Whose delegated tasks do you want to see?"
	}
	return "📊 برای چه کسی می‌خوای وضعیت تسک‌هاشو ببینی؟"
}

// WebCodeMessage builds the web login code message.
func WebCodeMessage(code string, minutes int, l Lang) string {
	if l == En {
		return fmt.Sprintf("🌐 Planix web panel login code\n\n"+
			"Code: %s\n\n"+
			"It is valid for %d minutes and works only once.\n"+
			"Ignore this message if you didn't request it.", code, minutes)
	}
	return fmt.Sprintf("🌐 کد ورود پنل وب پلنیکس\n\n"+
		"کد ورود: %s\n\n"+
		"این کد %d دقیقه اعتبار دارد و فقط یک‌بار قابل استفاده است.\n"+
		"اگر شما این درخواست را نداده بودید، پیام را نادیده بگیرید.", code, minutes)
}

// DailyReportLabel renders the on/off label for the daily report.
func DailyReportLabel(on bool, l Lang) string {
	if l == En {
		if on {
			return "On ✅"
		}
		return "Off ❌"
	}
	if on {
		return "روشن ✅"
	}
	return "خاموش ❌"
}

// LanguageLabel is the language row in settings.
func LanguageLabel(l Lang) string {
	if l == En {
		return "🌐 Language: English"
	}
	return "🌐 زبان: فارسی"
}

// SupportMessage builds the support reply.
func SupportMessage(owner string, l Lang) string {
	if l == En {
		return fmt.Sprintf("🛟 For support, message @%s.", owner)
	}
	return fmt.Sprintf("🛟 برای ارتباط با پشتیبانی به @%s پیام بده.", owner)
}

// ErrGeneric is the generic failure toast.
func ErrGeneric(l Lang) string {
	if l == En {
		return "Something went wrong, try again"
	}
	return "خطا! دوباره تلاش کن"
}

// CancelledToast is shown on the cancel button.
func CancelledToast(l Lang) string {
	if l == En {
		return "Cancelled ❌"
	}
	return "لغو شد ❌"
}

// TaskSaved confirms a created task.
func TaskSaved(title string, l Lang) string {
	if l == En {
		return fmt.Sprintf("✅ Task «%s» saved.\nTick it off from today's plan whenever it's done.", title)
	}
	return fmt.Sprintf("✅ تسک «%s» ثبت شد.\nهر زمان انجامش دادی از برنامه‌ی امروز تیکش بزن.", title)
}

// TaskDoneToast is the toast after completing a task.
func TaskDoneToast(l Lang) string {
	if l == En {
		return "Task done ✅"
	}
	return "تسک انجام شد ✅"
}

// TaskReopenedToast is the toast after reopening a task.
func TaskReopenedToast(l Lang) string {
	if l == En {
		return "Task reopened 🔄"
	}
	return "تسک دوباره باز شد 🔄"
}

// RefreshedToast is the toast after a refresh.
func RefreshedToast(l Lang) string {
	if l == En {
		return "Refreshed 🔄"
	}
	return "بروزرسانی شد 🔄"
}

// DeletedToast is the toast after deleting a task.
func DeletedToast(l Lang) string {
	if l == En {
		return "Deleted 🗑"
	}
	return "حذف شد 🗑"
}

// TaskNotFoundToast is shown when a task is gone.
func TaskNotFoundToast(l Lang) string {
	if l == En {
		return "Task not found"
	}
	return "تسک پیدا نشد"
}

// InvalidTaskIDToast is shown for a malformed callback id.
func InvalidTaskIDToast(l Lang) string {
	if l == En {
		return "Invalid task id"
	}
	return "شناسه‌ی تسک نامعتبر است"
}

// OwnerDoneNotify tells the owner a delegated task was completed.
func OwnerDoneNotify(actor, title string, l Lang) string {
	if l == En {
		return fmt.Sprintf("📣 %s completed the task «%s».", actor, title)
	}
	return fmt.Sprintf("📣 %s تسک «%s» را انجام داد.", actor, title)
}

// OwnerReopenedNotify tells the owner a delegated task was reopened.
func OwnerReopenedNotify(actor, title string, l Lang) string {
	if l == En {
		return fmt.Sprintf("📣 %s reopened the task «%s».", actor, title)
	}
	return fmt.Sprintf("📣 %s تسک «%s» را بازگشایی کرد.", actor, title)
}

// AssignedNotify tells the assignee about new delegated tasks.
func AssignedNotify(owner string, count int, l Lang) string {
	if l == En {
		return fmt.Sprintf("📣 Planix update\n%s assigned %d tasks to you:", owner, count)
	}
	return fmt.Sprintf("📣 گزارش پلنیکس\n%s %d تسک برای تو ثبت کرد:", owner, count)
}

// AssignedConfirm tells the owner the delegation worked.
func AssignedConfirm(count int, assignee string, l Lang) string {
	if l == En {
		return fmt.Sprintf("%d tasks assigned to %s successfully ✅", count, assignee)
	}
	return fmt.Sprintf("ثبت %d تسک برای %s با موفقیت انجام شد ✅", count, assignee)
}

// UsernameNotFound is the reply for an unknown username.
func UsernameNotFound(l Lang) string {
	if l == En {
		return "Username not found. Please try again."
	}
	return "یوزرنیم پیدا نشد. لطفا دوباره امتحان کن."
}

// UsernameNotFoundToast is the toast variant.
func UsernameNotFoundToast(l Lang) string {
	if l == En {
		return "Username not found ❌"
	}
	return "یوزرنیم پیدا نشد ❌"
}

// SelfAssignError blocks delegating to yourself.
func SelfAssignError(l Lang) string {
	if l == En {
		return "You can't delegate a task to yourself 🙂"
	}
	return "نمی‌تونی تسک رو به خودت واگذار کنی 🙂"
}

// AssignFormatError asks for the right delegation format.
func AssignFormatError(l Lang) string {
	if l == En {
		return "Send the username on the first line and each task on its own line."
	}
	return "لطفا یوزرنیم را در خط اول و هر تسک را در یک خط جدا بفرست."
}

// DelegatedEmpty says there are no delegated tasks yet.
func DelegatedEmpty(target string, l Lang) string {
	if l == En {
		return fmt.Sprintf("You haven't delegated any tasks to %s yet.", target)
	}
	return fmt.Sprintf("تا کنون به %s تسکی واگذار نکرده‌ای.", target)
}

// DelegatedHeader opens the delegated status message.
func DelegatedHeader(target string, l Lang) string {
	if l == En {
		return fmt.Sprintf("📊 Status of tasks delegated to %s:\n\n", target)
	}
	return fmt.Sprintf("📊 وضعیت تسک‌های واگذارشده به %s:\n\n", target)
}

// DelegatedOpenTitle is the open section title.
func DelegatedOpenTitle(l Lang) string {
	if l == En {
		return "⏳ Open:\n"
	}
	return "⏳ باز:\n"
}

// DelegatedDoneTitle is the done section title.
func DelegatedDoneTitle(l Lang) string {
	if l == En {
		return "✅ Done:\n"
	}
	return "✅ انجام‌شده:\n"
}

// ProfileText builds the profile message.
func ProfileText(name, username string, pending, completed, cancelled int64, rate float64, l Lang) string {
	if l == En {
		return fmt.Sprintf("👤 Profile\n\n"+
			"🪪 %s\n"+
			"🔗 Username: %s\n"+
			"⏳ Open tasks: %d\n"+
			"✅ Completed: %d\n"+
			"❌ Cancelled: %d\n"+
			"📈 Progress: %.0f%%", name, username, pending, completed, cancelled, rate)
	}
	return fmt.Sprintf("👤 پروفایل\n\n"+
		"🪪 %s\n"+
		"🔗 یوزرنیم: %s\n"+
		"⏳ تسک‌های باز: %d\n"+
		"✅ انجام‌شده: %d\n"+
		"❌ لغو‌شده: %d\n"+
		"📈 پیشرفت: %.0f%%", name, username, pending, completed, cancelled, rate)
}

// ProfileDefaultName is used when the user has no name.
func ProfileDefaultName(l Lang) string {
	if l == En {
		return "A planix user"
	}
	return "کاربر پلنیکس"
}

// NoUsername is shown when the user has no username.
func NoUsername(l Lang) string {
	if l == En {
		return "none"
	}
	return "ندارد"
}

// SettingsText builds the settings screen.
func SettingsText(dailyLabel string, l Lang) string {
	if l == En {
		return fmt.Sprintf("⚙️ Settings\n\nCurrent state:\nDaily report: %s", dailyLabel)
	}
	return fmt.Sprintf("⚙️ تنظیمات\n\nوضعیت فعلی:\nگزارش روزانه: %s", dailyLabel)
}

// ProfileEditText builds the profile edit menu screen.
func ProfileEditText(first, last, tz string, l Lang) string {
	if l == En {
		return fmt.Sprintf("✏️ Edit profile\n\n"+
			"🪪 First name: %s\n"+
			"🏷 Last name: %s\n"+
			"🌍 Timezone: %s\n\n"+
			"Tap a field and send the new value.", first, last, tz)
	}
	return fmt.Sprintf("✏️ ویرایش اطلاعات\n\n"+
		"🪪 نام: %s\n"+
		"🏷 نام خانوادگی: %s\n"+
		"🌍 تایم‌زون: %s\n\n"+
		"روی فیلد موردنظر بزن و مقدار جدید را بفرست.", first, last, tz)
}

// EditFirstNamePrompt asks for a new first name.
func EditFirstNamePrompt(l Lang) string {
	if l == En {
		return "🪪 Send your new first name:"
	}
	return "🪪 نام جدیدت را بفرست:"
}

// EditLastNamePrompt asks for a new last name.
func EditLastNamePrompt(l Lang) string {
	if l == En {
		return "🏷 Send the new last name (send - to clear it):"
	}
	return "🏷 نام خانوادگی جدید را بفرست (برای حذف، «-» بفرست):"
}

// EditTimezonePrompt asks for a timezone.
func EditTimezonePrompt(l Lang) string {
	if l == En {
		return "🌍 Send your timezone, e.g. Asia/Tehran"
	}
	return "🌍 تایم‌زون را بفرست؛ مثال: Asia/Tehran"
}

// ValidationError is the reply for an invalid profile value.
func ValidationError(l Lang) string {
	if l == En {
		return "⚠️ Invalid value, please try again"
	}
	return "⚠️ مقدار معتبر نیست؛ دوباره تلاش کن"
}

// IssueCodeError is the reply when code issuance fails.
func IssueCodeError(l Lang) string {
	if l == En {
		return "Could not issue a code, please try again."
	}
	return "خطا در صدور کد؛ دوباره تلاش کن."
}

// SearchResultsHeader opens the search results message.
func SearchResultsHeader(query string, l Lang) string {
	if l == En {
		return fmt.Sprintf("🔍 Search results for «%s»:\n\n", query)
	}
	return fmt.Sprintf("🔍 نتایج جستجو برای «%s»:\n\n", query)
}

// SearchEmpty is shown when search finds nothing.
func SearchEmpty(l Lang) string {
	if l == En {
		return "Nothing found 🤷"
	}
	return "چیزی پیدا نشد 🤷"
}

// ListHeader opens the task list message.
func ListHeader(label string, page, pages int, l Lang) string {
	if l == En {
		return fmt.Sprintf("📋 %s — page %d of %d\n\n", label, page, pages)
	}
	return fmt.Sprintf("📋 %s — صفحه‌ی %d از %d\n\n", label, page, pages)
}

// CardHeader opens the task card message.
func CardHeader(l Lang) string {
	if l == En {
		return "🗂 Task card\n\n"
	}
	return "🗂 کارت تسک\n\n"
}

// DeleteConfirmText builds the delete confirmation.
func DeleteConfirmText(title string, l Lang) string {
	if l == En {
		return fmt.Sprintf("🗑 Delete the task «%s»?\nThis cannot be undone.", title)
	}
	return fmt.Sprintf("🗑 مطمئنی می‌خوای تسک «%s» را حذف کنی؟\nاین عمل قابل بازگشت نیست.", title)
}

// DeletedMessage is shown after a delete.
func DeletedMessage(l Lang) string {
	if l == En {
		return "🗑 Task deleted."
	}
	return "🗑 تسک حذف شد."
}

// EditTitlePrompt asks for the new title.
func EditTitlePrompt(l Lang) string {
	if l == En {
		return "📝 Send the new task title in your next message."
	}
	return "📝 لطفاً عنوان جدید تسک را در پیام بعدی بفرست."
}

// EditTitleToast is the toast for the title edit flow.
func EditTitleToast(l Lang) string {
	if l == En {
		return "Send the new title ✏️"
	}
	return "عنوان جدید را ارسال کنید ✏️"
}

// EditDescPrompt asks for the new description.
func EditDescPrompt(l Lang) string {
	if l == En {
		return "📄 Send the new description in your next message."
	}
	return "📄 لطفاً توضیحات جدید تسک را در پیام بعدی بفرست."
}

// EditDescToast is the toast for the description edit flow.
func EditDescToast(l Lang) string {
	if l == En {
		return "Send the new description ✏️"
	}
	return "توضیحات جدید را ارسال کنید ✏️"
}

// DuePickerText opens the due date picker.
func DuePickerText(title string, l Lang) string {
	if l == En {
		return fmt.Sprintf("📅 Pick a due date for «%s»:", title)
	}
	return fmt.Sprintf("📅 موعد تسک «%s» را انتخاب کن:", title)
}

// DueSetToast confirms the new due date.
func DueSetToast(due string, l Lang) string {
	if l == En {
		return "Due date saved 📅 " + due
	}
	return "موعد ثبت شد 📅 " + due
}

// DueClearedToast confirms the due date was removed.
func DueClearedToast(l Lang) string {
	if l == En {
		return "Due date removed 🗓"
	}
	return "موعد حذف شد 🗓"
}

// PriorityPickerText opens the priority picker.
func PriorityPickerText(title string, l Lang) string {
	if l == En {
		return fmt.Sprintf("⚡ Pick a priority for «%s»:", title)
	}
	return fmt.Sprintf("⚡ اولویت تسک «%s» را انتخاب کن:", title)
}

// PrioritySetToast confirms the new priority.
func PrioritySetToast(label string, l Lang) string {
	if l == En {
		return "Priority saved ⚡ " + label
	}
	return "اولویت ثبت شد ⚡ " + label
}

// TaskTitleUpdated confirms a title edit.
func TaskTitleUpdated(l Lang) string {
	if l == En {
		return "✏️ Task title updated."
	}
	return "✏️ عنوان تسک بروزرسانی شد."
}

// TaskDescUpdated confirms a description edit.
func TaskDescUpdated(l Lang) string {
	if l == En {
		return "📄 Task description updated."
	}
	return "📄 توضیحات تسک بروزرسانی شد."
}

// DailyReportHeader opens the daily report.
func DailyReportHeader(l Lang) string {
	if l == En {
		return "🌅 Planix daily report\n\n"
	}
	return "🌅 گزارش روزانه پلنیکس\n\n"
}

// DailyReportEmpty is the empty daily report body.
func DailyReportEmpty(l Lang) string {
	if l == En {
		return "No open tasks today; have a great day ✨"
	}
	return "امروز تسک بازی نداری؛ روز خوبی داشته باشی ✨"
}

// DailyReportCount opens the open tasks section of the report.
func DailyReportCount(count int, l Lang) string {
	if l == En {
		return fmt.Sprintf("You have %d open tasks:\n\n", count)
	}
	return fmt.Sprintf("شما %d تسک باز دارید:\n\n", count)
}

// ReminderText builds the due date reminder.
func ReminderText(title, human, due string, l Lang) string {
	if l == En {
		return fmt.Sprintf("⏳ Planix reminder\n\nThe task «%s» is due %s.\n📅 Due: %s", title, human, due)
	}
	return fmt.Sprintf("⏳ یادآوری پلنیکس\n\nتسک «%s» تا %s دیگر موعدش تمام می‌شود.\n📅 موعد: %s", title, human, due)
}

// HumanDuration renders a duration in words.
func HumanDuration(d time.Duration, l Lang) string {
	if d <= 0 {
		if l == En {
			return "right now"
		}
		return "همین حالا"
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		if l == En {
			return fmt.Sprintf("%d minutes", minutes)
		}
		return fmt.Sprintf("%d دقیقه", minutes)
	}
	hours := int(d.Hours())
	if hours < 24 {
		if l == En {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d ساعت", hours)
	}
	if l == En {
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	}
	return fmt.Sprintf("%d روز", int(d.Hours()/24))
}
