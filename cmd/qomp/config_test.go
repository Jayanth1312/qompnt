package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveFindConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := defaultConfig("http://example.test")
	cfg.Theme = "apple"
	cfg.Accent = "#0071e3"
	cfg.Components = []string{"button"}
	if err := saveConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}

	sub := filepath.Join(dir, "nested", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, root, err := findConfig(sub)
	if err != nil {
		t.Fatal(err)
	}
	if root != dir {
		t.Fatalf("root %q want %q", root, dir)
	}
	if got.Theme != "apple" || got.Accent != "#0071e3" || len(got.Components) != 1 {
		t.Fatalf("unexpected config: %+v", got)
	}
}

func TestAccentCSSContrast(t *testing.T) {
	css := accentCSS("#0071e3")
	if !strings.Contains(css, "--primary: #0071e3") || !strings.Contains(css, "--primary-foreground: #ffffff") {
		t.Fatalf("unexpected css:\n%s", css)
	}
	light := accentCSS("#f5f5f7")
	if !strings.Contains(light, "--primary-foreground: #000000") {
		t.Fatalf("expected black foreground for light accent:\n%s", light)
	}
}

func TestResolveSlugs(t *testing.T) {
	m := &Manifest{
		Minimal: []string{"button", "input"},
		Components: []Component{
			{Slug: "button"}, {Slug: "input"}, {Slug: "card"},
		},
	}
	if got := resolveSlugs(m, "none", nil); len(got) != 0 {
		t.Fatalf("none: %v", got)
	}
	if got := resolveSlugs(m, "minimal", nil); len(got) != 2 {
		t.Fatalf("minimal: %v", got)
	}
	if got := resolveSlugs(m, "all", nil); len(got) != 3 {
		t.Fatalf("all: %v", got)
	}
	if got := resolveSlugs(m, "selective", []string{"card"}); len(got) != 1 || got[0] != "card" {
		t.Fatalf("selective: %v", got)
	}
}
