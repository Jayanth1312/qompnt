package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Component is one directory under components/.
type Component struct {
	Slug      string
	Name      string
	Blurb     string
	Created   time.Time
	Tags      []string
	Mechanism string   // the CSS or platform feature doing the work
	Caveats   []string // what you give up - accessibility first
	Support   []string // browser support notes for the features used

	// Motion is meta.json's "motion": the component has micro-interactions you
	// can actually see switch off. Only those pages show the Animation toggle -
	// on a Badge or a Table it would be a control that visibly does nothing, and
	// a dead switch reads as a broken one. Loaders (Spinner, Skeleton, the
	// indeterminate Progress sweep) are excluded on purpose: their animation is
	// the component, not decoration on it.
	Motion bool

	// Companion is another component's slug, rendered in full below this one's
	// code section. For a variant that only makes sense next to its parent.
	Companion string

	Preview       template.HTML // preview.html, rendered as-is
	PreviewSource string        // the same markup as text, for the code block
	Styles        string        // styles.css, optional - the component's own rules
	Options       template.HTML // options.html, optional - variant toggles
	Extras        template.HTML // extras.html, optional - shown below the code
	Files         []SourceFile  // src/*, in filename order - what the registry ships
}

// SourceFile is one tab in the Code section.
type SourceFile struct {
	Name string
	Body string
}

type meta struct {
	Name      string   `json:"name"`
	Blurb     string   `json:"blurb"`
	Created   string   `json:"created"`
	Tags      []string `json:"tags"`
	Mechanism string   `json:"mechanism"`
	Caveats   []string `json:"caveats"`
	Support   []string `json:"support"`
	Motion    bool     `json:"motion"`
	Companion string   `json:"companion"`
}

// CreatedLabel formats as "Jul 28, 2026" to match the reference layout.
func (c Component) CreatedLabel() string { return c.Created.Format("Jan 2, 2006") }

// TagLabel joins tags for the meta table row.
func (c Component) TagLabel() string { return strings.Join(c.Tags, ", ") }

