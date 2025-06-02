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
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/theykk/git-switcher/utils"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new git configuration profile.",
	Long: `Creates a new git configuration profile.
You will be prompted to enter a name for the new profile.
A new configuration file will be created in the ~/.config/gitconfigs directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}

		// Ensure confPath directory exists
		if _, err := os.Stat(confPath); os.IsNotExist(err) {
			if err = os.MkdirAll(confPath, os.ModeDir|0o700); err != nil {
				log.Fatalf("Failed to create config directory %s: %v", confPath, err)
			}
		}

		prom := promptui.Prompt{
			Label: "Profile name",
		}

		result, err := prom.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				log.Println("Create operation cancelled.")
				os.Exit(0)
			}
			log.Fatalf("Prompt failed %v\n", err)
		}

		profilePath := filepath.Join(confPath, result)

		// File is not exist, write to new file
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			utils.Write(profilePath, []byte("[user]\n\tname = "+result+"\n\temail = your_email@example.com"))
			color.HiGreen("Profile %q created successfully at %s", result, profilePath)
			color.HiYellow("Please edit the file to set your desired git user name and email.")
		} else {
			color.HiRed("Profile %q already exists at %s", result, profilePath)
		}
	},
}

func init() {
	// This function is called when the package is initialized.
	// We are adding the createCmd to the rootCmd here.
	// rootCmd.AddCommand(createCmd) // Commands are added in root.go's init
}
