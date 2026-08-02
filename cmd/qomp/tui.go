package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

var hexColorRe = regexp.MustCompile(`(?i)^#([0-9a-f]{3}|[0-9a-f]{6})$`)

type initAnswers struct {
	Theme      string
	Accent     string
	Mode       string // all | minimal | selective | none
	Selected   []string
}

func runInitWizard(m *Manifest) (initAnswers, error) {
	var a initAnswers

	themeOpts := make([]huh.Option[string], 0, len(m.Themes))
	for _, t := range m.Themes {
		label := t.Title
		if t.ID == "claude" {
			label = "Claude (default tokens)"
		}
		themeOpts = append(themeOpts, huh.NewOption(label, t.ID))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Theme").
				Description("Design system token set").
				Options(themeOpts...).
				Value(&a.Theme),
			huh.NewInput().
				Title("Accent color").
				Description("Hex color, or leave blank to keep the theme primary").
				Placeholder("#0071e3").
				Value(&a.Accent).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return nil
					}
					if !hexColorRe.MatchString(s) {
						return fmt.Errorf("use a hex color like #0071e3")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Components").
				Options(
					huh.NewOption("All components", "all"),
					huh.NewOption("Minimal installation", "minimal"),
					huh.NewOption("Selective installation", "selective"),
					huh.NewOption("No installation", "none"),
				).
				Value(&a.Mode),
		),
	)
	if err := form.Run(); err != nil {
		return a, err
	}
	a.Accent = strings.TrimSpace(a.Accent)

	if a.Mode == "selective" {
		opts := make([]huh.Option[string], 0, len(m.Components))
		for _, c := range m.Components {
			label := c.Title
			if c.Description != "" {
				label = c.Title + " — " + truncate(c.Description, 60)
			}
			opts = append(opts, huh.NewOption(label, c.Slug))
		}
		sel := huh.NewMultiSelect[string]().
			Title("Select components").
			Description("Type to filter, space to toggle, enter to confirm").
			Options(opts...).
			Value(&a.Selected).
			Filterable(true)
		if err := huh.NewForm(huh.NewGroup(sel)).Run(); err != nil {
			return a, err
		}
	}
	return a, nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func resolveSlugs(m *Manifest, mode string, selected []string) []string {
	switch mode {
	case "none":
		return nil
	case "minimal":
		return append([]string(nil), m.Minimal...)
	case "all":
		out := make([]string, 0, len(m.Components))
		for _, c := range m.Components {
			out = append(out, c.Slug)
		}
		return out
	case "selective":
		return append([]string(nil), selected...)
	default:
		return nil
	}
}