// loadComponents walks dir and returns components in name order.
func loadComponents(fsys fs.FS, dir string) ([]Component, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	var out []Component
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		c, err := loadComponent(fsys, path.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		out = append(out, c)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func loadComponent(fsys fs.FS, dir string) (Component, error) {
	var c Component
	c.Slug = path.Base(dir)

	raw, err := fs.ReadFile(fsys, path.Join(dir, "meta.json"))
	if err != nil {
		return c, err
	}
	var m meta
	if err := json.Unmarshal(raw, &m); err != nil {
		return c, fmt.Errorf("meta.json: %w", err)
	}
	if m.Name == "" {
		return c, fmt.Errorf("meta.json: name is required")
	}
	created, err := time.Parse("2006-01-02", m.Created)
	if err != nil {
		return c, fmt.Errorf("meta.json: created must be YYYY-MM-DD: %w", err)
	}
	c.Name, c.Blurb, c.Created, c.Tags = m.Name, m.Blurb, created, m.Tags
	c.Mechanism, c.Caveats, c.Support, c.Motion = m.Mechanism, m.Caveats, m.Support, m.Motion
	c.Companion = m.Companion

	preview, err := fs.ReadFile(fsys, path.Join(dir, "preview.html"))
	if err != nil {
		return c, err
	}
	c.Preview = template.HTML(preview)
	c.PreviewSource = strings.TrimSpace(string(preview))

	// options.html is optional: the toggles that reshape the preview in place.
	if opts, err := fs.ReadFile(fsys, path.Join(dir, "options.html")); err == nil {
		c.Options = template.HTML(opts)
	}

	// styles.css is optional: the rules a component needs beyond the tokens.
	if css, err := fs.ReadFile(fsys, path.Join(dir, "styles.css")); err == nil {
		c.Styles = string(css)
	}

	// extras.html is optional: markup shown below the code block, for the cases
	// one preview cannot carry - a button group, say.
	if extras, err := fs.ReadFile(fsys, path.Join(dir, "extras.html")); err == nil {
		c.Extras = template.HTML(extras)
	}

	srcDir := path.Join(dir, "src")
	srcs, err := fs.ReadDir(fsys, srcDir)
	if err != nil {
		return c, fmt.Errorf("src/: %w", err)
	}
	for _, s := range srcs {
		if s.IsDir() {
			continue
		}
		body, err := fs.ReadFile(fsys, path.Join(srcDir, s.Name()))
		if err != nil {
			return c, err
		}
		c.Files = append(c.Files, SourceFile{Name: s.Name(), Body: string(body)})
	}
	if len(c.Files) == 0 {
		return c, fmt.Errorf("src/: no files")
	}
	return c, nil
}

// inlineCode renders `backticked` spans as mono and escapes everything else.
// The display face has blank glyphs for < and >, so bare angle brackets in prose
// come out as gaps - code voice is the honest fix rather than a font hack.
func inlineCode(text string) template.HTML {
	var b strings.Builder
	for i, part := range strings.Split(text, "`") {
		if i%2 == 1 {
			b.WriteString("<code>")
			b.WriteString(template.HTMLEscapeString(part))
			b.WriteString(`</code>`)
			continue
		}
		b.WriteString(template.HTMLEscapeString(part))
	}
	return template.HTML(b.String())
}

// MechanismHTML and CaveatsHTML are what the detail page renders.
func (c Component) MechanismHTML() template.HTML { return inlineCode(c.Mechanism) }

func (c Component) CaveatsHTML() []template.HTML {
	out := make([]template.HTML, len(c.Caveats))
	for i, s := range c.Caveats {
		out[i] = inlineCode(s)
	}
	return out
}

func (c Component) BlurbHTML() template.HTML { return inlineCode(c.Blurb) }

func (c Component) SupportHTML() []template.HTML {
	out := make([]template.HTML, len(c.Support))
	for i, s := range c.Support {
		out[i] = inlineCode(s)
	}
	return out
}

// CodeTailwind is the markup as shipped: utility classes bound to the token
// contract.
func (c Component) CodeTailwind() string {
	return strings.TrimSpace(demoHooksRe.ReplaceAllString(c.PreviewSource, ""))
}

// utilityClass matches the utility soup, leaving semantic hooks (data-*, role,
// aria-*) and the handful of component classes that carry real styling.
var (
	classAttr = regexp.MustCompile(`\s+class="([^"]*)"`)
	keepClass = regexp.MustCompile(`^(tbl|steps|tip|skeleton|spin|progress|slider|crumbs|sidebar|toast-life|tone-[a-z]+|nav-menu|navbar|options-grid|preview-inert|panes|dot|num|tick|load|step-label|sr-only|indicator)(-[a-z0-9]+)*$`)
)

// CodeHTML is the same structure without the utility classes - the version you
// want if you are bringing your own stylesheet.
func (c Component) CodeHTML() string {
	src := demoHooksRe.ReplaceAllString(c.PreviewSource, "")
	out := classAttr.ReplaceAllStringFunc(src, func(m string) string {
		sub := classAttr.FindStringSubmatch(m)
		var kept []string
		for _, cl := range strings.Fields(sub[1]) {
			if keepClass.MatchString(cl) {
				kept = append(kept, cl)
			}
		}
		if len(kept) == 0 {
			return ""
		}
		return ` class="` + strings.Join(kept, " ") + `"`
	})
	return strings.TrimSpace(out)
}

// CodeCSS is the stylesheet side: the component's own rules if it has any, and
// always the tokens its markup consumes.
func (c Component) CodeCSS() string {
	if c.Styles != "" {
		return strings.TrimSpace(c.Styles)
	}
	return `/* ` + c.Name + ` needs no CSS of its own: every value comes from the
   token contract in tokens.css, through the utility classes in the
   HTML + Tailwind tab.

   The tokens it reads:
     surface   --background --card --popover --muted --accent --secondary
     text      --foreground --muted-foreground --primary-foreground
     line      --border --input --ring
     structure --radius (--r-xs … --r-pill), --control-px/py, --row-px/py,
               --pad-card, --shadow-1/2, --type-sans/display/mono

   Swap those and the component follows; nothing here is hard-coded. */`
}

// AgentPrompt is what the "Copy prompt" button hands to a coding agent: enough
// context to rebuild the component exactly - the contract it consumes, the
// markup, the mechanism, and the caveats it must not quietly drop.
func (c Component) AgentPrompt(base string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Recreate the qompnt %s component exactly.\n\n", c.Name)
	fmt.Fprintf(&b, "WHAT IT IS\n%s\n\n", c.Blurb)

	if c.Mechanism != "" {
		fmt.Fprintf(&b, "HOW IT WORKS\n%s\n\n", strings.ReplaceAll(c.Mechanism, "`", ""))
	}

	b.WriteString(`STYLING CONTRACT
Use only these CSS variables - never a literal colour, radius or pad. They are
defined in tokens.css and swapped per design system:
  colour     --background --foreground --card --card-foreground --popover
             --popover-foreground --primary --primary-foreground --secondary
             --secondary-foreground --muted --muted-foreground --accent
             --accent-foreground --destructive --border --input --ring
  structure  --radius (scale: --r-xs/sm/md/lg/xl/pill), --control-px, --control-py,
             --shadow-1, --shadow-2, --type-sans, --type-display, --type-mono
The markup below uses utility classes bound to those variables: bg-background,
text-muted-foreground, border-border, rounded-md, px-control-x, py-control-y,
shadow-e1. Keep them; do not substitute hard-coded Tailwind colours or spacing,
or the component will ignore a theme change.

`)

	fmt.Fprintf(&b, "MARKUP\n```html\n%s\n```\n\n", strings.TrimSpace(demoHooksRe.ReplaceAllString(c.PreviewSource, "")))

	if len(c.Caveats) > 0 {
		b.WriteString("CONSTRAINTS - keep these true\n")
		for _, cav := range c.Caveats {
			fmt.Fprintf(&b, "- %s\n", strings.ReplaceAll(cav, "`", ""))
		}
		b.WriteString("\n")
	}
	if len(c.Support) > 0 {
		b.WriteString("BROWSER SUPPORT\n")
		for _, sup := range c.Support {
			fmt.Fprintf(&b, "- %s\n", strings.ReplaceAll(sup, "`", ""))
		}
		b.WriteString("\n")
	}

	b.WriteString(`RULES
- No framework, no runtime dependency. Plain HTML plus the utility classes above.
- Prefer the native element over a rebuilt one; if you replace a native control,
  say what accessibility you gave up.
- Any JavaScript must be delegated from document and guarded so pasting the
  snippet twice binds once.
- No transitions.
`)
	fmt.Fprintf(&b, "\nSource: %s/c/%s\n", base, c.Slug)
	return b.String()
}

// demoHooksRe strips the data-x-* attributes this site uses to drive its options
// panel - scaffolding, not part of the component.
var demoHooksRe = regexp.MustCompile(`\s+data-x-[a-z-]+(="[^"]*")?`)

// find returns the component with the given slug.
func find(cs []Component, slug string) (Component, bool) {
	for _, c := range cs {
		if c.Slug == slug {
			return c, true
		}
	}
	return Component{}, false
}

// search filters by name, blurb, and tags, case-insensitively.
func search(cs []Component, q string) []Component {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return cs
	}
	var out []Component
	for _, c := range cs {
		hay := strings.ToLower(c.Name + " " + c.Blurb + " " + strings.Join(c.Tags, " "))
		if strings.Contains(hay, q) {
			out = append(out, c)
		}
	}
	return out
}
