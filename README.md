# qompnt

HTML components that share one set of design tokens. Pick a look, drop the
markup in, and swap the whole system by changing a stylesheet.

Native elements first — browse the [live catalog](https://qompnt.vercel.app),
open a component, customize it, copy what you need. Ten design systems share the
same markup; one stylesheet changes the whole look.

## Install the CLI

Download a release from [GitHub Releases](https://github.com/Jayanth1312/qompnt/releases),
or use one of the commands below for your platform.

### Windows

**PowerShell** (downloads the release and adds `qomp` to your user PATH):

```powershell
irm https://raw.githubusercontent.com/Jayanth1312/qompnt/main/install/windows/install.ps1 | iex
```

**Scoop:**

```powershell
scoop install https://raw.githubusercontent.com/Jayanth1312/qompnt/main/install/scoop/qomp.json
```

### macOS

**Homebrew:**

```sh
brew install --formula https://raw.githubusercontent.com/Jayanth1312/qompnt/main/install/homebrew/qomp.rb
```

### Linux

**Homebrew** (Linuxbrew):

```sh
brew install --formula https://raw.githubusercontent.com/Jayanth1312/qompnt/main/install/homebrew/qomp.rb
```

**Or** download the `qomp_linux_amd64` / `qomp_linux_arm64` archive from
[GitHub Releases](https://github.com/Jayanth1312/qompnt/releases), extract it,
and put `qomp` on your PATH.

### After install

In any project:

```sh
qomp init
```

`qomp init` walks you through a design system and which components to install.
Files land under `components/qompnt/` with a `qomp.json` config.

```sh
qomp add accordion    # add one later
qomp update           # refresh from the registry
```

Default registry: `https://qompnt.vercel.app` (override with `--registry` or `QOMP_REGISTRY`).

Non-interactive init:

```sh
qomp init --theme apple --components minimal
# --components: all | minimal | none | button,card,...
```

Link stylesheets from `components/qompnt/styles/` (`tokens.css`, `qompnt.css`,
and your theme file).

**Without the CLI** — copy markup from the site and link:

```html
<link rel="stylesheet" href="https://qompnt.vercel.app/static/tokens.css">
<link rel="stylesheet" href="https://qompnt.vercel.app/static/qompnt.css">
```

**Or use shadcn** — install from the registry at `/r/<slug>.json`. Your project needs `components.json` and `tsconfig.json`. Files land in `components/qompnt/`, same layout as `qomp init`:

```sh
npx shadcn@latest add https://qompnt.vercel.app/r/button.json
```

Link `components/qompnt/styles/tokens.css` and `qompnt.css` in your HTML after install.

## Want to clone the project?

You need [Go](https://go.dev/dl/) installed.

```sh
git clone https://github.com/Jayanth1312/qompnt.git
cd qompnt
```

With Make:

```sh
make dev      # generate CSS, serve on :8080 with live reload
make          # generate CSS and build ./bin/qompnt
make qomp     # build the ./bin/qomp installer CLI
make test     # go test ./...
```

Without Make (same steps, plain Go):

```sh
go generate ./...
go run . -dev          # dev server on :8080
go build -o bin/qompnt .   # site binary
go build -o bin/qomp ./cmd/qomp
go test ./...
```

`make dev` / `go run . -dev` re-reads `components/` and `templates/` on every
request. The production binary reads them once at startup.

MIT.
