# qompnt

Thirty HTML components on a shared token contract. Ten design systems share the
same markup — swap one stylesheet and the whole look changes. Copy the markup,
install with `shadcn`, or pull from a CDN.

Go serves the site, htmx handles the interactive previews, UnoCSS generates the
stylesheet at build time. No JavaScript framework, nothing to build in a
consumer's project.

## Run it

```
make dev      # generate CSS, then serve on :8080 with live reload
make          # generate CSS and build the ./qompnt binary
make test     # go test ./...
```

`make dev` re-reads `components/` and `templates/` on every request. The binary
without `-dev` reads them once at startup.

## Add a component

Create a directory under `components/`. There is no registration step — the
server picks it up, the home page lists it, and the registry serves it.

```
components/tooltip/
  meta.json      name, blurb, created (YYYY-MM-DD), tags, motion
  preview.html   what renders in the card and the detail hero
  notes.md       prose, blank-line separated paragraphs (gitignored)
  src/
    tooltip.html the copyable markup - one file per code tab, max 6
```

Then `make css`, because the stylesheet only contains classes that appear in the
markup. Forgetting this is the one way to get an unstyled component.

## Consuming a component

**Copy-paste.** Take the markup from the Code section and link the two
stylesheets:

```html
<link rel="stylesheet" href="https://<host>/static/tokens.css">
<link rel="stylesheet" href="https://<host>/static/qompnt.css">
```

**shadcn.** Installs the markup and both stylesheets, and writes the colour
tokens into your own CSS:

```
npx shadcn@latest add https://<host>/r/button.json
```

Requires a `components.json` and a `tsconfig.json` in the target project — the
shadcn CLI refuses to run without both, even for plain files.

On Tailwind 4 you don't need `qompnt.css` at all; your build compiles the classes
in the markup, and `tokens.css` carries the typography and radius scale in a
`@theme` block. One caveat: shadcn writes dark-mode tokens under `.dark`, while
this site switches on `[data-theme="dark"]`. Pick whichever your app already
uses — the token names are identical either way.

## Layout

```
main.go           routes, startup scan
component.go      Component, loaded from disk
registry.go       shadcn registry JSON, plus the token maps
markdown.go       the "Copy markdown" blob
demo.go           endpoints backing the interactive previews
uno.config.ts     design tokens as an UnoCSS theme
templates/        layout.html (shell), home.html (index), detail.html
static/           tokens.css, themes/*.css (design systems), qompnt.css
                  (generated), components.css (generated), htmx.js, qompnt.js
```

## Notes on the stack

**Design systems.** Claude is the default (`tokens.css` alone). Nine more live in
`static/themes/` — Apple, ClickHouse, Cohere, Coinbase, Cursor, Linear, Notion,
Stripe, Vercel — each a single stylesheet of token overrides. The palette
switcher and the pen button in the header cycle through them; choice persists in
`localStorage` as `qompnt-ds`.

**Home previews.** The component list on `/` stashes each `preview.html` in a
`<template>` and shows a cursor-following popover after a short hover delay.
`static/qompnt.js` owns the timing and positioning; fine-pointer devices only.

**Themes.** UnoCSS colours point at CSS variables and `tokens.css` swaps the
variables, so one set of classes covers light and dark. Nothing is duplicated per
theme.

**Motion.** Every animation is CSS, times off four tokens in `tokens.css`
(`--motion-fast`, `--motion`, `--motion-slow`, `--ease`), and turns off with one
attribute:

```html
<html data-motion="off">
```

That zeroes the tokens and clamps anything carrying its own timing, so it also
catches utility classes and third-party CSS. `prefers-reduced-motion` is the
default when nothing is stored; an explicit `data-motion="on"` overrides it,
because a visitor turning motion on for this site outranks their OS setting.
The **Animation** switch in the Customize panel writes `qompnt-motion` to
localStorage; `layout.html` applies it before first paint. It appears only where
`meta.json` sets `"motion": true` — sixteen of the thirty. A switch that
visibly does nothing reads as a broken one, so Badge, Table, Pagination and the
rest do not get it, and neither do the loaders: a Spinner's animation is the
component, not decoration on it.

If a component looks completely static, check that switch first — a browser
reporting `prefers-reduced-motion: reduce` starts with motion off, which is the
correct default and not a missing stylesheet.

One exemption, in `components/toast/styles.css`: a toast's 4.5s animation is its
lifetime — `animationend` removes the node — so with motion off the duration is
held and the timing function becomes `steps(1, end)`. Nothing moves, the toast
still lasts 4.5 seconds. Add the same exemption for any animation whose end
event does real work.

The blanket `transition: none` that used to sit in `tokens.css` is now scoped to
`[data-swapping]`, which `qompnt.js` sets for two frames while it rewrites the
tokens. That was there to stop a theme swap cross-fading the whole page; it was
also switching off every component's own motion.

**htmx 4.** Pinned to `4.0.0-beta6`, vendored in `static/`. It is a rewrite onto
`fetch()`: events are namespaced (`htmx:after:swap`, not `htmx:afterSwap`),
extensions are gone, and attribute inheritance resolves differently than in
htmx 2. Check the docs before assuming htmx 2 behaviour.

**Wind4 theme keys.** `theme.text` entries must be objects with camelCase keys
(`{ fontSize, lineHeight, letterSpacing }`). The Wind3 tuple form and
kebab-case keys both fail silently — you get a utility class referencing a
variable that was never defined.

MIT.
