package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// PageSize is the number of tasks per list page.
const PageSize = 5

// ListFilter is the status filter of the task list.
type ListFilter string

// task list filter values.
const (
	FilterPending   ListFilter = "pending"
	FilterCompleted ListFilter = "completed"
	FilterAll       ListFilter = "all"
)

// Label is the display name of the filter.
func (f ListFilter) Label(l Lang) string {
	if f == FilterCompleted {
		if l == En {
			return "Completed"
		}
		return "انجام‌شده‌ها"
	}
	if f == FilterAll {
		if l == En {
			return "All tasks"
		}
		return "همه‌ی تسک‌ها"
	}
	if l == En {
		return "Today's plan"
	}
	return "برنامه امروز"
}

// EmptyText is the message shown for an empty list.
func (f ListFilter) EmptyText(l Lang) string {
	if l == En {
		switch f {
		case FilterCompleted:
			return "You haven't finished anything yet; now is a good start 💪"
		case FilterAll:
			return "No tasks yet. Start with ➕ New task!"
		default:
			return "No open tasks today; a fresh start awaits ✨"
		}
	}
	switch f {
	case FilterCompleted:
		return "هنوز تسکی را تمام نکرده‌ای؛ یک شروع خوب، همین حالاست 💪"
	case FilterAll:
		return "هنوز تسکی نساخته‌ای. با ➕ تسک جدید شروع کن!"
	default:
		return "امروز تسک بازی نداری؛ وقت یک شروع تازه است ✨"
	}
}

// ParseListFilter reads a filter from callback data.
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

// AtoiOr parses a page number from callback data, falling back on error.
func AtoiOr(s string, fallback int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return fallback
}

// SplitCallback splits callback data into parts.
func SplitCallback(data string) []string {
	return strings.Split(data, ":")
}

// Truncate shortens a title to fit inside a button.
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// TaskOrigin says where a task view came from: the list (with filter and page) or a standalone card.
type TaskOrigin struct {
	List   bool
	Filter string
	Page   int
}

// NoOrigin means the view was opened outside the task list.
var NoOrigin = TaskOrigin{}

// Suffix builds the origin suffix for callback data.
func (o TaskOrigin) Suffix() string {
	if !o.List {
		return ":C"
	}
	return fmt.Sprintf(":L:%s:%d", o.Filter, o.Page)
}

// ParseTaskOrigin reads the origin suffix from callback data parts.
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

// TaskData builds callback data for a task action with its origin.
func TaskData(action string, id interface{ String() string }, o TaskOrigin) string {
	return "task:" + action + ":" + id.String() + o.Suffix()
}

// ListData builds callback data for the task list.
func ListData(f ListFilter, page int) string {
	return fmt.Sprintf("list:tasks:%s:%d", f, page)
}

// BackDataOf returns the right back button callback for an origin.
func BackDataOf(o TaskOrigin) string {
	if o.List {
		return ListData(ListFilter(o.Filter), o.Page)
	}
	return "nav:menu"
}
