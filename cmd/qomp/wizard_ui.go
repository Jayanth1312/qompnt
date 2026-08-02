package main

import (
	"fmt"
	"strings"
	"unicode"

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

func (m *searchPicker) finishMulti() (tea.Model, tea.Cmd) {
	m.done = true
	return m, tea.Quit
}

func (m *searchPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = max(20, msg.Width)
		m.height = max(5, msg.Height-6)
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		case "backspace":
			if len(m.filter) > 0 {
				// Drop last rune, not last byte.
				r := []rune(m.filter)
				m.filter = string(r[:len(r)-1])
				m.refilter()
			}
		case "ctrl+enter", "ctrl+j", "ctrl+m", "tab":
			// ctrl+enter: Windows/conhost often emits ctrl+j / ctrl+m instead.
			// tab is a reliable fallback when the terminal swallows ctrl+enter.
			if m.multi {
				return m.finishMulti()
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
			// Never let ctrl/alt/meta chords or non-printables poison the filter
			// (that emptied the component list when Ctrl was pressed on Windows).
			if strings.HasPrefix(key, "ctrl+") || strings.HasPrefix(key, "alt+") || strings.HasPrefix(key, "shift+") {
				break
			}
			if msg.Type != tea.KeyRunes {
				break
			}
			added := false
			for _, r := range msg.Runes {
				if unicode.IsPrint(r) && !unicode.IsSpace(r) {
					m.filter += string(r)
					added = true
				}
			}
			if added {
				m.refilter()
			}
		}
	}
	return m, nil
}

func (m *searchPicker) refilter() {
	q := strings.ToLower(strings.TrimSpace(m.filter))
	// Strip leftover non-printables so a bad key never leaves the list empty.
	q = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, q)
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

	help := "type to search · ↑/↓ move"
	if m.multi {
		help += " · enter/space toggle · tab or ctrl+enter done"
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
