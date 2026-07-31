package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"
)

// shadcn registry item. Components ship as registry:file - raw HTML partials
// dropped at a target path - so the same markup serves this site, a Go
// consumer, and a Tailwind consumer whose build compiles the classes.
type registryItem struct {
	Schema               string            `json:"$schema"`
	Name                 string            `json:"name"`
	Type                 string            `json:"type"`
	Title                string            `json:"title"`
	Description          string            `json:"description,omitempty"`
	RegistryDependencies []string          `json:"registryDependencies,omitempty"`
	Files                []registryFile    `json:"files"`
	CSSVars              map[string]cssMap `json:"cssVars,omitempty"`
}

type registryFile struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Target  string `json:"target"`
	Content string `json:"content"`
}

type cssMap map[string]string

type registryIndex struct {
	Schema   string             `json:"$schema"`
	Name     string             `json:"name"`
	Homepage string             `json:"homepage"`
	Items    []registryListItem `json:"items"`
}

type registryListItem struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// handleRegistry serves /r/registry.json and /r/<slug>.json.
func (s *server) handleRegistry(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	name, ok := strings.CutSuffix(file, ".json")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if name == "registry" {
		s.registryIndex(w, r)
		return
	}
	s.registryItem(w, r, name)
}

func (s *server) registryIndex(w http.ResponseWriter, r *http.Request) {
	cs, _ := s.snapshot()

	idx := registryIndex{
		Schema:   "https://ui.shadcn.com/schema/registry.json",
		Name:     "qompnt",
		Homepage: baseURL(r),
		Items: []registryListItem{{
			Name:        themeItemName,
			Type:        "registry:file",
			Title:       "qompnt theme",
			Description: "Claude design tokens and the generated utility stylesheet.",
		}},
	}
	for _, c := range cs {
		idx.Items = append(idx.Items, registryListItem{
			Name: c.Slug, Type: "registry:file", Title: c.Name, Description: c.Blurb,
		})
	}
	writeJSON(w, idx)
}

const themeItemName = "qompnt-theme"

func (s *server) registryItem(w http.ResponseWriter, r *http.Request, slug string) {
	if slug == themeItemName {
		item, err := themeItem()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, item)
		return
	}

	cs, _ := s.snapshot()
	c, ok := find(cs, slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, componentItem(c, baseURL(r)))
}

func componentItem(c Component, base string) registryItem {
	item := registryItem{
		Schema:               "https://ui.shadcn.com/schema/registry-item.json",
		Name:                 c.Slug,
		Type:                 "registry:file",
		Title:                c.Name,
		Description:          c.Blurb,
		RegistryDependencies: []string{base + "/r/" + themeItemName + ".json"},
		CSSVars:              map[string]cssMap{"light": lightVars, "dark": darkVars},
	}
	for _, f := range c.Files {
		item.Files = append(item.Files, registryFile{
			Path:    path.Join("components", c.Slug, f.Name),
			Type:    "registry:file",
			Target:  path.Join("qompnt", f.Name),
			Content: f.Body,
		})
	}
	// A component with rules of its own ships them; markup alone would install
	// something that only looks right on this site.
	if c.Styles != "" {
		item.Files = append(item.Files, registryFile{
			Path:    path.Join("components", c.Slug, "styles.css"),
			Type:    "registry:file",
			Target:  path.Join("qompnt", c.Slug+".css"),
			Content: c.Styles,
		})
	}
	return item
}

// themeItem ships the stylesheets themselves, for consumers who are not on
// Tailwind and so cannot compile the utility classes in the markup.
func themeItem() (registryItem, error) {
	item := registryItem{
		Schema:      "https://ui.shadcn.com/schema/registry-item.json",
		Name:        themeItemName,
		Type:        "registry:file",
		Title:       "qompnt theme",
		Description: "Claude design tokens and the generated utility stylesheet.",
		CSSVars:     map[string]cssMap{"light": lightVars, "dark": darkVars},
	}
	for _, name := range []string{"tokens.css", "qompnt.css"} {
		body, err := os.ReadFile(path.Join("static", name))
		if err != nil {
			return item, err
		}
		item.Files = append(item.Files, registryFile{
			Path:    path.Join("static", name),
			Type:    "registry:file",
			Target:  path.Join("qompnt", name),
			Content: string(body),
		})
	}
	return item, nil
}

// Kept in sync with static/tokens.css by TestTokensMatchRegistry. Names follow
// shadcn/ui's convention so a theme written for shadcn themes qompnt unchanged.
var lightVars = cssMap{
	"background":             "#faf9f5",
	"foreground":             "#141413",
	"card":                   "#efe9de",
	"card-foreground":        "#141413",
	"popover":                "#faf9f5",
	"popover-foreground":     "#141413",
	"primary":                "#cc785c",
	"primary-foreground":     "#ffffff",
	"secondary":              "#e8e0d2",
	"secondary-foreground":   "#141413",
	"muted":                  "#f5f0e8",
	"muted-foreground":       "#6c6a64",
	"accent":                 "#efe9de",
	"accent-foreground":      "#141413",
	"destructive":            "#c64545",
	"destructive-foreground": "#ffffff",
	"border":                 "#e6dfd8",
	"input":                  "#e6dfd8",
	"ring":                   "#3b6ef6",
	"code":                   "#1f1e1b",
	"code-foreground":        "#faf9f5",
	"success-bg":             "#eaf4ea",
	"success-fg":             "#2f6b3a",
	"success-bd":             "#c3e0c7",
	"error-bg":               "#fbeceb",
	"error-fg":               "#a33028",
	"error-bd":               "#f0cbc7",
	"info-bg":                "#e9eef9",
	"info-fg":                "#2f5199",
	"info-bd":                "#c8d5ef",
	"radius":                 "8px",
	"control-px":             "8px",
	"control-py":             "8px",
}

var darkVars = cssMap{
	"background":             "#181715",
	"foreground":             "#faf9f5",
	"card":                   "#252320",
	"card-foreground":        "#faf9f5",
	"popover":                "#1f1e1b",
	"popover-foreground":     "#faf9f5",
	"primary":                "#cc785c",
	"primary-foreground":     "#ffffff",
	"secondary":              "#2e2b27",
	"secondary-foreground":   "#faf9f5",
	"muted":                  "#1f1e1b",
	"muted-foreground":       "#a09d96",
	"accent":                 "#252320",
	"accent-foreground":      "#faf9f5",
	"destructive":            "#c64545",
	"destructive-foreground": "#ffffff",
	"border":                 "#302d29",
	"input":                  "#302d29",
	"ring":                   "#4b7dff",
	"code":                   "#0f0e0d",
	"code-foreground":        "#faf9f5",
	"success-bg":             "#17231c",
	"success-fg":             "#8ed6a0",
	"success-bd":             "#2c4633",
	"error-bg":               "#2a1917",
	"error-fg":               "#f09e96",
	"error-bd":               "#4d2b27",
	"info-bg":                "#171d2b",
	"info-fg":                "#9fbaf0",
	"info-bd":                "#2b3a5c",
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
