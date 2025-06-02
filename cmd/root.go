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

var rootCmd = &cobra.Command{
	Use:   "git-switcher",
	Short: "A tool to easily switch between different git configurations.",
	Long: `git-switcher allows you to manage multiple git configurations
and switch between them with a simple interactive prompt or direct commands.

This is useful when you work on different projects that require
different user names or email addresses for git commits.`,
	Run: func(cmd *cobra.Command, args []string) {
		confPath, err := homedir.Expand("~/.config/gitconfigs")
		if err != nil {
			log.Panic(err)
		}

		log.SetFlags(log.Lshortfile)
		configs := make(map[string]string)

		if _, err := os.Stat(confPath); os.IsNotExist(err) {
			err = os.MkdirAll(confPath, os.ModeDir|0o700)
			if err != nil {
				log.Println(err)
			}
		}

		err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
			if d.IsDir() {
				return nil
			}
			if e != nil {
				log.Printf("Warning: error accessing path %s: %v\n", path, e)
				return e
			}
			configs[utils.Hash(path)] = filepath.Base(path)
			return nil
		})
		if err != nil {
			log.Printf("Error walking directory %s: %v\n", confPath, err)
		}

		gitConfig, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		if _, err := os.Stat(gitConfig); os.IsNotExist(err) {
			utils.Write(gitConfig, []byte("[user]\n\tname = username"))
		}
		gitConfigHash := utils.Hash(gitConfig)

		if _, ok := configs[gitConfigHash]; !ok {
			oldConfigsPath := filepath.Join(confPath, "old-configs")
			lstatInfo, lstatErr := os.Lstat(gitConfig)
			isSymlink := false
			if lstatErr == nil && (lstatInfo.Mode()&os.ModeSymlink != 0) {
				isSymlink = true
			}

			if !isSymlink {
				if _, statErr := os.Stat(oldConfigsPath); os.IsNotExist(statErr) {
					errLink := os.Link(gitConfig, oldConfigsPath)
					if errLink != nil {
						log.Printf("Warning: Failed to link current .gitconfig to %s: %v\n", oldConfigsPath, errLink)
					} else {
						log.Printf("Info: Current .gitconfig backed up to %s\n", oldConfigsPath)
						configs[utils.Hash(oldConfigsPath)] = filepath.Base(oldConfigsPath) 
					}
				} else if statErr == nil {
					log.Printf("Info: %s already exists. Current .gitconfig not linked as old-configs.\n", oldConfigsPath)
				}
			} else {
				log.Printf("Info: Current .gitconfig at %s is a symlink, not backing up to old-configs.\n", gitConfig)
			}
		}
		
		configs = make(map[string]string)
		err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
			if d.IsDir() { return nil }
			if e != nil { log.Printf("Warning: error accessing path %s: %v\n", path, e); return e }
			configs[utils.Hash(path)] = filepath.Base(path)
			return nil
		})
		if err != nil { log.Printf("Error re-walking directory %s: %v\n", confPath, err) }

		var profiles []string
		var currentConfigPos int = -1
		i := 0
		currentConfigFilename := "unknown (current .gitconfig may not be a saved profile)"
		
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
			fmt.Println("You can create one using 'git-switcher create'.")
			return
		}
		
		selectLabel := "Select Git Config"
		if currentConfigPos != -1 {
			selectLabel += " (Current: " + currentConfigFilename + ")"
		} else {
			selectLabel += " (Current: " + currentConfigFilename + " - not in saved profiles)"
		}

		prompt := promptui.Select{
			Label:        selectLabel,
			Items:        profiles,
			CursorPos:    currentConfigPos,
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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(listCmd)
}


