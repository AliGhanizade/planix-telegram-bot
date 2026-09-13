package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
)

func TestPriorityLabel(t *testing.T) {
	if got := PriorityLabel("low", Fa); got != "🟢 کم" {
		t.Errorf("PriorityLabel(low, fa) = %q", got)
	}
	if got := PriorityLabel("urgent", En); got != "🔴 Urgent" {
		t.Errorf("PriorityLabel(urgent, en) = %q", got)
	}
	if got := PriorityLabel("unknown", En); got != "unknown" {
		t.Errorf("PriorityLabel fallback = %q, want raw value", got)
	}
}

func TestStatusLabel(t *testing.T) {
	if got := StatusLabel("pending", Fa); got != "⏳ در انتظار" {
		t.Errorf("StatusLabel(pending, fa) = %q", got)
	}
	if got := StatusLabel("pending", En); got != "⏳ Pending" {
		t.Errorf("StatusLabel(pending, en) = %q", got)
	}
	if got := StatusLabel("mystery", En); got != "mystery" {
		t.Errorf("StatusLabel fallback = %q", got)
	}
}

func TestFormatSmallInfo(t *testing.T) {
	due := time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC)
	task := &domain.Task{Title: "طراحی API", Priority: "high", Status: "pending", DueAt: &due}
	got := FormatSmallInfo(task, Fa)

	for _, want := range []string{"⏳", "طراحی API", "🟠 زیاد", "📅", "09-20 09:00"} {
		if !strings.Contains(got, want) {
			t.Errorf("FormatSmallInfo missing %q in %q", want, got)
		}
	}

	// no due date means no timestamp in the output.
	task.DueAt = nil
	if strings.Contains(FormatSmallInfo(task, Fa), "📅") {
		t.Error("FormatSmallInfo should not contain due emoji when DueAt is nil")
	}
}

func TestFormatTask(t *testing.T) {
	task := &domain.Task{Title: "بررسی PR", Description: "", Priority: "urgent", Status: "completed"}
	got := FormatTask(task, Fa)

	for _, want := range []string{"بررسی PR", "ندارد", "🔴 فوری", "✅ انجام شده", "❌"} {
		if !strings.Contains(got, want) {
			t.Errorf("FormatTask missing %q", want)
		}
	}

	en := FormatTask(task, En)
	for _, want := range []string{"Title:", "Description:", "none", "🔴 Urgent", "✅ Completed"} {
		if !strings.Contains(en, want) {
			t.Errorf("FormatTask(en) missing %q", want)
		}
	}
}

func TestTruncate(t *testing.T) {
	short := "short task"
	if got := Truncate(short, 20); got != short {
		t.Errorf("Truncate should not change short strings, got %q", got)
	}
	long := strings.Repeat("a", 50)
	got := Truncate(long, 10)
	if want := 10; len([]rune(got)) != want {
		t.Errorf("Truncate(long, %d) has %d runes, want %d", 10, len([]rune(got)), want)
	}
	if !strings.HasSuffix(got, "…") {
		t.Error("truncated string should end with ellipsis")
	}
}

func TestNormalize(t *testing.T) {
	if Normalize("") != Fa {
		t.Error("empty language should default to fa")
	}
	if Normalize("en") != En {
		t.Error("en should map to En")
	}
	if Normalize("fr") != Fa {
		t.Error("unsupported language should default to fa")
	}
}
