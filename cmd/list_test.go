package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/theykk/git-switcher/utils"
)

func TestListCmd(t *testing.T) {
	tests := []struct {
		name             string
		setupProfiles    []string
		currentProfile   string
		expectedInOutput []string
		expectError      bool
	}{
		{
			name:             "no profiles directory",
			setupProfiles:    nil,
			currentProfile:   "",
			expectedInOutput: []string{"No git configuration directory found"},
			expectError:      false,
		},
		{
			name:             "empty profiles directory",
			setupProfiles:    []string{},
			currentProfile:   "",
			expectedInOutput: []string{"No git configuration profiles found"},
			expectError:      false,
		},
		{
			name:          "single profile no current",
			setupProfiles: []string{"work"},
			currentProfile: "",
			expectedInOutput: []string{
				"Available git configuration profiles",
				"work",
				"Use 'git-switcher switch",
			},
			expectError: false,
		},
		{
			name:          "single profile with current",
			setupProfiles: []string{"work"},
			currentProfile: "work",
			expectedInOutput: []string{
				"Available git configuration profiles",
				"work (current)",
			},
			expectError: false,
		},
		{
			name:          "multiple profiles with current",
			setupProfiles: []string{"work", "personal", "opensource"},
			currentProfile: "personal",
			expectedInOutput: []string{
				"Available git configuration profiles",
				"personal (current)",
				"work",
				"opensource",
			},
			expectError: false,
		},
		{
			name:          "multiple profiles no current",
			setupProfiles: []string{"work", "personal"},
			currentProfile: "",
			expectedInOutput: []string{
				"Available git configuration profiles",
				"work",
				"personal",
				"Use 'git-switcher switch",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset homedir cache to ensure test isolation
			homedir.Reset()
			
			// Create temporary directories with unique suffix
			tempDir := t.TempDir()
			uniqueSuffix := fmt.Sprintf("_%d_%s", time.Now().UnixNano(), tt.name)
			tempDir = filepath.Join(tempDir, uniqueSuffix)
			err := os.MkdirAll(tempDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create unique temp directory: %v", err)
			}
			
			// Set environment variables to use temp directory FIRST
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", originalHome)

			// Disable color output for tests
			originalNoColor := os.Getenv("NO_COLOR")
			os.Setenv("NO_COLOR", "1")
			defer os.Setenv("NO_COLOR", originalNoColor)

			confPath := filepath.Join(tempDir, ".config", "gitconfigs")
			gitConfigPath := filepath.Join(tempDir, ".gitconfig")

			// Setup profiles if specified
			if tt.setupProfiles != nil {
				err := os.MkdirAll(confPath, 0755)
				if err != nil {
					t.Fatalf("Failed to create config directory: %v", err)
				}

				for _, profile := range tt.setupProfiles {
					profilePath := filepath.Join(confPath, profile)
					content := "[user]\n\tname = " + profile + "\n\temail = " + profile + "@example.com"
					utils.Write(profilePath, []byte(content))
				}
			}

			// Setup current profile if specified
			if tt.currentProfile != "" {
				currentProfilePath := filepath.Join(confPath, tt.currentProfile)
				if _, err := os.Stat(currentProfilePath); err == nil {
					// Remove existing .gitconfig if present
					os.Remove(gitConfigPath)
					// Create symlink to current profile
					err := os.Symlink(currentProfilePath, gitConfigPath)
					if err != nil {
						t.Fatalf("Failed to create symlink: %v", err)
					}
				}
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Capture stderr as well for log messages
			oldStderr := os.Stderr
			r2, w2, _ := os.Pipe()
			os.Stderr = w2

			// Execute the command
			listCmd.Run(listCmd, []string{})

			// Restore stdout/stderr
			w.Close()
			os.Stdout = oldStdout
			w2.Close()
			os.Stderr = oldStderr

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			var bufErr bytes.Buffer
			io.Copy(&bufErr, r2)
			errOutput := bufErr.String()

			// Combine both outputs for checking
			combinedOutput := output + errOutput

			// Check expected strings in output
			for _, expected := range tt.expectedInOutput {
				if !strings.Contains(combinedOutput, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, combinedOutput)
				}
			}

			// Additional checks for specific test cases
			if tt.currentProfile != "" {
				// Should contain the current profile marker
				expectedCurrentMarker := tt.currentProfile + " (current)"
				if !strings.Contains(combinedOutput, expectedCurrentMarker) {
					t.Errorf("Expected output to contain current profile marker %q, but got:\n%s", expectedCurrentMarker, combinedOutput)
				}
			}

			// For multiple profiles, check that all are listed
			if len(tt.setupProfiles) > 1 {
				for _, profile := range tt.setupProfiles {
					if profile != tt.currentProfile {
						// Non-current profiles should appear without marker
						if !strings.Contains(combinedOutput, "  "+profile) {
							t.Errorf("Expected output to contain profile %q without marker, but got:\n%s", profile, combinedOutput)
						}
					}
				}
			}
		})
	}
}

