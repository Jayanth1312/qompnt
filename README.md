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

```powershell
irm https://qompnt.vercel.app/install/windows.ps1 | iex
```

**Scoop:**

```powershell
scoop install https://qompnt.vercel.app/install/scoop.json
```

### macOS

```sh
brew install --formula https://qompnt.vercel.app/install/qomp.rb
```

### Linux

**Homebrew** (Linuxbrew):

```sh
brew install --formula https://qompnt.vercel.app/install/qomp.rb
```

**Or** download the `qomp_linux_amd64` / `qomp_linux_arm64` archive from
[GitHub Releases](https://github.com/Jayanth1312/qompnt/releases), extract it,
and put `qomp` on your PATH.

Install scripts are served from `qompnt.vercel.app/install/` — same files as the
repo, shorter URLs. **Security:** `irm | iex` and `curl | sh` always mean you
trust the host; inspect the script first if you want to verify. Scoop and
Homebrew additionally check SHA-256 hashes before installing. The Windows script
only downloads the official GitHub release binary — it does not run arbitrary
code from elsewhere.

### After install

In any project:

```sh
qomp init
```

`qomp init` walks you through a design system and which components to install.
Files land under `components/qompnt/` with a `qomp.json` config.

```sh
qomp add accordion    # add one later
qomp update           # refresh project files from the registry
qomp upgrade          # update the qomp binary itself
```

`qomp update` refreshes themes and components in your project. `qomp upgrade` downloads the latest release from GitHub and replaces the running binary (`qomp upgrade --check` only prints if one is available).

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
