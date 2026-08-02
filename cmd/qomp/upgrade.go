package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const githubRepo = "Jayanth1312/qompnt"

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpgrade(checkOnly bool) error {
	if version == "dev" {
		fmt.Fprintln(os.Stderr, "dev build — install a release binary or rebuild from source")
		return nil
	}

	latest, assetURL, err := fetchLatestRelease()
	if err != nil {
		return err
	}
	if !versionOlder(normalizeVersion(version), normalizeVersion(latest)) {
		fmt.Fprintf(os.Stderr, "qomp %s is up to date\n", version)
		return nil
	}

	fmt.Fprintf(os.Stderr, "qomp %s → %s available\n", version, latest)
	if checkOnly {
		return nil
	}

	exe, err := currentExecutable()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Downloading %s …\n", releaseAssetName())
	data, err := download(assetURL)
	if err != nil {
		return err
	}

	bin, err := extractBinary(data)
	if err != nil {
		return err
	}

	if err := replaceExecutable(exe, bin); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Upgraded to qomp %s\n", latest)
	return nil
}

func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(exe)
}

func fetchLatestRelease() (tag, assetURL string, err error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/"+githubRepo+"/releases/latest", nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "qomp")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("check release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", "", fmt.Errorf("check release: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", fmt.Errorf("parse release: %w", err)
	}
	if rel.TagName == "" {
		return "", "", fmt.Errorf("release has no tag")
	}

	want := releaseAssetName()
	for _, a := range rel.Assets {
		if a.Name == want {
			return rel.TagName, a.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("no asset %q in release %s", want, rel.TagName)
}

func releaseAssetName() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("qomp_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("qomp_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
}

func download(url string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("download: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return io.ReadAll(resp.Body)
}

func extractBinary(archive []byte) ([]byte, error) {
	if runtime.GOOS == "windows" {
		return extractFromZip(archive)
	}
	return extractFromTarGz(archive)
}

func extractFromTarGz(data []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		name := filepath.Base(hdr.Name)
		if name == "qomp" || name == "qomp.exe" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("binary not found in archive")
}

func extractFromZip(data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if name != "qomp.exe" && name != "qomp" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		return body, nil
	}
	return nil, fmt.Errorf("binary not found in archive")
}

func replaceExecutable(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, "qomp-upgrade-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return replaceExecutableWindows(dest, tmpPath)
	}
	return os.Rename(tmpPath, dest)
}

func replaceExecutableWindows(dest, tmpPath string) error {
	old := dest + ".old"
	_ = os.Remove(old)
	if err := os.Rename(dest, old); err != nil {
		return fmt.Errorf("replace binary: %w (try re-running the install script)", err)
	}
	if err := os.Rename(tmpPath, dest); err != nil {
		_ = os.Rename(old, dest)
		return err
	}
	_ = os.Remove(old)
	return nil
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// versionOlder reports whether current is strictly older than latest.
func versionOlder(current, latest string) bool {
	if current == "" || current == "dev" {
		return true
	}
	c := parseVersion(current)
	l := parseVersion(latest)
	for i := 0; i < 3; i++ {
		if c[i] < l[i] {
			return true
		}
		if c[i] > l[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(normalizeVersion(v), "-", 2)[0]
	for i, p := range strings.Split(parts, ".") {
		if i >= 3 {
			break
		}
		n, _ := strconv.Atoi(p)
		out[i] = n
	}
	return out
}