func TestListCmdEdgeCases(t *testing.T) {
	t.Run("corrupted gitconfig", func(t *testing.T) {
		// Reset homedir cache to ensure test isolation
		homedir.Reset()
		
		// Create temporary directories with unique suffix
		tempDir := t.TempDir()
		uniqueSuffix := fmt.Sprintf("_%d_corrupted", time.Now().UnixNano())
		tempDir = filepath.Join(tempDir, uniqueSuffix)
		err := os.MkdirAll(tempDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create unique temp directory: %v", err)
		}
		
		// Set environment variables to use temp directory FIRST
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		// Disable color output for tests
		originalNoColor := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")
		defer os.Setenv("NO_COLOR", originalNoColor)

		confPath := filepath.Join(tempDir, ".config", "gitconfigs")
		gitConfigPath := filepath.Join(tempDir, ".gitconfig")

		// Setup a profile
		err = os.MkdirAll(confPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		profilePath := filepath.Join(confPath, "work")
		utils.Write(profilePath, []byte("[user]\n\tname = work\n\temail = work@example.com"))

		// Create a corrupted .gitconfig (not a symlink to any profile)
		utils.Write(gitConfigPath, []byte("[user]\n\tname = other\n\temail = other@example.com"))

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		oldStderr := os.Stderr
		r2, w2, _ := os.Pipe()
		os.Stderr = w2

		// Execute the command
		listCmd.Run(listCmd, []string{})

		// Restore stdout/stderr
		w.Close()
		os.Stdout = oldStdout
		w2.Close()
		os.Stderr = oldStderr

		// Read captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		var bufErr bytes.Buffer
		io.Copy(&bufErr, r2)
		errOutput := bufErr.String()

		combinedOutput := output + errOutput

		// Should show profiles but indicate no active profile
		expectedStrings := []string{
			"Available git configuration profiles",
			"work",
			"Use 'git-switcher switch",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(combinedOutput, expected) {
				t.Errorf("Expected output to contain %q, but got:\n%s", expected, combinedOutput)
			}
		}
	})

	t.Run("broken symlink gitconfig", func(t *testing.T) {
		// Reset homedir cache to ensure test isolation
		homedir.Reset()
		
		// Create temporary directories with unique suffix
		tempDir := t.TempDir()
		uniqueSuffix := fmt.Sprintf("_%d_broken_symlink", time.Now().UnixNano())
		tempDir = filepath.Join(tempDir, uniqueSuffix)
		err := os.MkdirAll(tempDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create unique temp directory: %v", err)
		}
		
		// Set environment variables to use temp directory FIRST
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tempDir)
		defer os.Setenv("HOME", originalHome)

		// Disable color output for tests
		originalNoColor := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")
		defer os.Setenv("NO_COLOR", originalNoColor)

		confPath := filepath.Join(tempDir, ".config", "gitconfigs")
		gitConfigPath := filepath.Join(tempDir, ".gitconfig")

		// Setup a profile
		err = os.MkdirAll(confPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		profilePath := filepath.Join(confPath, "work")
		utils.Write(profilePath, []byte("[user]\n\tname = work\n\temail = work@example.com"))

		// Create a broken symlink
		brokenTarget := filepath.Join(confPath, "nonexistent")
		err = os.Symlink(brokenTarget, gitConfigPath)
		if err != nil {
			t.Fatalf("Failed to create broken symlink: %v", err)
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute the command (this should not panic)
		listCmd.Run(listCmd, []string{})

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		// Read captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Should still show profiles even with broken symlink
		if !strings.Contains(output, "work") {
			t.Errorf("Expected output to contain profile 'work', but got:\n%s", output)
		}
	})
} 