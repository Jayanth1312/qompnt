package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Short public URLs for install manifests — served at /install/{file}.
var installFiles = map[string]struct {
	fsPath      string
	contentType string
}{
	"windows.ps1": {fsPath: "install/windows/install.ps1", contentType: "text/plain; charset=utf-8"},
	"scoop.json":  {fsPath: "install/scoop/qomp.json", contentType: "application/json; charset=utf-8"},
	"qomp.rb":     {fsPath: "install/homebrew/qomp.rb", contentType: "text/plain; charset=utf-8"},
}

func (s *server) handleInstall(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.PathValue("file"), "/")
	meta, ok := installFiles[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	data, err := fs.ReadFile(assets, path.Clean(meta.fsPath))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", meta.contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Write(data)
}
