/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	// "io/fs" // Removed unused import
	// "log" // Removed unused import
	"os"
	// "path/filepath" // Removed unused import

	// "github.com/fatih/color" // Removed unused import
	// "github.com/manifoldco/promptui" // Removed unused import
	// "github.com/mitchellh/go-homedir" // Removed unused import
	"github.com/spf13/cobra"
	// "github.com/theykk/git-switcher/utils" // Removed unused import
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
		// Display help or a brief usage message
		// For example, cmd.Help() can be used if you want to show the full help.
		// Or a custom message:
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-switcher.yaml)")

	// Add subcommands
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(listCmd)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle") // Example flag removed
}


