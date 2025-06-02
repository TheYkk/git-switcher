package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-switcher",
	Short: "A tool to easily switch between different git configurations.",
	Long: `git-switcher allows you to manage multiple git configurations
and switch between them with a simple interactive prompt or direct commands.

This is useful when you work on different projects that require
different user names or email addresses for git commits.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to git-switcher! Use 'git-switcher help' to see available commands.")
		fmt.Println("To switch profiles interactively, use 'git-switcher switch'.")
		fmt.Println("To list profiles, use 'git-switcher list'.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(listCmd)
}


