package cmd

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/theykk/git-switcher/utils"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all available git configuration profiles.",
	Long: `Lists all available git configuration profiles stored in ~/.config/gitconfigs.
The currently active profile (if any) will be marked with an asterisk (*) and highlighted.`,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}

		gitConfigPath, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		configs := make(map[string]string)
		var profiles []string

		if _, err := os.Stat(confPath); os.IsNotExist(err) {
			fmt.Printf("No git configuration directory found at %s.\n", confPath)
			fmt.Println("You can create your first profile using 'git-switcher create'.")
			return
		}

		err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
			if d.IsDir() {
				return nil
			}
			if e != nil {
				log.Printf("Warning: error accessing path %s: %v\n", path, e)
				return e
			}
			baseName := filepath.Base(path)
			configs[utils.Hash(path)] = baseName
			profiles = append(profiles, baseName)
			return nil
		})
		if err != nil {
			log.Printf("Error walking directory %s: %v\n", confPath, err)
			return
		}

		if len(profiles) == 0 {
			fmt.Printf("No git configuration profiles found in %s.\n", confPath)
			fmt.Println("You can create your first profile using 'git-switcher create'.")
			return
		}

		var currentProfile string
		gitConfigHash := ""
		if _, err := os.Stat(gitConfigPath); err == nil {
			gitConfigHash = utils.Hash(gitConfigPath)
			if profileName, ok := configs[gitConfigHash]; ok {
				currentProfile = profileName
			}
		}

		fmt.Printf("Available git configuration profiles in %s:\n\n", confPath)

		for _, profile := range profiles {
			if profile == currentProfile {
				if os.Getenv("NO_COLOR") != "" {
					fmt.Printf("* %s (current)\n", profile)
				} else {
					color.HiGreen("* %s (current)", profile)
				}
			} else {
				fmt.Printf("  %s\n", profile)
			}
		}

		if currentProfile == "" && len(profiles) > 0 {
			if os.Getenv("NO_COLOR") != "" {
				fmt.Printf("\nNo active profile detected or current .gitconfig is not managed by git-switcher.\n")
			} else {
				color.HiYellow("\nNo active profile detected or current .gitconfig is not managed by git-switcher.")
			}
			fmt.Println("Use 'git-switcher switch <profile>' to activate a profile.")
		}
	},
}

func init() {
	// Will be added in root.go
} 