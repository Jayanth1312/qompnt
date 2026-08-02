package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func cacheDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("LocalAppData")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, "qomp"), nil
	default:
		if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
			return filepath.Join(xdg, "qomp"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".cache", "qomp"), nil
	}
}

func cachePath(url string) (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	host := "unknown"
	rest := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		host = sanitizePath(rest[:i])
	} else if rest != "" {
		host = sanitizePath(rest)
	}
	sub := filepath.Join(dir, host)
	if err := os.MkdirAll(sub, 0o755); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(url))
	return filepath.Join(sub, hex.EncodeToString(sum[:16])), nil
}

// fetchCached returns body for url. When contentHash is non-empty and a cached
// file's sha256 matches, the network is skipped (used for theme CSS).
func fetchCached(client *http.Client, url, contentHash string) ([]byte, error) {
	path, err := cachePath(url)
	if err != nil {
		return nil, err
	}
	if contentHash != "" {
		if data, err := os.ReadFile(path); err == nil {
			sum := sha256.Sum256(data)
			if hex.EncodeToString(sum[:]) == contentHash {
				return data, nil
			}
		}
	}

	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("fetch %s: %s: %s", url, resp.Status, strings.TrimSpace(string(body)))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if contentHash != "" {
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != contentHash {
			return nil, fmt.Errorf("hash mismatch for %s", url)
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, path); err != nil {
		return nil, err
	}
	return data, nil
}

func sanitizePath(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-':
			return r
		default:
			return '_'
		}
	}, s)
	if s == "" {
		return "unknown"
	}
	return s
}
