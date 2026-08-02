package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCLIManifest(t *testing.T) {
	s := &server{}
	if err := s.reload(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/r/cli/manifest.json", nil)
	req.Host = "example.test"
	rec := httptest.NewRecorder()
	s.handleCLIManifest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}

	var m cliManifest
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	if m.Registry != "http://example.test" {
		t.Fatalf("registry: %q", m.Registry)
	}
	if m.Assets.Tokens == "" || m.Assets.Utils == "" {
		t.Fatal("missing asset URLs")
	}
	if len(m.Minimal) == 0 {
		t.Fatal("empty minimal preset")
	}
	if len(m.Themes) < 2 {
		t.Fatalf("expected claude + theme files, got %d", len(m.Themes))
	}
	if m.Themes[0].ID != "claude" {
		t.Fatalf("first theme should be claude, got %q", m.Themes[0].ID)
	}
	if len(m.Components) == 0 {
		t.Fatal("no components")
	}
	for _, c := range m.Components {
		if c.Slug == "" || c.URL == "" || c.Hash == "" {
			t.Fatalf("incomplete component: %+v", c)
		}
	}
}
