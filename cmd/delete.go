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
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/theykk/git-switcher/utils"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an existing git configuration profile.",
	Long: `Deletes an existing git configuration profile.
You will be prompted to select a profile to delete from the available profiles.
The selected configuration file will be removed from the ~/.config/gitconfigs directory.
If the deleted profile is the currently active one, ~/.gitconfig will also be removed.`,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}
		gitConfigPath, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		configs := make(map[string]string) // hash: filename
		var profiles []string              // list of profile filenames

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
			log.Fatalf("Error walking directory %s: %v\n", confPath, err)
		}

		if len(profiles) == 0 {
			fmt.Println("No git configuration profiles found to delete in", confPath)
			return
		}
		
		gitConfigHash := ""
		if _, errStat := os.Lstat(gitConfigPath); errStat == nil { // Use Lstat to get info about symlink itself
			gitConfigHash = utils.Hash(gitConfigPath)
		}


		currentConfigFilename := "none"
		currentConfigPos := -1
		if cfName, ok := configs[gitConfigHash]; ok {
			currentConfigFilename = cfName
			for i, pName := range profiles {
				if pName == currentConfigFilename {
					currentConfigPos = i
					break
				}
			}
		}


		promptSelect := promptui.Select{
			Label:     "Select Git Config profile to delete (Current: " + currentConfigFilename + ")",
			Items:     profiles,
			CursorPos: currentConfigPos,
		}

		_, result, err := promptSelect.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				log.Println("Delete operation cancelled.")
				os.Exit(0)
			}
			log.Fatalf("Prompt failed %v\n", err)
		}

		confirmPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Are you sure you want to delete profile %q? (Y/N)", result),
			IsConfirm: true,
		}

		_, err = confirmPrompt.Run()
		if err != nil {
			// This means user entered 'N' or something other than 'y' or 'Y'
			// For ErrInterrupt (Ctrl+C), exit. For others, assume 'N'.
			if err == promptui.ErrInterrupt {
				log.Println("Delete operation cancelled.")
				os.Exit(0)
			}
			color.HiBlue("Profile %q not deleted.", result)
			return
		}

		profileToDeletePath := filepath.Join(confPath, result)
		err = os.Remove(profileToDeletePath)
		if err != nil {
			log.Fatalf("Failed to delete profile %q: %v", result, err)
		}

		// If the deleted profile was the currently active one
		if result == currentConfigFilename {
			err = os.Remove(gitConfigPath)
			if err != nil && !os.IsNotExist(err) { // Don't panic if .gitconfig already gone
				log.Fatalf("Failed to remove current .gitconfig symlink: %v", err)
			}
			color.HiGreen("Profile %q deleted. Current .gitconfig was also removed.", result)
		} else {
			color.HiGreen("Profile %q deleted.", result)
		}
	},
}

func init() {
	// rootCmd.AddCommand(deleteCmd) // Commands are added in root.go's init
}
