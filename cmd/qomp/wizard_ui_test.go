package main

import "testing"

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
	if m.resultSingle != "claude" || !m.selected["claude"] {
		t.Fatalf("expected claude selected by default, got %q", m.resultSingle)
	}
}
