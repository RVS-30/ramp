package output

import (
	"fmt"
	"strings"

	"github.com/RVS-30/ramp/internal/discovery"

	"github.com/charmbracelet/lipgloss"
)

var (
	killPromptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	killSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")). // green
				Bold(true)

	killFailStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("204")). // red/pink
			Bold(true)

	killCancelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	killLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
)

var (
	// Section headers: bold, same grey family as analyse's labelStyle
	// but bold to read as a heading, matching sectionHeaderStyle's
	// existing role in output.go.
	portsSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Bold(true)

	// Port numbers are the "find this fast" column — cyan, same hue
	// family as your subHeadingStyle (86), draws the eye first.
	portNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Width(7)

	// Project/database names: bold plain white, same weight as
	// valueStyle in the analyse output.
	portNameStyle = lipgloss.NewStyle().
			Bold(true).
			Width(26)

	// Stack/framework label: orange, matching descriptionStyle's hue
	// (214) so it visually pairs with how analyse shows Framework.
	portStackStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	// PID and other secondary metadata: dim grey, same role as
	// emptyStyle — present but visually de-emphasized.
	portMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

	// Source tag ("Docker" / "Local") gets its own subtle accent so
	// it reads as a badge rather than blending into the row.
	portSourceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	emptySectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)
)

// PrintPorts renders a DiscoveryResult in the same visual language as
// Print/PrintDetailed: aligned columns, muted metadata, accent colors
// reserved for the information that actually matters at a glance.
func PrintPorts(result *discovery.DiscoveryResult, containers []discovery.Container) {
	printDevPortsSection(result.DevPorts)
	fmt.Println()
	printDatabasesSection(result.Databases)

	if len(containers) > 0 {
		fmt.Println()
		printContainersSection(containers)
	}
}

func printDevPortsSection(ports []discovery.DevPort) {
	fmt.Println(portsSectionStyle.Render("DEV PORTS"))

	if len(ports) == 0 {
		fmt.Println("  " + emptySectionStyle.Render("No dev servers found — try ramp ports --all"))
		return
	}

	for _, p := range ports {
		fmt.Printf("  %s %s %s\n",
			portNumberStyle.Render(fmt.Sprintf("%d", p.Port)),
			portNameStyle.Render(p.Project),
			portStackStyle.Render(nonEmpty(p.Stack, "—")),
		)
	}
}

func printDatabasesSection(dbs []discovery.DatabaseMatch) {
	fmt.Println(portsSectionStyle.Render("DATABASES"))

	if len(dbs) == 0 {
		fmt.Println("  " + emptySectionStyle.Render("None running"))
		return
	}

	for _, db := range dbs {
		fmt.Printf("  %s %s %s\n",
			portNumberStyle.Render(fmt.Sprintf("%d", db.Port)),
			portNameStyle.Render(db.Name),
			portSourceStyle.Render(db.Source),
		)
	}
}

func printContainersSection(containers []discovery.Container) {
	fmt.Println(portsSectionStyle.Render("CONTAINERS"))

	for _, c := range containers {
		fmt.Printf("  %s\n", portMetaStyle.Render(c.Name))
	}
}

func nonEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// PrintKillPrompt renders the confirmation line before a kill, with
// the process label in cyan (matching portNumberStyle's family) and
// the port itself calling out clearly.
func PrintKillPrompt(label string, pid, port int) {
	fmt.Printf("Kill %s (pid %s) on port %d? [y/N] ",
		killLabelStyle.Render(label),
		killPromptStyle.Render(fmt.Sprintf("%d", pid)),
		port,
	)
}

// PrintKillResult renders the outcome of a Terminate call.
func PrintKillResult(killed bool, message string) {
	if killed {
		fmt.Println(killSuccessStyle.Render("✓ " + message))
		return
	}
	fmt.Println(killFailStyle.Render("✗ " + message))
}

// PrintKillCancelled renders a cancelled kill.
func PrintKillCancelled() {
	fmt.Println(killCancelStyle.Render("Cancelled."))
}

// PrintKillNotFound renders when no process is found on the port.
func PrintKillNotFound(port int) {
	fmt.Println(killCancelStyle.Render(fmt.Sprintf("Nothing found listening on port %d", port)))
}
