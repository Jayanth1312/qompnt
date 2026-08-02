package main

import (
	"fmt"
	"os"
	"strings"
)

func runInit(registry string, opts initOpts) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if configExists(cwd) {
		return fmt.Errorf("qomp.json already exists in %s (use qomp add / qomp update)", cwd)
	}

	client := newClient(registry)
	fmt.Fprintf(os.Stderr, "Fetching registry from %s …\n", client.Base)
	m, err := client.fetchManifest()
	if err != nil {
		return err
	}

	var answers initAnswers
	if opts.nonInteractive() {
		answers, err = answersFromFlags(m, opts)
		if err != nil {
			return err
		}
	} else {
		answers, err = runInitWizard(m)
		if err != nil {
			return err
		}
	}

	cfg := defaultConfig(client.Base)
	cfg.Theme = answers.Theme
	cfg.Accent = answers.Accent
	slugs := resolveSlugs(m, answers.Mode, answers.Selected)

	fmt.Fprintf(os.Stderr, "Installing theme %q …\n", cfg.Theme)
	if err := installStyles(cwd, cfg, client); err != nil {
		return err
	}

	cfg.Hashes = map[string]string{}
	cfg.Hashes["styles"] = "styles:" + cfg.Theme + ":" + cfg.Accent + ":" + m.Version
	if t, ok := client.findTheme(cfg.Theme); ok && t.Hash != "" {
		cfg.Hashes["theme"] = t.Hash
	}
	cfg.Hashes["manifest"] = m.Version

	for _, slug := range slugs {
		fmt.Fprintf(os.Stderr, "Installing %s …\n", slug)
		if err := installComponent(cwd, cfg, client, slug); err != nil {
			return err
		}
		ensureComponentListed(&cfg, slug)
		if c, ok := client.findComponent(slug); ok {
			cfg.Hashes["component:"+slug] = c.Hash
		}
	}

	if err := saveConfig(cwd, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\n", configFileName)
	fmt.Fprintf(os.Stderr, "Components: %s\n", cfg.Paths.Components)
	fmt.Fprintf(os.Stderr, "Styles:     %s\n", cfg.Paths.Styles)
	if len(cfg.Components) == 0 {
		fmt.Fprintln(os.Stderr, "No components installed. Run: qomp add <name>")
	} else {
		fmt.Fprintf(os.Stderr, "Installed %d component(s).\n", len(cfg.Components))
	}
	return nil
}

type initOpts struct {
	Theme      string
	Accent     string
	Components string // all | minimal | none | comma-separated slugs
}

func (o initOpts) nonInteractive() bool {
	return o.Theme != "" || o.Components != ""
}

func answersFromFlags(m *Manifest, opts initOpts) (initAnswers, error) {
	a := initAnswers{
		Theme:  opts.Theme,
		Accent: strings.TrimSpace(opts.Accent),
	}
	if a.Theme == "" {
		a.Theme = "claude"
	}
	if _, ok := findThemeID(m, a.Theme); !ok {
		return a, fmt.Errorf("unknown theme %q", a.Theme)
	}
	if a.Accent != "" && !hexColorRe.MatchString(a.Accent) {
		return a, fmt.Errorf("use a hex color like #0071e3")
	}

	comp := strings.TrimSpace(opts.Components)
	if comp == "" {
		comp = "minimal"
	}
	switch comp {
	case "all", "minimal", "none":
		a.Mode = comp
	default:
		a.Mode = "selective"
		for _, part := range strings.Split(comp, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if _, ok := findCompSlug(m, part); !ok {
				return a, fmt.Errorf("no component named %q available", part)
			}
			a.Selected = append(a.Selected, part)
		}
		if len(a.Selected) == 0 {
			return a, fmt.Errorf("--components needs all, minimal, none, or a comma-separated list")
		}
	}
	return a, nil
}

func findThemeID(m *Manifest, id string) (Theme, bool) {
	for _, t := range m.Themes {
		if t.ID == id {
			return t, true
		}
	}
	return Theme{}, false
}

func findCompSlug(m *Manifest, slug string) (Component, bool) {
	for _, c := range m.Components {
		if c.Slug == slug {
			return c, true
		}
	}
	return Component{}, false
}
