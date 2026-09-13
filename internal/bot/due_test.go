package bot

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDueFromPreset(t *testing.T) {
	// یک لحظه‌ی مرجع ثابت برای تست.
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
	now := time.Date(2026, 9, 13, 20, 0, 0, 0, time.UTC) // ساعت ۲۳:۳۰ به وقت تهران
	loc, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Skip("timezone data unavailable")
	}
	today, ok := dueFromPreset("today", now, loc)
	if !ok {
		t.Fatal("today preset should be valid")
	}
	// پایان روز به وقت تهران = 20:29 UTC همان روز.
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

func TestTaskOriginSuffixAndParse(t *testing.T) {
	o := taskOrigin{List: true, Filter: "pending", Page: 2}
	suffix := o.suffix()
	if suffix != ":L:pending:2" {
		t.Errorf("suffix = %q", suffix)
	}
	parts := []string{"task", "done", uuid.New().String(), "L", "pending", "2"}
	got, err := parseTaskOrigin(parts, 3)
	if err != nil {
		t.Fatalf("parseTaskOrigin failed: %v", err)
	}
	if !got.List || got.Filter != "pending" || got.Page != 2 {
		t.Errorf("parsed origin = %+v", got)
	}

	card := taskOrigin{}.suffix()
	if card != ":C" {
		t.Errorf("card suffix = %q, want :C", card)
	}
	got, err = parseTaskOrigin([]string{"task", "info", uuid.New().String(), "C"}, 3)
	if err != nil || got.List {
		t.Errorf("card origin parse = %+v, err %v", got, err)
	}
}

func TestTaskDataRoundTrip(t *testing.T) {
	id := uuid.New()
	o := taskOrigin{List: true, Filter: "completed", Page: 3}
	data := taskData("done", id, o)
	if data != "task:done:"+id.String()+":L:completed:3" {
		t.Errorf("taskData = %q", data)
	}
	if len(data) > 64 {
		t.Errorf("callback data longer than 64 bytes: %d", len(data))
	}
}
