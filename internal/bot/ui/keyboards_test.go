package ui

import (
	"testing"

	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"
)

// buttonsOf flattens all buttons of a markup.
func buttonsOf(m models.InlineKeyboardMarkup) []models.InlineKeyboardButton {
	var out []models.InlineKeyboardButton
	for _, row := range m.InlineKeyboard {
		out = append(out, row...)
	}
	return out
}

// hasButton reports whether a button with the given text exists.
func hasButton(m models.InlineKeyboardMarkup, text string) bool {
	for _, b := range buttonsOf(m) {
		if b.Text == text {
			return true
		}
	}
	return false
}

// hasStyledButton reports whether a button with text and style exists.
func hasStyledButton(m models.InlineKeyboardMarkup, text, style string) bool {
	for _, b := range buttonsOf(m) {
		if b.Text == text && b.Style == style {
			return true
		}
	}
	return false
}

func TestMainKeyboardColoredButtons(t *testing.T) {
	kb := MainKeyboard(Fa)

	if len(kb.Keyboard) == 0 {
		t.Fatal("main keyboard has no rows")
	}

	styles := map[string]string{}
	for _, row := range kb.Keyboard {
		for _, btn := range row {
			if btn.Style != "" {
				styles[btn.Text] = btn.Style
			}
		}
	}

	if styles["تسک جدید"] != StyleSuccess {
		t.Errorf("new task button style = %q, want %q", styles["تسک جدید"], StyleSuccess)
	}
	if styles["برنامه امروز"] != StylePrimary {
		t.Errorf("today button style = %q, want %q", styles["برنامه امروز"], StylePrimary)
	}
	if styles["تنظیمات"] != "" {
		t.Errorf("settings button should have no style, got %q", styles["تنظیمات"])
	}
	if !kb.IsPersistent || !kb.ResizeKeyboard {
		t.Error("main keyboard should be persistent and resized")
	}
}

func TestTaskCardKeyboardColoredButtons(t *testing.T) {
	id := uuid.New()

	pending := &domain.Task{BaseModel: domain.BaseModel{ID: id}, Title: "تست", Status: "pending"}
	kb := TaskCardKeyboard(pending, NoOrigin, Fa, false)
	if !hasStyledButton(kb, "✅ انجام شد", StyleSuccess) {
		t.Error("pending task card must have green done button")
	}
	if !hasStyledButton(kb, "🗑 حذف", StyleDanger) {
		t.Error("task card must have red delete button")
	}
	if hasButton(kb, "🔄 بازگشایی تسک") {
		t.Error("pending task card must not have reopen button")
	}

	completed := &domain.Task{BaseModel: domain.BaseModel{ID: id}, Title: "تست", Status: "completed"}
	kb = TaskCardKeyboard(completed, NoOrigin, Fa, false)
	if !hasStyledButton(kb, "🔄 بازگشایی تسک", StylePrimary) {
		t.Error("completed task card must have blue reopen button")
	}
	if hasButton(kb, "✅ انجام شد") {
		t.Error("completed task card must not have done button")
	}
}

func TestTaskListKeyboardFiltersAndPagination(t *testing.T) {
	tasks := []domain.Task{
		{BaseModel: domain.BaseModel{ID: uuid.New()}, Title: "یکی"},
		{BaseModel: domain.BaseModel{ID: uuid.New()}, Title: "دو"},
	}
	kb := TaskListKeyboard(tasks, FilterPending, 2, 3, Fa)

	if !hasButton(kb, "● ⏳ باز") {
		t.Error("active filter tab should be marked")
	}
	if !hasButton(kb, "⬅️ قبلی") || !hasButton(kb, "بعدی ➡️") {
		t.Error("pagination buttons missing on middle page")
	}

	single := TaskListKeyboard(tasks, FilterAll, 1, 1, Fa)
	if hasButton(single, "⬅️ قبلی") || hasButton(single, "بعدی ➡️") {
		t.Error("pagination buttons should be hidden on single page")
	}
}

func TestDeleteConfirmKeyboard(t *testing.T) {
	kb := DeleteConfirmKeyboard(&domain.Task{BaseModel: domain.BaseModel{ID: uuid.New()}}, NoOrigin, Fa)
	if !hasStyledButton(kb, "🗑 بله، حذف کن", StyleDanger) {
		t.Error("confirm delete button must be danger styled")
	}
	if !hasButton(kb, "انصراف") {
		t.Error("confirm dialog must have cancel button")
	}
}
