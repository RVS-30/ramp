package output

import (
	"fmt"
	"strings"

	"github.com/RVS-30/ramp/internal/analyser"
	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Width(11)
	valueStyle = lipgloss.NewStyle().Bold(true)
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	sectionHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Bold(true)

	langNameStyle = lipgloss.NewStyle().Width(12)
	percentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
)

// Print renders a ProjectInfo as aligned, lightly-colored key-value lines.
func Print(info *analyser.ProjectInfo) {
	rows := []struct{ label, value string }{
		{"Project", info.Name},
		{"Language", info.Language},
		{"Framework", info.Framework},
		{"Version", info.Version},
		{"Files", fmt.Sprintf("%d", info.FileCount)},
	}

	var b strings.Builder
	for _, row := range rows {
		val := row.value
		if val == "" || val == "0" {
			b.WriteString(labelStyle.Render(row.label))
			b.WriteString(emptyStyle.Render("—"))
			b.WriteString("\n")
			continue
		}
		b.WriteString(labelStyle.Render(row.label))
		b.WriteString(valueStyle.Render(val))
		b.WriteString("\n")
	}
	fmt.Print(b.String())
}

// PrintDetailed renders a language breakdown with percentages, plus
// total file and line counts. Call after Print for the base summary.
func PrintDetailed(info *analyser.DetailedInfo) {
	if len(info.Languages) == 0 {
		return
	}

	var b strings.Builder
	b.WriteString(sectionHeaderStyle.Render("Languages:"))
	b.WriteString("\n")

	for _, lang := range info.Languages {
		pct := 0.0
		if info.TotalLines > 0 {
			pct = float64(lang.Lines) / float64(info.TotalLines) * 100
		}
		b.WriteString("  ")
		b.WriteString(langNameStyle.Render(lang.Name))
		b.WriteString(percentStyle.Render(fmt.Sprintf("%5.1f%%", pct)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Files"))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", info.TotalFiles)))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Lines"))
	b.WriteString(valueStyle.Render(fmt.Sprintf("%d", info.TotalLines)))
	b.WriteString("\n")

	fmt.Print(b.String())
}