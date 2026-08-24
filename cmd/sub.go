/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// subCmd represents the sub command
var subCmd = &cobra.Command{
	Use:   "sub [num1] [num2]",
	Aliases: []string{"subtraction"},
	Short: "Ramp sub command performs subtraction of two numbers",
	DisableFlagParsing: true,
	Long: `Ramp sub command performs subtraction of two numbers.
It takes two numbers as input and returns their difference as output.
The numbers can be provided as command line arguments or through standard input.
For example:
ramp sub 5 3`,
    Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Subtraction of %s and %s is: %s\n", args[0], args[1], Subtract(args[0], args[1]))
	},
}

func init() {
	rootCmd.AddCommand(subCmd)
}
