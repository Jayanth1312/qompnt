package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"runtime"
	"strings"
	"testing"
)

func TestVersionOlder(t *testing.T) {
	tests := []struct {
		cur, lat string
		want     bool
	}{
		{"1.0.0", "1.0.1", true},
		{"1.0.1", "1.0.1", false},
		{"v1.0.0", "v1.1.0", true},
		{"1.2.0", "1.1.9", false},
		{"dev", "1.0.1", true},
	}
	for _, tc := range tests {
		got := versionOlder(normalizeVersion(tc.cur), normalizeVersion(tc.lat))
		if got != tc.want {
			t.Fatalf("%q vs %q: got %v want %v", tc.cur, tc.lat, got, tc.want)
		}
	}
}

func TestExtractFromTarGz(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	if err := tw.WriteHeader(&tar.Header{Name: "qomp", Mode: 0o755, Size: 3, Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("bin")); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()

	got, err := extractFromTarGz(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "bin" {
		t.Fatalf("got %q", got)
	}
}

func TestReleaseAssetName(t *testing.T) {
	name := releaseAssetName()
	if name == "" {
		t.Fatal("empty asset name")
	}
	if !strings.Contains(name, runtime.GOOS) {
		t.Fatalf("expected GOOS in %q", name)
	}
}
