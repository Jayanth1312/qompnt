package main

import (
	"regexp"

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

	themeItems := make([]pickerItem, 0, len(m.Themes))
	for _, t := range m.Themes {
		label := t.Title
		if t.ID == "claude" {
			label = "Claude"
		}
		themeItems = append(themeItems, pickerItem{label: label, value: t.ID})
	}

	theme, _, err := runSearchPicker("Search themes", "Select themes", themeItems, false)
	if err != nil {
		return a, err
	}
	a.Theme = theme

	modeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Install components").
				Description("Choose what to copy into your project. Enter to continue.").
				Options(
					huh.NewOption("All components", "all"),
					huh.NewOption("Minimal set", "minimal"),
					huh.NewOption("Pick individually", "selective"),
					huh.NewOption("None — theme only", "none"),
				).
				Value(&a.Mode),
		).WithHeight(12),
	).WithTheme(wizardTheme()).WithWidth(64)
	if err := modeForm.Run(); err != nil {
		return a, err
	}

	if a.Mode == "selective" {
		compItems := make([]pickerItem, 0, len(m.Components))
		for _, c := range m.Components {
			compItems = append(compItems, pickerItem{label: c.Title, value: c.Slug})
		}
		_, selected, err := runSearchPicker("Search components", "Select components", compItems, true)
		if err != nil {
			return a, err
		}
		a.Selected = selected
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
