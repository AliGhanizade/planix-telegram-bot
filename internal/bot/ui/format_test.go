package ui

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
		if got := PriorityLabel(in); got != want {
			t.Errorf("PriorityLabel(%q) = %q, want %q", in, got, want)
		}
	}
	if got := PriorityLabel("unknown"); got != "unknown" {
		t.Errorf("PriorityLabel fallback = %q, want raw value", got)
	}
}

func TestStatusLabel(t *testing.T) {
	if got := StatusLabel("pending"); got != "⏳ در انتظار" {
		t.Errorf("StatusLabel(pending) = %q", got)
	}
	if got := StatusLabel("mystery"); got != "mystery" {
		t.Errorf("StatusLabel fallback = %q", got)
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
	if got := Truncate(short, 20); got != short {
		t.Errorf("Truncate should not change short strings, got %q", got)
	}
	long := strings.Repeat("الف", 50)
	got := Truncate(long, 10)
	if want := 10; len([]rune(got)) != want {
		t.Errorf("Truncate(long, %d) has %d runes, want %d", 10, len([]rune(got)), want)
	}
	if !strings.HasSuffix(got, "…") {
		t.Error("truncated string should end with ellipsis")
	}
}
