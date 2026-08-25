package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KillCandidate is one selectable row in the picker — deliberately
// decoupled from discovery.DevPort/DatabaseMatch so this package has
// zero dependency on discovery's internals.
type KillCandidate struct {
	PID   int
	Port  int
	Label string
	Stack string
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Bold(true)

	chevronStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	selectedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	plainTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
)

type pickerModel struct {
	candidates []KillCandidate
	selected   int
	choice     *KillCandidate
	quit       bool
}

// newPickerModel creates a simple inline picker.
//
// Unlike bubbles/list, this does not have a viewport. Every candidate
// is rendered directly, so the picker never internally scrolls.
func newPickerModel(candidates []KillCandidate) pickerModel {
	return pickerModel{
		candidates: candidates,
		selected:   0,
	}
}

func (m pickerModel) Init() tea.Cmd {
	return nil
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit

		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.selected < len(m.candidates)-1 {
				m.selected++
			}

		case "home":
			m.selected = 0

		case "end":
			if len(m.candidates) > 0 {
				m.selected = len(m.candidates) - 1
			}

		case "enter":
			if len(m.candidates) > 0 {
				c := m.candidates[m.selected]
				m.choice = &c
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m pickerModel) View() string {
	if len(m.candidates) == 0 {
		return titleStyle.Render("Select a process to kill")
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render("Select a process to kill"))
	b.WriteString("\n\n")

	// Find the longest label so the stack/pid description starts
	// at the same column for every row.
	maxLabelWidth := 0
	for _, c := range m.candidates {
		if len(c.Label) > maxLabelWidth {
			maxLabelWidth = len(c.Label)
		}
	}

	for i, c := range m.candidates {
		cursor := "  "

		// Keep port + process name exactly in the existing title colors.
		//
		// Port is six characters wide, then the process/project label.
		titleText := fmt.Sprintf(
			"%-6d %-*s",
			c.Port,
			maxLabelWidth,
			c.Label,
		)

		title := plainTitleStyle.Render(titleText)

		if i == m.selected {
			cursor = chevronStyle.Render("> ")
			title = selectedTitleStyle.Render(titleText)
		}

		// Keep the existing dim/italic styling for:
		// "Cobra · pid 43160"
		desc := descStyle.Render(
			fmt.Sprintf("%s · pid %d", c.Stack, c.PID),
		)

		b.WriteString(cursor)
		b.WriteString(title)
		b.WriteString("  ")
		b.WriteString(desc)

		if i < len(m.candidates)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// RunKillPicker launches the interactive picker and returns
// the candidate the user selected, or nil if they cancelled.
func RunKillPicker(candidates []KillCandidate) (*KillCandidate, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	m := newPickerModel(candidates)

	// No WithAltScreen:
	// the picker stays inline in the terminal rather than taking over
	// the entire screen.
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	final := finalModel.(pickerModel)

	if final.quit {
		return nil, nil
	}

	return final.choice, nil
}