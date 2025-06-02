package cmd

import (
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/google/shlex"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Opens the current ~/.gitconfig file in your default editor.",
	Long: `Opens the currently active ~/.gitconfig file in your system's
default editor (or $EDITOR environment variable if set).
This allows you to directly modify the active git configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		gitConfigPath, err := homedir.Expand("~/.gitconfig")
		if err != nil {
			log.Panic(err)
		}

		// If .gitconfig doesn't exist, inform the user.
		// Unlike the root command, 'edit' probably shouldn't create it.
		if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
			color.HiYellow("No active .gitconfig found at %s to edit.", gitConfigPath)
			color.HiYellow("Consider switching to or creating a profile first.")
			return
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "vim" // default to vim on Unix-like systems
			}
		}

		// Use shlex to properly split editor command into parts (e.g., "code -w")
		editorParts, err := shlex.Split(editor)
		if err != nil || len(editorParts) == 0 {
			log.Fatalf("Failed to parse editor command %q: %v", editor, err)
		}

		cmdArgs := append(editorParts[1:], gitConfigPath)
		editorCmd := exec.Command(editorParts[0], cmdArgs...)

		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		color.HiBlue("Opening %s with %s...", gitConfigPath, editor)
		if err := editorCmd.Run(); err != nil {
			color.HiRed("Editor command %q failed: %s", editor, err)
		}
	},
}

func init() {
	// Will be added in root.go
	// rootCmd.AddCommand(editCmd)
}
