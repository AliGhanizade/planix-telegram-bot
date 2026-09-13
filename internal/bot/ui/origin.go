package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// PageSize تعداد تسک در هر صفحه‌ی فهرست است.
const PageSize = 5

// ListFilter فیلتر وضعیت فهرست تسک‌ها.
type ListFilter string

// مقادیر فیلتر فهرست.
const (
	FilterPending   ListFilter = "pending"
	FilterCompleted ListFilter = "completed"
	FilterAll       ListFilter = "all"
)

// Label عنوان فارسی فیلتر.
func (f ListFilter) Label() string {
	switch f {
	case FilterCompleted:
		return "انجام‌شده‌ها"
	case FilterAll:
		return "همه‌ی تسک‌ها"
	default:
		return "برنامه امروز"
	}
}

// EmptyText پیام حالت خالی فیلتر.
func (f ListFilter) EmptyText() string {
	switch f {
	case FilterCompleted:
		return "هنوز تسکی را تمام نکرده‌ای؛ یک شروع خوب، همین حالاست 💪"
	case FilterAll:
		return "هنوز تسکی نساخته‌ای. با ➕ تسک جدید شروع کن!"
	default:
		return "امروز تسک بازی نداری؛ وقت یک شروع تازه است ✨"
	}
}

// ParseListFilter فیلتر را از دیتای کال‌بک می‌خواند.
func ParseListFilter(s string) ListFilter {
	switch s {
	case string(FilterCompleted):
		return FilterCompleted
	case string(FilterAll):
		return FilterAll
	default:
		return FilterPending
	}
}

// AtoiOr عدد را از دیتای کال‌بک می‌خواند؛ در خطا مقدار پیش‌فرض.
func AtoiOr(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return fallback
}

// SplitCallback دیتای کال‌بک را به اجزا می‌شکند.
func SplitCallback(data string) []string {
	return strings.Split(data, ":")
}

// Truncate عنوان را برای جا شدن در دکمه کوتاه می‌کند.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// TaskOrigin مبدأ نمایش یک تسک را مشخص می‌کند: فهرست (با فیلتر و صفحه) یا کارت مستقل.
type TaskOrigin struct {
	List   bool
	Filter string
	Page   int
}

// NoOrigin مبدأ مستقل (کارت باز‌شده از جستجو و…).
var NoOrigin = TaskOrigin{}

// Suffix پسوند مبدأ را برای دیتای کال‌بک می‌سازد.
func (o TaskOrigin) Suffix() string {
	if !o.List {
		return ":C"
	}
	return fmt.Sprintf(":L:%s:%d", o.Filter, o.Page)
}

// ParseTaskOrigin پسوند مبدأ را از اجزای دیتای کال‌بک می‌خواند.
func ParseTaskOrigin(parts []string, idx int) (TaskOrigin, error) {
	if idx >= len(parts) {
		return NoOrigin, nil
	}
	switch parts[idx] {
	case "C":
		return NoOrigin, nil
	case "L":
		o := TaskOrigin{List: true, Filter: "pending", Page: 1}
		if idx+1 < len(parts) {
			o.Filter = parts[idx+1]
		}
		if idx+2 < len(parts) {
			if p, err := strconv.Atoi(parts[idx+2]); err == nil {
				o.Page = p
			}
		}
		return o, nil
	default:
		return NoOrigin, fmt.Errorf("unknown origin: %s", parts[idx])
	}
}

// TaskData دیتای کال‌بک یک عملیات تسک را با مبدأ می‌سازد.
func TaskData(action string, id interface{ String() string }, o TaskOrigin) string {
	return "task:" + action + ":" + id.String() + o.Suffix()
}

// ListData دیتای کال‌بک فهرست تسک‌ها را می‌سازد.
func ListData(f ListFilter, page int) string {
	return fmt.Sprintf("list:tasks:%s:%d", f, page)
}

// BackDataOf دکمه‌ی بازگشتِ مناسب برای یک مبدأ را برمی‌گرداند.
func BackDataOf(o TaskOrigin) string {
	if o.List {
		return ListData(ListFilter(o.Filter), o.Page)
	}
	return "nav:menu"
}
