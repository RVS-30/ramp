package cmd

import (
	"fmt"

	"github.com/RVS-30/ramp/internal/analyser"
	"github.com/RVS-30/ramp/internal/output"
	"github.com/spf13/cobra"
)

var detailedFlag bool

var analyseCmd = &cobra.Command{
	Use:   "analyse [path]",
	Short: "Analyse a project's language, framework, and stats",
	Long: `Analyse inspects a directory and reports its primary language,
framework, and version by checking for known manifest files
(go.mod, package.json, Cargo.toml, pyproject.toml, pom.xml, build.gradle).

If no path is given, the current directory is used. If no manifest
is found, ramp falls back to scanning source files by extension.`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true, // don't re-render the full help block on error
	SilenceErrors: true, // let root.go's Execute() print the one error line
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		info, err := analyser.AnalyseProject(root)
		if err != nil {
			return fmt.Errorf("could not analyse %q: %w", root, err)
		}
		output.Print(info)

		if detailedFlag {
			detail, err := analyser.DetailedStats(root)
			if err != nil {
				return fmt.Errorf("could not compute detailed stats for %q: %w", root, err)
			}
			fmt.Println()
			output.PrintDetailed(detail)
		}
		return nil
	},
}

func init() {
	analyseCmd.Flags().BoolVarP(&detailedFlag, "detailed", "d", false, "Show language breakdown and file statistics")
	rootCmd.AddCommand(analyseCmd)
}
