package main

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
)

var hexColorRe = regexp.MustCompile(`(?i)^#([0-9a-f]{3}|[0-9a-f]{6})$`)

type initAnswers struct {
	Theme    string
	Accent   string
	Mode     string // all | minimal | selective | none
	Selected []string
}

func runInitWizard(m *Manifest) (initAnswers, error) {
	var a initAnswers

	themeOpts := make([]huh.Option[string], 0, len(m.Themes))
	for _, t := range m.Themes {
		label := t.Title
		if t.ID == "claude" {
			label = "Claude (default)"
		}
		themeOpts = append(themeOpts, huh.NewOption(label, t.ID))
	}

	compOpts := make([]huh.Option[string], 0, len(m.Components))
	for _, c := range m.Components {
		label := c.Title
		if c.Description != "" {
			label = c.Title + " — " + truncate(c.Description, 50)
		}
		compOpts = append(compOpts, huh.NewOption(label, c.Slug))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pick a design system").
				Description("One stylesheet swaps the whole look. Enter to continue.").
				Options(themeOpts...).
				Value(&a.Theme),
		).WithHeight(18),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Install components").
				Description("Choose what to copy into your project. Enter to finish.").
				Options(
					huh.NewOption("All components", "all"),
					huh.NewOption("Minimal set", "minimal"),
					huh.NewOption("Pick individually", "selective"),
					huh.NewOption("None — theme only", "none"),
				).
				Value(&a.Mode),
		).WithHeight(12),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select components").
				Description("Space to toggle · / to filter · enter when done").
				Options(compOpts...).
				Value(&a.Selected).
				Filterable(true),
		).
			WithHideFunc(func() bool { return a.Mode != "selective" }).
			WithHeight(20),
	).WithTheme(wizardTheme()).WithWidth(64)

	if err := form.Run(); err != nil {
		return a, err
	}
	return a, nil
}

// wizardTheme loosens vertical rhythm for a calmer step-by-step flow.
func wizardTheme() *huh.Theme {
	t := huh.ThemeCharm()
	t.Group.Title = t.Group.Title.MarginBottom(1)
	t.Group.Description = t.Group.Description.MarginBottom(1)
	t.Focused.Title = t.Focused.Title.MarginTop(1)
	t.Blurred.Title = t.Blurred.Title.MarginTop(1)
	t.Focused.Base = t.Focused.Base.PaddingTop(0).PaddingBottom(0)
	t.Blurred.Base = t.Blurred.Base.PaddingTop(0).PaddingBottom(0)
	t.Help.ShortKey = t.Help.ShortKey.MarginTop(1)
	t.Help.ShortDesc = t.Help.ShortDesc.MarginTop(1)
	return t
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
