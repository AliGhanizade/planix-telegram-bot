package bot

import (
	"time"
)

// dueFromPreset زمان موعد را بر اساس انتخاب سریع کاربر محاسبه می‌کند.
// خروجی دوم یعنی پریست شناخته‌شده است.
func dueFromPreset(preset string, now time.Time, loc *time.Location) (time.Time, bool) {
	local := now.In(loc)
	switch preset {
	case "today":
		y, m, d := local.Date()
		return time.Date(y, m, d, 23, 59, 0, 0, loc), true
	case "tomorrow":
		next := local.AddDate(0, 0, 1)
		y, m, d := next.Date()
		return time.Date(y, m, d, 9, 0, 0, 0, loc), true
	case "week":
		next := local.AddDate(0, 0, 7)
		y, m, d := next.Date()
		return time.Date(y, m, d, 9, 0, 0, 0, loc), true
	default:
		return time.Time{}, false
	}
}

// userLocation تایم‌زون کاربر را برمی‌گرداند؛ در خطا UTC.
func userLocation(tz string) *time.Location {
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	return time.UTC
}
