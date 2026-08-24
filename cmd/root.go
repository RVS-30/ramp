/*
Copyright © 2026 Rajveer Singh <rajveer104c@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ramp",
	Short: "Ramp is a simple CLI calculator",
	Long: `Ramp is a simple CLI calculator that can perform basic arithmetic operations like 
addition, subtraction, multiplication, and division.
It is designed to be easy to use and efficient for quick calculations.`,
	Version: "1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Ramp 🚀")
	},
}

var (
	headingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	subHeadingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86"))

	usageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

)

func hasNonHelpFlags(cmd *cobra.Command) bool {
	count := 0

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}

		count++
	})

	return count > 0
}

func nonFlagsUsage(cmd *cobra.Command) string {
	parts := strings.Split(cmd.Use, " ")

	if len(parts) > 1 {
		return strings.Join(parts[1:], " ")
	}

	return ""
}

var customHelpTemplate = `
{{heading "Help for"}} {{.Name}}

{{subHeading "Description:"}}
{{description .Long}}

{{subHeading "Usage:"}}
  {{usageLine .}}

{{subHeading "Aliases:"}}
  {{.NameAndAliases}}

{{if .HasAvailableLocalFlags}}{{subHeading "Flags:"}}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`


func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Oops, Something is wrong: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.AddTemplateFunc("heading", func(s string) string {
		return headingStyle.Render(s)
	})

	cobra.AddTemplateFunc("subHeading", func(s string) string {
		return subHeadingStyle.Render(s)
	})

	cobra.AddTemplateFunc("usage", func(s string) string {
		return usageStyle.Render(s)
	})

	cobra.AddTemplateFunc("description", func(s string) string {
		return descriptionStyle.Render(s)
	})

	cobra.AddTemplateFunc("HasNonHelpFlags", hasNonHelpFlags)
	cobra.AddTemplateFunc("NonFlagsUsage", nonFlagsUsage)

	cobra.AddTemplateFunc("usageLine", func(cmd *cobra.Command) string {
		var lines []string

		if cmd.Runnable() {
			usage := cmd.CommandPath()

			if hasNonHelpFlags(cmd) {
				usage += " [flags]"
			}

			if nonFlags := nonFlagsUsage(cmd); nonFlags != "" {
				usage += " " + nonFlags
			}

			lines = append(lines, usage)
		}

		if cmd.HasAvailableSubCommands() {
			lines = append(lines, cmd.CommandPath()+" [command]")
		}

		return usageStyle.Render(strings.Join(lines, "\n"))
	})

	// cobra.AddTemplateFunc("HasNonHelpFlags", func(cmd *cobra.Command) bool {
	// 	count := 0

	// 	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
	// 		if f.Name != "help" && f.Shorthand != "h" {
	// 			count++
	// 		}
	// 	})

	// 	return count > 0
	// })

	// cobra.AddTemplateFunc("NonFlagsUsage", func(cmd *cobra.Command) string {
	// 	parts := strings.Split(cmd.Use, " ")

	// 	if len(parts) > 1 {
	// 		return strings.Join(parts[1:], " ")
	// 	}

	// 	return ""
	// })

	rootCmd.SetHelpTemplate(customHelpTemplate)
	rootCmd.SetUsageTemplate(customHelpTemplate)
}
