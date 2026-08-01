package main

import (
	"embed"
	"io/fs"
	"os"
)

// Everything the server reads at runtime is compiled into the binary. That makes
// the deployable artifact a single file with no working-directory assumptions -
// it does not matter where the container starts it from, and there is nothing to
// COPY alongside it.
//
// The all: prefix is load-bearing: a plain `embed templates` walks the tree but
// drops every file whose name starts with `_` or `.`. Keep all: so future
// underscored template partials still embed.
//
//go:embed all:components all:templates all:static
var embedded embed.FS

// assets is the root every load goes through. In dev it is the working
// directory, so editing a component and reloading the page shows the change
// without a rebuild; otherwise it is the embedded copy.
//
// One fs.FS either way, so nothing downstream knows which it got. Note that
// fs.FS paths are always slash-separated and relative - path.Join, never
// filepath.Join, below this line.
var assets fs.FS = embedded

func useDiskAssets() {
	assets = os.DirFS(".")
}
