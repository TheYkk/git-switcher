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
)

// renameCmd represents the rename command
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Renames an existing git configuration profile.",
	Long: `Renames an existing git configuration profile.
You will be prompted to select the profile to rename and then to enter the new name.
The configuration file in ~/.config/gitconfigs will be renamed.
If the renamed profile is the currently active one, the ~/.gitconfig symlink will be updated.`,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}
		gitConfigPath, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		var profiles []string
		err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
			if d.IsDir() {
				return nil
			}
			if e != nil {
				log.Printf("Warning: error accessing path %s: %v\n", path, e)
				return e
			}
			profiles = append(profiles, filepath.Base(path))
			return nil
		})
		if err != nil {
			log.Fatalf("Error listing profiles in %s: %v\n", confPath, err)
		}

		if len(profiles) == 0 {
			fmt.Println("No git configuration profiles found to rename in", confPath)
			return
		}
		
		promptOldName := promptui.Select{
			Label: "Select profile to rename",
			Items: profiles,
		}
		_, oldNameStr, err := promptOldName.Run() // Correctly capture the string result
		if err != nil {
			if err == promptui.ErrInterrupt { log.Println("Rename operation cancelled."); os.Exit(0)}
			log.Fatalf("Prompt failed %v\n", err)
		}

		promptNewName := promptui.Prompt{
			Label:   fmt.Sprintf("Enter new name for profile %q", oldNameStr),
			Validate: func(input string) error {
				if input == "" {
					return fmt.Errorf("profile name cannot be empty")
				}
				// Check if new name already exists
				for _, p := range profiles {
					if p == input && p != oldNameStr { 
						return fmt.Errorf("profile %q already exists", input)
					}
				}
				return nil
			},
		}
		newName, err := promptNewName.Run()
		if err != nil {
			if err == promptui.ErrInterrupt { log.Println("Rename operation cancelled."); os.Exit(0)}
			log.Fatalf("Prompt failed %v\n", err)
		}

		if oldNameStr == newName {
			color.HiYellow("New name is the same as the old name. No changes made.")
			return
		}

		oldProfilePath := filepath.Join(confPath, oldNameStr)
		newProfilePath := filepath.Join(confPath, newName)

		err = os.Rename(oldProfilePath, newProfilePath)
		if err != nil {
			log.Fatalf("Failed to rename profile %q to %q: %v", oldNameStr, newName, err)
		}
		color.HiGreen("Profile %q renamed to %q.", oldNameStr, newName)

		// Check if the renamed profile was the active one
		// Readlink correctly resolves the symlink path
		currentTarget, errReadLink := os.Readlink(gitConfigPath)
		if errReadLink == nil { // If .gitconfig is a symlink
			if currentTarget == oldProfilePath { // And it pointed to the old profile path
				errRemove := os.Remove(gitConfigPath)
				if errRemove != nil && !os.IsNotExist(errRemove) {
					log.Printf("Warning: failed to remove old symlink %s: %v", gitConfigPath, errRemove)
				}
				errSymlink := os.Symlink(newProfilePath, gitConfigPath)
				if errSymlink != nil {
					log.Fatalf("Failed to update symlink for active profile to %q: %v", newName, errSymlink)
				}
				color.HiBlue("Active profile symlink updated to %q.", newName)
			}
		} else if !os.IsNotExist(errReadLink) { 
            // If os.Readlink failed for a reason other than .gitconfig not existing (e.g. it's not a symlink)
            log.Printf("Warning: Could not determine if active profile needed symlink update: %v", errReadLink)
        }
	},
}

func init() {
	// Will be added in root.go
	// rootCmd.AddCommand(renameCmd)
}
