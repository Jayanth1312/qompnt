package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type server struct {
	dev bool

	mu         sync.RWMutex
	components []Component
	tmpl       templates
}

type templates struct {
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
	Query     string
	BaseURL   string
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dev := flag.Bool("dev", false, "reload components and templates on every request")
	flag.Parse()

	s := &server{dev: *dev}
	if err := s.reload(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /c/{slug}", s.handleDetail)
	mux.HandleFunc("GET /p/search", s.handleSearch)
	mux.HandleFunc("GET /src/{slug}", s.handleSource)
	mux.HandleFunc("GET /prompt/{slug}", s.handlePrompt)
	// ServeMux wildcards must be whole segments, so ".json" cannot be a literal
	// suffix in the pattern - one handler takes the filename and splits it.
	mux.HandleFunc("GET /r/{file}", s.handleRegistry)
	mux.Handle("GET /static/", staticCache(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))))
	s.routeDemos(mux)

	log.Printf("qompnt listening on %s (dev=%v)", *addr, *dev)
	log.Fatal(http.ListenAndServe(*addr, s.withReload(mux)))
}

// staticCache makes the browser revalidate every static file.
//
// Without a Cache-Control header a browser is free to invent a freshness
// lifetime from Last-Modified and serve the file for hours without asking. These
// filenames are not content-hashed, so components.css keeps its name across a
// rule change and a client that skipped revalidation renders the old rules -
// which looks like a CSS bug that reproduces for nobody else. no-cache still
// allows the cache: the file is stored and a 304 costs a header exchange, not a
// download. Immutable caching is available once filenames carry a hash, not
// before.
func staticCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		h.ServeHTTP(w, r)
	})
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

// reload re-reads templates and components from disk.
func (s *server) reload() error {
	cs, err := loadComponents("components")
	if err != nil {
		return err
	}

	// One template set per page: layout.html is shared, but each page defines
	// its own "content", so they cannot live in the same set.
	parse := func(files ...string) *template.Template {
		return template.Must(template.ParseFiles(files...))
	}
	t := templates{
		index:  parse("templates/layout.html", "templates/index.html", "templates/_cards.html"),
		detail: parse("templates/layout.html", "templates/detail.html"),
		cards:  parse("templates/_cards.html"),
	}

	// Written to disk rather than served from memory so a plain-HTML consumer can
	// copy static/ and get the same stylesheet the site uses.
	if err := os.WriteFile("static/components.css", componentsCSS(cs), 0o644); err != nil {
		return err
	}

	s.mu.Lock()
	s.components, s.tmpl = cs, t
	s.mu.Unlock()
	return nil
}

func (s *server) withReload(next http.Handler) http.Handler {
	if !s.dev {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.reload(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) snapshot() ([]Component, templates) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.components, s.tmpl
}

// render buffers first. Executing straight into the ResponseWriter means a
// template error - a renamed field, say - ships a 200 that just stops mid-page,
// which looks like a styling bug rather than the failure it is.
func (s *server) render(w http.ResponseWriter, t *template.Template, name string, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Dev site: never let a browser (or htmx, which uses normal cache rules on
	// boosted navigations) reuse HTML from an earlier build.
	w.Header().Set("Cache-Control", "no-store")
	buf.WriteTo(w)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	s.render(w, t.index, "layout", pageData{Components: cs})
}

// handleSearch returns just the card stack, swapped in by htmx.
func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	q := r.URL.Query().Get("q")
	s.render(w, t.cards, "cards", pageData{Components: search(cs, q), Query: q})
}

func (s *server) handleDetail(w http.ResponseWriter, r *http.Request) {
	cs, t := s.snapshot()
	c, ok := find(cs, r.PathValue("slug"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	data := pageData{Components: cs, C: &c, BaseURL: baseURL(r)}
	if c.Companion != "" {
		if comp, ok := find(cs, c.Companion); ok {
			data.Companion = &comp
		}
	}
	s.render(w, t.detail, "layout", data)
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
