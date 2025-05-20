// Copyright 2021 Kaan Karakaya
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	// main_pkg "github.com/theykk/git-switcher" // Not needed for this command
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [profile_name]",
	Short: "Switches the active git configuration to the specified profile.",
	Long: `Switches the active git configuration to the specified profile.
The command takes exactly one argument: the name of the profile to switch to.
This profile must exist in the ~/.config/gitconfigs directory.
The ~/.gitconfig file will be updated to be a symlink to the selected profile.`,
	Args: cobra.ExactArgs(1), // Ensures exactly one argument (profile_name) is passed
	Run: func(cmd *cobra.Command, args []string) {
		profileName := args[0]

		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}
		gitConfigPath, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		targetProfilePath := filepath.Join(confPath, profileName)

		// Check if the target profile exists
		if _, err := os.Stat(targetProfilePath); os.IsNotExist(err) {
			color.HiRed("Error: Profile %q does not exist at %s.", profileName, targetProfilePath)
			// List available profiles for user convenience
			var availableProfiles []string
			errList := filepath.WalkDir(confPath, func(path string, d os.DirEntry, e error) error {
				if !d.IsDir() && path != targetProfilePath { // Exclude the non-existent one
					availableProfiles = append(availableProfiles, filepath.Base(path))
				}
				return nil
			})
			if errList == nil && len(availableProfiles) > 0 {
				fmt.Println("Available profiles:")
				for _, p := range availableProfiles {
					fmt.Printf("  - %s\n", p)
				}
			} else if errList != nil {
				log.Printf("Could not list available profiles: %v", errList)
			}
			os.Exit(1)
		}

		// Remove current .gitconfig symlink or file if it exists
		err = os.Remove(gitConfigPath)
		if err != nil && !os.IsNotExist(err) { // Ignore if it doesn't exist, panic for other errors
			log.Fatalf("Failed to remove existing .gitconfig at %s: %v", gitConfigPath, err)
		}

		// Create the new symlink
		err = os.Symlink(targetProfilePath, gitConfigPath)
		if err != nil {
			log.Fatalf("Failed to create symlink from %s to %s: %v", targetProfilePath, gitConfigPath, err)
		}

		color.HiBlue("Switched to profile %q. ~/.gitconfig now points to %s.", profileName, targetProfilePath)
	},
}

func init() {
	// Will be added in root.go
	// rootCmd.AddCommand(switchCmd)
}
