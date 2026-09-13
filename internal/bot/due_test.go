package bot

import (
	"testing"
	"time"
)

func TestDueFromPreset(t *testing.T) {
	// a fixed reference point for tests.
	now := time.Date(2026, 9, 13, 15, 30, 0, 0, time.UTC)
	loc := time.UTC

	today, ok := dueFromPreset("today", now, loc)
	if !ok {
		t.Fatal("today preset should be valid")
	}
	if today.Format("2006-01-02 15:04") != "2026-09-13 23:59" {
		t.Errorf("today preset = %s, want end of same day", today)
	}

	tomorrow, ok := dueFromPreset("tomorrow", now, loc)
	if !ok {
		t.Fatal("tomorrow preset should be valid")
	}
	if tomorrow.Format("2006-01-02 15:04") != "2026-09-14 09:00" {
		t.Errorf("tomorrow preset = %s, want 9am next day", tomorrow)
	}

	week, ok := dueFromPreset("week", now, loc)
	if !ok {
		t.Fatal("week preset should be valid")
	}
	if week.Format("2006-01-02 15:04") != "2026-09-20 09:00" {
		t.Errorf("week preset = %s, want 9am in 7 days", week)
	}

	if _, ok := dueFromPreset("bogus", now, loc); ok {
		t.Error("unknown preset should not be valid")
	}
}

func TestDueFromPresetTimezone(t *testing.T) {
	now := time.Date(2026, 9, 13, 20, 0, 0, 0, time.UTC) // ۲۳:۳۰ به وقت تهران
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Skip("timezone data unavailable")
	}
	today, ok := dueFromPreset("today", now, loc)
	if !ok {
		t.Fatal("today preset should be valid")
	}
	if today.In(loc).Format("2006-01-02 15:04") != "2026-09-13 23:59" {
		t.Errorf("tehran end of day = %s", today.In(loc))
	}
}

func TestUserLocationFallback(t *testing.T) {
	if got := userLocation("Asia/Tehran"); got.String() != "Asia/Tehran" {
		t.Errorf("userLocation(Asia/Tehran) = %s", got)
	}
	if got := userLocation("Not/AZone"); got != time.UTC {
		t.Errorf("userLocation fallback should be UTC, got %s", got)
	}
}
