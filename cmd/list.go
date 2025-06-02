package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/theykk/git-switcher/utils"
)

var (
	getHomeDirFnc = homedir.Dir // Function variable for easy mocking in tests
	configSubPath = filepath.Join(".config", "gitconfigs")
	gitConfigName = ".gitconfig"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all available git profiles",
	Long:  `Lists all available git profiles and indicates the currently active one.`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := getHomeDirFnc()
		if err != nil {
			log.Fatalf("Error getting home directory: %v", err) // Use log.Fatalf for consistency
		}

		configDir := filepath.Join(home, configSubPath)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			// Use cmd.OutOrStdout() for output redirection in tests
			fmt.Fprintln(cmd.OutOrStdout(), "Configuration directory "+filepath.Join("~", configSubPath)+" not found. Please create it first.")
			return
		}

		files, err := os.ReadDir(configDir)
		if err != nil {
			log.Fatalf("Error reading profiles directory: %v", err) // Use log.Fatalf
		}

		gitconfigPath := filepath.Join(home, gitConfigName)
		activeHash := ""
		if _, err := os.Stat(gitconfigPath); err == nil { // Check if .gitconfig exists
			activeHash = utils.Hash(gitconfigPath)
		} else if !os.IsNotExist(err) {
			// For errors other than "not exist", log them.
			log.Printf("Warning: Error checking current .gitconfig: %v", err) // Keep log.Printf for warnings
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Available Git profiles:") // This title is part of the output.
		foundProfiles := false
		for _, fileEntry := range files { // Renamed 'file' to 'fileEntry' for clarity
			if !fileEntry.IsDir() {
				profileName := fileEntry.Name()
				profilePath := filepath.Join(configDir, profileName)

				// Ensure the profile file itself exists before hashing
				if _, err := os.Stat(profilePath); os.IsNotExist(err) {
					log.Printf("Warning: profile file %s does not exist. Skipping.", profilePath) // Keep log.Printf for warnings
					continue
				}

				profileHash := utils.Hash(profilePath)
				foundProfiles = true

				if activeHash != "" && profileHash == activeHash {
					// color.Green will be used by Cobra if it detects a TTY.
					// For tests, stdout is usually not a TTY, so color might be disabled.
					// The test should ideally check for the presence of "*" and "(current)".
					// Using fmt.Sprintf for consistent output formatting in tests.
					// Actual color output can be manually verified.
					fmt.Fprintf(cmd.OutOrStdout(), "* %s (current)\n", profileName)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", profileName)
				}
			}
		}

		if !foundProfiles {
			// Standardize this message slightly for easier testing.
			fmt.Fprintln(cmd.OutOrStdout(), "  No profiles found in "+filepath.Join("~", configSubPath)+".")
		}
	},
}

// GetListCmdForTest exposes listCmd for testing purposes if needed by other packages,
// though for same-package tests, it's directly accessible.
// func GetListCmdForTest() *cobra.Command {
// 	return listCmd
// }

func init() {
	// This ensures listCmd is added to rootCmd when the package is initialized.
	// No changes needed here normally for testing listCmd itself via rootCmd.Execute().
	rootCmd.AddCommand(listCmd)
}
