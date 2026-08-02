package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInstallScripts(t *testing.T) {
	s := &server{}
	tests := []struct {
		path    string
		contain string
	}{
		{"/install/windows.ps1", "$ver = "},
		{"/install/scoop.json", `"bin": "qomp.exe"`},
		{"/install/qomp.rb", "class Qomp"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, "http://example.com"+tc.path, nil)
		req.SetPathValue("file", strings.TrimPrefix(tc.path, "/install/"))
		rec := httptest.NewRecorder()
		s.handleInstall(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", tc.path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), tc.contain) {
			t.Fatalf("%s: body missing %q", tc.path, tc.contain)
		}
	}
}

func TestInstallScriptsNotFound(t *testing.T) {
	s := &server{}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/install/nope.txt", nil)
	req.SetPathValue("file", "nope.txt")
	rec := httptest.NewRecorder()
	s.handleInstall(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}
