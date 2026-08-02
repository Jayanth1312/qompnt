package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type pickerItem struct {
	label string
	value string
}

type searchPicker struct {
	searchPlaceholder string
	title             string
	items             []pickerItem
	filtered          []pickerItem
	filter            string
	cursor            int
	multi             bool
	selected          map[string]bool
	width             int
	height            int
	done              bool
	quitting          bool
	resultSingle      string
}

func newSearchPicker(searchPlaceholder, title string, items []pickerItem, multi bool) *searchPicker {
	m := &searchPicker{
		searchPlaceholder: searchPlaceholder,
		title:             title,
		items:             items,
		filtered:          append([]pickerItem(nil), items...),
		multi:             multi,
		selected:          map[string]bool{},
		width:             64,
		height:            20,
	}
	if !multi && len(items) > 0 {
		m.resultSingle = items[0].value
	}
	return m
}

func (m *searchPicker) Init() tea.Cmd { return nil }

func (m *searchPicker) toggleCursor() {
	if len(m.filtered) == 0 {
		return
	}
	v := m.filtered[m.cursor].value
	m.selected[v] = !m.selected[v]
}

func (m *searchPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "backspace":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.refilter()
			}
		case "ctrl+enter", "ctrl+j":
			// Terminals often send ctrl+j for ctrl+enter.
			if m.multi {
				m.done = true
				return m, tea.Quit
			}
		case "enter":
			if m.multi {
				m.toggleCursor()
			} else if len(m.filtered) > 0 {
				m.resultSingle = m.filtered[m.cursor].value
				m.done = true
				return m, tea.Quit
			}
		case " ":
			if m.multi {
				m.toggleCursor()
			} else if len(m.filtered) > 0 {
				m.resultSingle = m.filtered[m.cursor].value
			}
		case "esc":
			m.filter = ""
			m.refilter()
		default:
			if k := msg.String(); len(k) == 1 && k != " " {
				m.filter += k
				m.refilter()
			}
		}
	}
	return m, nil
}

func (m *searchPicker) refilter() {
	q := strings.ToLower(strings.TrimSpace(m.filter))
	if q == "" {
		m.filtered = append(m.filtered[:0], m.items...)
	} else {
		m.filtered = m.filtered[:0]
		for _, it := range m.items {
			if strings.Contains(strings.ToLower(it.label), q) {
				m.filtered = append(m.filtered, it)
			}
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m *searchPicker) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	searchText := m.filter
	if searchText == "" {
		searchText = m.searchPlaceholder
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("["))
	if m.filter == "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render(" " + searchText + " "))
	} else {
		b.WriteString(" " + searchText + " ")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("]"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(m.title))
	b.WriteString("\n")

	visible := min(len(m.filtered), m.height)
	start := 0
	if m.cursor >= visible {
		start = m.cursor - visible + 1
	}
	end := min(start+visible, len(m.filtered))

	for i := start; i < end; i++ {
		it := m.filtered[i]
		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> ")
		}
		var line string
		if m.multi {
			mark := "[ ]"
			if m.selected[it.value] {
				mark = "[✓]"
			}
			line = cursor + mark + " " + it.label
		} else {
			line = cursor + it.label
		}
		if i == m.cursor {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	help := "type to search"
	if m.multi {
		help += " · enter/space toggle · ctrl+enter done"
	} else {
		help += " · enter select"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("243")).MarginTop(1).Render(help))
	return b.String()
}

func runSearchPicker(searchPlaceholder, title string, items []pickerItem, multi bool) (single string, multiOut []string, err error) {
	m := newSearchPicker(searchPlaceholder, title, items, multi)
	p := tea.NewProgram(m, tea.WithAltScreen())
	final, runErr := p.Run()
	if runErr != nil {
		return "", nil, runErr
	}
	res := final.(*searchPicker)
	if res.quitting {
		return "", nil, fmt.Errorf("cancelled")
	}
	if multi {
		out := make([]string, 0)
		for _, it := range items {
			if res.selected[it.value] {
				out = append(out, it.value)
			}
		}
		return "", out, nil
	}
	return res.resultSingle, nil, nil
}
