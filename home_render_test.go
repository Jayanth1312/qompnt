package main

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestHomeRendersPreviewStashAndPopover(t *testing.T) {
	cs, err := loadComponents(assets, "components")
	if err != nil {
		t.Fatal(err)
	}
	// parse is a local closure in reload; mirror it here for the same FS paths.
	tmpl, err := template.ParseFS(assets, "templates/layout.html", "templates/home.html")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pageData{
		Components: cs,
		Home:       true,
		V:          "test",
	}); err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	if !strings.Contains(html, `data-home-preview-pop`) {
		t.Error("missing data-home-preview-pop")
	}
	// Pick a known slug that always exists.
	if !strings.Contains(html, `data-home-preview="button"`) {
		t.Error("missing template data-home-preview=button")
	}
	if !strings.Contains(html, `data-slug="button"`) {
		t.Error("missing data-slug on row")
	}
	if !strings.Contains(html, `home-comp-row`) {
		t.Error("missing home-comp-row")
	}
}
