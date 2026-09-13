package bot

import (
	"strings"
	"testing"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

func TestPriorityLabel(t *testing.T) {
	cases := map[string]string{
		"low":    "🟢 کم",
		"normal": "🟡 معمولی",
		"high":   "🟠 زیاد",
		"urgent": "🔴 فوری",
	}
	for in, want := range cases {
		if got := priorityLabel(in); got != want {
			t.Errorf("priorityLabel(%q) = %q, want %q", in, got, want)
		}
	}
	if got := priorityLabel("unknown"); got != "unknown" {
		t.Errorf("priorityLabel fallback = %q, want raw value", got)
	}
}

func TestStatusLabel(t *testing.T) {
	if got := statusLabel("pending"); got != "⏳ در انتظار" {
		t.Errorf("statusLabel(pending) = %q", got)
	}
	if got := statusLabel("mystery"); got != "mystery" {
		t.Errorf("statusLabel fallback = %q", got)
	}
}

func TestFormatSmallInfo(t *testing.T) {
	due := time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)
	task := &domain.Task{Title: "طراحی API", Priority: "high", Status: "pending", DueAt: &due}
	got := FormatSmallInfo(task)

	for _, want := range []string{"⏳", "طراحی API", "🟠 زیاد", "📅", "09-20 09:00"} {
		if !strings.Contains(got, want) {
			t.Errorf("FormatSmallInfo missing %q in %q", want, got)
		}
	}

	// بدون موعد نباید تایم‌استمپ داشته باشد.
	task.DueAt = nil
	if strings.Contains(FormatSmallInfo(task), "📅") {
		t.Error("FormatSmallInfo should not contain due emoji when DueAt is nil")
	}
}

func TestFormatTask(t *testing.T) {
	task := &domain.Task{Title: "بررسی PR", Description: "", Priority: "urgent", Status: "completed"}
	got := FormatTask(task)

	for _, want := range []string{"بررسی PR", "ندارد", "🔴 فوری", "✅ انجام شده", "❌ خیر"} {
		if !strings.Contains(got, want) {
			t.Errorf("FormatTask missing %q", want)
		}
	}
}

func TestTruncate(t *testing.T) {
	short := "تسک کوتاه"
	if got := truncate(short, 20); got != short {
		t.Errorf("truncate should not change short strings, got %q", got)
	}
	long := strings.Repeat("الف", 50)
	got := truncate(long, 10)
	if want := 10; len([]rune(got)) != want {
		t.Errorf("truncate(%q, %d) has %d runes, want %d", "long", 10, len([]rune(got)), want)
	}
	if !strings.HasSuffix(got, "…") {
		t.Error("truncated string should end with ellipsis")
	}
}
