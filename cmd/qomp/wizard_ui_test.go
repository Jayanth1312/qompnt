package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchPickerRefilter(t *testing.T) {
	m := newSearchPicker("Search components", "Select components", []pickerItem{
		{label: "Accordion", value: "accordion"},
		{label: "Avatar", value: "avatar"},
		{label: "Button", value: "button"},
	}, true)
	m.filter = "av"
	m.refilter()
	if len(m.filtered) != 1 || m.filtered[0].value != "avatar" {
		t.Fatalf("filtered: %+v", m.filtered)
	}
}

func TestSearchPickerSingleDefault(t *testing.T) {
	m := newSearchPicker("Search themes", "Select themes", []pickerItem{
		{label: "Claude", value: "claude"},
		{label: "Apple", value: "apple"},
	}, false)
	if m.resultSingle != "claude" {
		t.Fatalf("expected claude selected by default, got %q", m.resultSingle)
	}
	if m.selected["claude"] {
		t.Fatal("single mode should not use checkbox selected map")
	}
}

func TestSearchPickerMultiEnterToggles(t *testing.T) {
	m := newSearchPicker("Search components", "Select components", []pickerItem{
		{label: "Button", value: "button"},
		{label: "Card", value: "card"},
	}, true)
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.selected["button"] {
		t.Fatal("enter should toggle mark, not quit")
	}
	if m.done {
		t.Fatal("enter must not finish multi select")
	}
	m.cursor = 1
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.selected["card"] {
		t.Fatal("space should toggle mark")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if !m.done {
		t.Fatal("ctrl+enter/j should finish multi select")
	}
}
