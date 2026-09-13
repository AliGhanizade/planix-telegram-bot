package ui

import (
	"testing"

	"github.com/google/uuid"
)

func TestTaskOriginSuffixAndParse(t *testing.T) {
	o := TaskOrigin{List: true, Filter: "pending", Page: 2}
	if suffix := o.Suffix(); suffix != ":L:pending:2" {
		t.Errorf("Suffix = %q", suffix)
	}
	parts := []string{"task", "done", uuid.New().String(), "L", "pending", "2"}
	got, err := ParseTaskOrigin(parts, 3)
	if err != nil {
		t.Fatalf("ParseTaskOrigin failed: %v", err)
	}
	if !got.List || got.Filter != "pending" || got.Page != 2 {
		t.Errorf("parsed origin = %+v", got)
	}

	card := TaskOrigin{}.Suffix()
	if card != ":C" {
		t.Errorf("card suffix = %q, want :C", card)
	}
	got, err = ParseTaskOrigin([]string{"task", "info", uuid.New().String(), "C"}, 3)
	if err != nil || got.List {
		t.Errorf("card origin parse = %+v, err %v", got, err)
	}
}

func TestTaskDataRoundTrip(t *testing.T) {
	id := uuid.New()
	o := TaskOrigin{List: true, Filter: "completed", Page: 3}
	data := TaskData("done", id, o)
	if data != "task:done:"+id.String()+":L:completed:3" {
		t.Errorf("TaskData = %q", data)
	}
	if len(data) > 64 {
		t.Errorf("callback data longer than 64 bytes: %d", len(data))
	}
}

func TestParseListFilter(t *testing.T) {
	if ParseListFilter("completed") != FilterCompleted {
		t.Error("completed filter parse failed")
	}
	if ParseListFilter("anything") != FilterPending {
		t.Error("unknown filter should default to pending")
	}
}
