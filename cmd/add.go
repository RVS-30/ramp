/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Aliases: []string{"addition"},
	Short: "Ramp add command performs addition of two numbers",
	Long: `Ramp add command performs addition of two numbers.
It takes two numbers as input and returns their sum as output.
The numbers can be provided as command line arguments or through standard input.
For example:
ramp add 2 3`,
DisableFlagParsing: true,
    Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Addition of %s and %s is: %s\n", args[0], args[1], Add(args[0], args[1]))
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
