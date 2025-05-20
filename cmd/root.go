/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-switcher",
	Short: "A tool to easily switch between different git configurations.",
	Long: `git-switcher allows you to manage multiple git configurations
and switch between them with a simple interactive prompt or direct commands.

This is useful when you work on different projects that require
different user names or email addresses for git commits.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This is the default action when no subcommand is provided.
		// This logic is taken from the original main() function, after os.Args parsing.

		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}

		log.SetFlags(log.Lshortfile)
		configs := make(map[string]string)

		if _, err := os.Stat(confPath); os.IsNotExist(err) {
			err = os.MkdirAll(confPath, os.ModeDir|0o700)
			if err != nil {
				log.Println(err) // Log and continue, maybe it's just a read-only system
			}
		}

		err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
			if d.IsDir() {
				return nil
			}
			if e != nil { // Check for errors from WalkDir itself
				log.Printf("Warning: error accessing path %s: %v\n", path, e)
				return e // or return nil to attempt to continue
			}
			configs[utils.Hash(path)] = filepath.Base(path)
			return nil
		})
		if err != nil {
			log.Printf("Error walking directory %s: %v\n", confPath, err)
			// Decide if this is fatal or if the program can continue (e.g. if it's just for listing)
		}

		gitConfig, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		if _, err := os.Stat(gitConfig); os.IsNotExist(err) {
			utils.Write(gitConfig, []byte("[user]\n\tname = username"))
		}
		gitConfigHash := utils.Hash(gitConfig)

		// Ensure old-configs link is handled (idempotently)
		// This part was originally before os.Args check.
		// It makes sense to ensure the current .gitconfig is backed up if it's not a known profile.
		if _, ok := configs[gitConfigHash]; !ok {
			oldConfigsPath := filepath.Join(confPath, "old-configs")
			// Check if .gitconfig is not already a symlink before attempting to link it
			// This avoids linking a symlink itself if .gitconfig is already managed.
			lstatInfo, lstatErr := os.Lstat(gitConfig)
			isSymlink := false
			if lstatErr == nil && (lstatInfo.Mode()&os.ModeSymlink != 0) {
				isSymlink = true
			}

			if !isSymlink { // Only try to link if .gitconfig is a regular file
				if _, statErr := os.Stat(oldConfigsPath); os.IsNotExist(statErr) {
					errLink := os.Link(gitConfig, oldConfigsPath)
					if errLink != nil {
						log.Printf("Warning: Failed to link current .gitconfig to %s: %v\n", oldConfigsPath, errLink)
					} else {
						log.Printf("Info: Current .gitconfig backed up to %s\n", oldConfigsPath)
						// Add the newly backed-up config to the current session's list
						configs[utils.Hash(oldConfigsPath)] = filepath.Base(oldConfigsPath) 
					}
				} else if statErr == nil {
					log.Printf("Info: %s already exists. Current .gitconfig not linked as old-configs.\n", oldConfigsPath)
				}
			} else {
				log.Printf("Info: Current .gitconfig at %s is a symlink, not backing up to old-configs.\n", gitConfig)
			}
		}
		
		// Re-populate configs map after potential backup, to ensure `old-configs` is listed if created.
		// This is a bit redundant if the backup didn't happen or already existed, but ensures consistency.
		configs = make(map[string]string) // Reset before re-populating
		err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
			if d.IsDir() { return nil }
			if e != nil { log.Printf("Warning: error accessing path %s: %v\n", path, e); return e }
			configs[utils.Hash(path)] = filepath.Base(path)
			return nil
		})
		if err != nil { log.Printf("Error re-walking directory %s: %v\n", confPath, err) }


		var profiles []string
		var currentConfigPos int = -1 // Initialize to -1 to indicate not found
		i := 0
		currentConfigFilename := "unknown (current .gitconfig may not be a saved profile)"
		
		// Check if gitConfigHash is valid and present in configs
		// This might happen if .gitconfig is empty or unreadable initially
		_, gitConfigHashOk := configs[gitConfigHash]
		if gitConfigHashOk {
			currentConfigFilename = configs[gitConfigHash]
		}


		for hash, val := range configs {
			if hash == gitConfigHash {
				currentConfigPos = i
			}
			profiles = append(profiles, val)
			i++
		}
		
		if len(profiles) == 0 {
			fmt.Printf("No git configuration profiles found in %s.\n", confPath)
			fmt.Println("You can create one using 'git-switcher create' (once implemented as a subcommand).")
			return
		}
		
		// If currentConfigPos remained -1, it means .gitconfig's hash wasn't in `configs`.
		// In promptui, CursorPos defaults to 0 if out of bounds, which is fine.
		// But the label should be accurate.
		selectLabel := "Select Git Config"
		if currentConfigPos != -1 {
			selectLabel += " (Current: " + currentConfigFilename + ")"
		} else {
			selectLabel += " (Current: " + currentConfigFilename + " - not in saved profiles)"
		}


		prompt := promptui.Select{
			Label:        selectLabel,
			Items:        profiles,
			CursorPos:    currentConfigPos, // promptui handles -1 by defaulting to 0
			HideSelected: true,
		}

		_, result, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				fmt.Println("Operation cancelled.")
				os.Exit(0)
			}
			fmt.Printf("Prompt failed %v\n", err)
			os.Exit(1)
		}
		newConfig := result

		err = os.Remove(gitConfig)
		if err != nil && !os.IsNotExist(err) {
			log.Panic(err)
		}

		err = os.Symlink(filepath.Join(confPath, newConfig), gitConfig)
		if err != nil {
			log.Panic(err)
		}
		color.HiBlue("Switched to profile %q", newConfig)
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle") // Example flag removed
}


