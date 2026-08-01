package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type server struct {
	dev bool

	mu         sync.RWMutex
	components []Component
	tmpl       templates
	// componentsCSS is generated from the components on load and served from
	// memory - see reload.
	componentsCSS []byte
	// version stamps every asset URL and doubles as the cache key for them.
	version string
}

type templates struct {
	home   *template.Template // "layout"
	index  *template.Template // "layout"
	detail *template.Template // "layout"
	cards  *template.Template // "cards"
}

// pageData is what every template receives.
type pageData struct {
	Components []Component
	C          *Component
	// Companion is a second component shown in full - preview, Customize panel
	// and code - below the main one's code section. Button Group lives under
	// Button because it is a way of arranging buttons rather than a thing you go
	// looking for on its own; it still has its own page.
	Companion *Component
	Query   string
	BaseURL string
	// V is the asset version, stamped onto every /static/ URL in the layout.
	V string
	// Home is the alran-style landing page at /. Gallery and detail leave it false.
	Home bool
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dev := flag.Bool("dev", false, "reload components and templates on every request")
	flag.Parse()

	// Cloud Run (and most PaaS) assign the port and pass it in the environment;
	// binding anything else means the container is killed as unhealthy. The flag
	// stays the local default.
	if p := os.Getenv("PORT"); p != "" {
		*addr = ":" + p
	}

	s := &server{dev: *dev}
	// Under -dev the files on disk win, so an edit shows up on the next request
	// without a rebuild. A build serves what was compiled into it.
	if *dev {
		useDiskAssets()
	}
	if err := s.reload(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /components", s.handleIndex)
	mux.HandleFunc("GET /c/{slug}", s.handleDetail)
	mux.HandleFunc("GET /p/search", s.handleSearch)
	mux.HandleFunc("GET /src/{slug}", s.handleSource)
	mux.HandleFunc("GET /prompt/{slug}", s.handlePrompt)
	// ServeMux wildcards must be whole segments, so ".json" cannot be a literal
	// suffix in the pattern - one handler takes the filename and splits it.
	mux.HandleFunc("GET /r/{file}", s.handleRegistry)
	// More specific than the pattern below it, so ServeMux routes the generated
	// stylesheet here and the rest to the files.
	mux.Handle("GET /static/components.css", staticCache(s.assetVersion, http.HandlerFunc(s.handleComponentsCSS)))
	static, err := fs.Sub(assets, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("GET /static/", staticCache(s.assetVersion, http.StripPrefix("/static/", http.FileServer(http.FS(static)))))
	s.routeDemos(mux)

	log.Printf("qompnt listening on %s (dev=%v)", *addr, *dev)
	log.Fatal(http.ListenAndServe(*addr, s.withReload(mux)))
}

// staticCache decides how long a static file may be reused.
//
// Filenames here are stable - components.css keeps its name across a rule change
// - so a browser caching them by name would render old rules, which looks like a
// CSS bug that reproduces for nobody else. The version is carried in the query
// string instead: every asset in layout.html is stamped with ?v=<build>, and the
// build is a hash of the files themselves. A stamped URL is therefore safe to
// cache forever, because editing a file changes the URL. An unstamped one - a
// file typed in by hand, or fetched by something that dropped the query - falls
// back to revalidating, which still costs a header exchange rather than a
// download.
func staticCache(version func() string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read per request, not once at startup: in dev the version changes every
		// time a file is edited.
		if v := version(); v != "" && r.URL.Query().Get("v") == v {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			// Cache at the CDN as well as in the browser. Without this the edge
			// forwards every first-visit asset request to the function, and on a
			// pay-per-CPU platform five stylesheets and scripts per new visitor is
			// five invocations that produce a byte-identical response. The CDN
			// cache is purged on deploy, and the URL carries the build hash
			// anyway, so nothing can go stale.
			w.Header().Set("CDN-Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		h.ServeHTTP(w, r)
	})
}

// staticVersion hashes every file the layout links, so one string identifies the
// whole asset set. Per-file hashes would let an unchanged file keep its URL, but
// there are six of them: the extra plumbing buys a few kilobytes on the one load
// after a deploy.
func staticVersion(generated []byte) string {
	var names []string
	// Walked, not globbed: static/themes/*.css is stamped and cached for a year
	// like everything else, so a design system edited but not hashed here would
	// keep serving from cache until the filename changed.
	fs.WalkDir(assets, "static", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			names = append(names, p)
		}
		return nil
	})
	sort.Strings(names)

	h := sha256.New()
	for _, n := range names {
		b, err := fs.ReadFile(assets, n)
		if err != nil {
			continue
		}
		fmt.Fprintf(h, "%s:%x", n, sha256.Sum256(b))
	}
	// The generated stylesheet is not a file on the asset root in a build, and it
	// changes whenever a component's styles.css does - so it is hashed directly
	// or an edited component would keep serving from cache for a year.
	fmt.Fprintf(h, "components.css:%x", sha256.Sum256(generated))
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// componentsCSS concatenates every component's styles.css. The per-component
// file is the source of truth: this is generated from it, so a rule cannot drift
// between what the CSS tab shows and what the site loads.
func componentsCSS(cs []Component) []byte {
	var b bytes.Buffer
	b.WriteString("/* GENERATED - do not edit.\n" +
		"   Concatenated from components/<name>/styles.css by the server on load.\n" +
		"   Edit the component's file; this is only the delivery format. */\n")
	for _, c := range cs {
		if strings.TrimSpace(c.Styles) == "" {
			continue
		}
		fmt.Fprintf(&b, "\n/* ---- %s ---- */\n%s\n", c.Slug, strings.TrimSpace(c.Styles))
	}
	return b.Bytes()
}

// reload re-reads templates and components from the asset root - the embedded
// copy in a build, the working directory under -dev.
func (s *server) reload() error {
	cs, err := loadComponents(assets, "components")
	if err != nil {
		return err
	}

	// One template set per page: layout.html is shared, but each page defines
	// its own "content", so they cannot live in the same set.
	parse := func(files ...string) *template.Template {
		return template.Must(template.ParseFS(assets, files...))
	}
	t := templates{
		home:   parse("templates/layout.html", "templates/home.html"),
		index:  parse("templates/layout.html", "templates/index.html", "templates/_cards.html"),
		detail: parse("templates/layout.html", "templates/detail.html"),
		cards:  parse("templates/_cards.html"),
	}

	// components.css is generated, so it is held in memory and served from there
	// rather than being a file the request path depends on. It used to be written
	// on every reload, which raced with the file server reading it.
	//
	// Still written to disk under -dev, because static/ is committed: a consumer
	// can copy the directory and get the same stylesheet the site serves, and the
	// checked-in file stays in step with the component sources it came from. A
	// temp file and a rename, so a concurrent reader gets one whole version.
	css := componentsCSS(cs)
	if s.dev {
		tmp := "static/.components.css.tmp"
		if err := os.WriteFile(tmp, css, 0o644); err != nil {
			return err
		}
		if err := os.Rename(tmp, "static/components.css"); err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.components, s.tmpl, s.componentsCSS = cs, t, css
	s.version = staticVersion(css)
	s.mu.Unlock()
	return nil
}

func (s *server) withReload(next http.Handler) http.Handler {
	if !s.dev {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Pages only. A static file is served straight off disk and gains nothing
		// from re-reading every component - and the page that asked for it has
		// just done exactly that. Reloading here also meant a request for
		// components.css rewrote components.css while the file server was reading
		// it, which is a race no amount of atomic writing should have to cover.
		if !strings.HasPrefix(r.URL.Path, "/static/") {
			if err := s.reload(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) snapshot() ([]Component, templates) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.components, s.tmpl
}

// handleComponentsCSS serves the generated stylesheet from memory. http.ServeContent
// rather than a bare Write: it handles range requests and the conditional headers
// the file server would have given this file for free.
func (s *server) handleComponentsCSS(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	css := s.componentsCSS
	s.mu.RUnlock()
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	http.ServeContent(w, r, "components.css", time.Time{}, bytes.NewReader(css))
}

func (s *server) assetVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// render buffers first. Executing straight into the ResponseWriter means a
// template error - a renamed field, say - ships a 200 that just stops mid-page,
// which looks like a styling bug rather than the failure it is.
func (s *server) render(w http.ResponseWriter, r *http.Request, t *template.Template, name string, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Revalidate rather than no-store. The page is always checked with the
	// server, so a new build is never missed, but an unchanged one comes back as
	// a 304 with no body - and no-store additionally disables the back/forward
	// cache, which is what made every Back press re-render the whole page.
	//
	// The ETag is the rendered bytes, so it covers the markup, the component
	// list and the asset version in one.
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(buf.Bytes()))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	// The page is the same bytes for everyone - theme, design system and motion
	// are all client state - so the CDN can answer for it and the function only
	// runs on a cache miss. Safe at this TTL because Vercel purges the edge cache
	// on every deploy; the browser still revalidates, so a new build is never
	// served stale from a laptop that was open through it.
	w.Header().Set("CDN-Cache-Control", "public, s-maxage=86400, stale-while-revalidate=604800")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	buf.WriteTo(w)
}

func (s *server) handleHome(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	s.render(w, r, t.home, "layout", pageData{Components: cs, Home: true, V: s.assetVersion()})
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	s.render(w, r, t.index, "layout", pageData{Components: cs, V: s.assetVersion()})
}

// handleSearch returns just the card stack, swapped in by htmx.
func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	q := r.URL.Query().Get("q")
	s.render(w, r, t.cards, "cards", pageData{Components: search(cs, q), Query: q})
}

func (s *server) handleDetail(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	c, ok := find(cs, r.PathValue("slug"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	data := pageData{Components: cs, C: &c, BaseURL: baseURL(r), V: s.assetVersion()}
	if c.Companion != "" {
		if comp, ok := find(cs, c.Companion); ok {
			data.Companion = &comp
		}
	}
	s.render(w, r, t.detail, "layout", data)
}

// handleSource serves a component's markup as plain text - what the copy button
// on each card puts on the clipboard.
func (s *server) handleSource(w http.ResponseWriter, r *http.Request) {
	cs, _ := s.snapshot()
	c, ok := find(cs, r.PathValue("slug"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, demoHooksRe.ReplaceAllString(c.PreviewSource, ""))
}

// handlePrompt serves the agent prompt for a component as plain text.
func (s *server) handlePrompt(w http.ResponseWriter, r *http.Request) {
	cs, _ := s.snapshot()
	c, ok := find(cs, r.PathValue("slug"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, c.AgentPrompt(baseURL(r)))
}
