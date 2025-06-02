package cmd

import (
	"bytes"
	// "fmt" // Not directly used in test functions
	// "io" // Not directly used in test functions
	"os"
	"path/filepath"
	"strings"
	"testing"
	// "github.com/theykk/git-switcher/cmd" // Removed to break import cycle
	// "github.com/theykk/git-switcher/utils" // utils.Hash is used by the main code, not directly in test funcs
)

// setupTestEnvironment creates a temporary home directory with a .config/gitconfigs subdir.
// It returns the path to the temp home, config dir, and gitconfig path, plus a cleanup function.
func setupTestEnvironment(t *testing.T) (string, string, string, func()) {
	t.Helper()

	tempHome, err := os.MkdirTemp("", "test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}

	configDir := filepath.Join(tempHome, ".config", "gitconfigs")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		os.RemoveAll(tempHome)
		t.Fatalf("Failed to create temp config dir: %v", err)
	}

	gitConfigPath := filepath.Join(tempHome, ".gitconfig")

	// originalHome := os.Getenv("HOME") // No longer using os.Setenv("HOME") for this
	// os.Setenv("HOME", tempHome)

	originalGetHomeDirFnc := getHomeDirFnc // Store original homedir func from cmd/list.go
	getHomeDirFnc = func() (string, error) { // Mock it
		return tempHome, nil
	}

	cleanup := func() {
		// os.Setenv("HOME", originalHome) // Restore original HOME if it was set
		getHomeDirFnc = originalGetHomeDirFnc // Restore original homedir func
		os.RemoveAll(tempHome)
	}

	// Return tempHome as well, though it's now primarily for creating files within it.
	return tempHome, configDir, gitConfigPath, cleanup
}

func createDummyProfile(t *testing.T, dir, name, content string) string {
	t.Helper()
	profilePath := filepath.Join(dir, name)
	err := os.WriteFile(profilePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy profile %s: %v", name, err)
	}
	return profilePath
}

func TestListCmdOutput(t *testing.T) { // Renamed TestListCmd to TestListCmdOutput for clarity
	// listCmd is an unexported variable in the same 'cmd' package (from list.go)
	// rootCmd is also unexported (from root.go)
	// Ensure listCmd is added to rootCmd for proper execution context if needed,
	// though Execute() on listCmd itself should work for simple cases.
	// The init() function in root.go should handle rootCmd.AddCommand(listCmd).
	// We just need to ensure that init() has run, which it should have by the time tests are run.

	t.Run("Basic Listing - All profiles shown and none active", func(t *testing.T) {
		_, configDir, gitConfigPath, cleanup := setupTestEnvironment(t)
		defer cleanup()

		createDummyProfile(t, configDir, "profileA", "[user]\n  name = UserA\n  email = usera@example.com")
		createDummyProfile(t, configDir, "profileB", "[user]\n  name = UserB\n  email = userb@example.com")
		createDummyProfile(t, configDir, "profileC", "[user]\n  name = UserC\n  email = userc@example.com")

		// Ensure .gitconfig does not exist or is empty for this test case, so no profile is "active"
		// Or, make its content not match any profile
		_ = os.WriteFile(gitConfigPath, []byte("[user]\n name = NonExistent"), 0644)


		output := new(bytes.Buffer)
		// listCmd.SetOut(output) // Output should be set on rootCmd
		// listCmd.SetErr(output) // Error output should be set on rootCmd
		rootCmd.SetOut(output)
		rootCmd.SetErr(output)

		rootCmd.SetArgs([]string{"list"})
		err := rootCmd.Execute()
		if err != nil {
			// If HOME override didn't work, homedir.Dir() might return actual home,
			// leading to errors or unexpected behavior.
			t.Fatalf("listCmd.Execute() failed: %v. Output: %s", err, output.String())
		}

		result := output.String()
		t.Logf("Captured output for 'Basic Listing':\nSTART_OUTPUT\n%s\nEND_OUTPUT", result) // Debugging line

		// Exact string matching including leading spaces and newline
		if !strings.Contains(result, "  profileA\n") {
			t.Errorf("Output does not contain '  profileA\\n'. Got:\n%s", result)
		}
		if !strings.Contains(result, "  profileB\n") {
			t.Errorf("Output does not contain '  profileB\\n'. Got:\n%s", result)
		}
		if !strings.Contains(result, "  profileC\n") {
			t.Errorf("Output does not contain '  profileC\\n'. Got:\n%s", result)
		}
		if strings.Contains(result, "*") { // Still check for asterisk for non-active
			t.Errorf("Output contains '*' indicating active profile, but none should be active. Got:\n%s", result)
		}
	})

	t.Run("Current Profile Indication", func(t *testing.T) {
		_, configDir, gitConfigPath, cleanup := setupTestEnvironment(t) // tempHomeUsed changed to _
		defer cleanup()

		createDummyProfile(t, configDir, "profileX", "[user]\n  name = UserX\n  email = userx@example.com")
		activeProfileContent := "[user]\n  name = UserY\n  email = usery@example.com"
		createDummyProfile(t, configDir, "profileY", activeProfileContent)
		createDummyProfile(t, configDir, "profileZ", "[user]\n  name = UserZ\n  email = userz@example.com")

		// Set .gitconfig to match profileY
		err := os.WriteFile(gitConfigPath, []byte(activeProfileContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write .gitconfig: %v", err)
		}


		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetErr(output)

		rootCmd.SetArgs([]string{"list"})
		executeErr := rootCmd.Execute()
		if executeErr != nil {
			t.Fatalf("listCmd.Execute() failed: %v. Output: %s", executeErr, output.String())
		}
		result := output.String()

		expectedActive := "* profileY (current)\n" // Added newline
		if !strings.Contains(result, expectedActive) {
			t.Errorf("Output does not correctly indicate active profile 'profileY'. Expected to contain '%s'. Got:\n%s", expectedActive, result)
		}
		// Check for other profiles not being active
		if strings.Contains(result, "* profileX") {
			t.Errorf("Output incorrectly marks profileX as active. Got:\n%s", result)
		}
		if strings.Contains(result, "* profileZ") {
			t.Errorf("Output incorrectly marks profileZ as active. Got:\n%s", result)
		}
		// Count check is still good
		if strings.Count(result, "*") > 1 { // Ensure only one asterisk
			t.Errorf("Output indicates more than one active profile. Got:\n%s", result)
		}
		if strings.Contains(result, "* profileX") || strings.Contains(result, "* profileZ") {
			t.Errorf("Output incorrectly marks profileX or profileZ as active. Got:\n%s", result)
		}
	})

	t.Run("No Profiles", func(t *testing.T) {
		// _, _, _, cleanup := setupTestEnvironment(t) // First call removed
		// defer cleanup() // First defer removed

		_, _, _, cleanup := setupTestEnvironment(t) // configDir is created but left empty. This is the only call needed.
		defer cleanup()

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetErr(output)

		rootCmd.SetArgs([]string{"list"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("listCmd.Execute() failed: %v. Output: %s", err, output.String())
		}
		result := output.String()

		// Message from list.go: fmt.Println("  No profiles found in " + filepath.Join("~", configSubPath) + ".")
		expectedMsg := "  No profiles found in " + filepath.Join("~", ".config", "gitconfigs") + "."
		if !strings.Contains(result, expectedMsg) {
			t.Errorf("Output does not indicate 'No profiles found'. Expected to contain\n'%s'. Got:\n'%s'", expectedMsg, result)
		}
	})

	t.Run(".gitconfig Not Matching Any Profile", func(t *testing.T) {
		_, configDir, gitConfigPath, cleanup := setupTestEnvironment(t)
		defer cleanup()

		createDummyProfile(t, configDir, "profileOne", "[user]\n  name = UserOne\n  email = userone@example.com")
		createDummyProfile(t, configDir, "profileTwo", "[user]\n  name = UserTwo\n  email = usertwo@example.com")

		err := os.WriteFile(gitConfigPath, []byte("[user]\n  name = UnknownUser\n  email = unknown@example.com"), 0644)
		if err != nil {
			t.Fatalf("Failed to write .gitconfig: %v", err)
		}

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetErr(output)

		rootCmd.SetArgs([]string{"list"})
		executeErr := rootCmd.Execute()
		if executeErr != nil {
			t.Fatalf("listCmd.Execute() failed: %v. Output: %s", executeErr, output.String())
		}
		result := output.String()

		if strings.Contains(result, "*") { // No active profile expected
			t.Errorf("Output indicates an active profile with '*', but none should be. Got:\n%s", result)
		}
		if !strings.Contains(result, "  profileOne\n") {
			t.Errorf("Output does not list '  profileOne\\n'. Got:\n%s", result)
		}
		if !strings.Contains(result, "  profileTwo\n") {
			t.Errorf("Output does not list '  profileTwo\\n'. Got:\n%s", result)
		}
	})

	t.Run("Configuration directory does not exist", func(t *testing.T) {
		tempHome, _, _, cleanup := setupTestEnvironment(t)
		defer cleanup()

		// Remove the .config/gitconfigs directory that setupTestEnvironment creates
		err := os.RemoveAll(filepath.Join(tempHome, ".config"))
		if err != nil {
			t.Fatalf("Failed to remove .config directory for test: %v", err)
		}

		output := new(bytes.Buffer)
		rootCmd.SetOut(output)
		rootCmd.SetErr(output)

		rootCmd.SetArgs([]string{"list"})
		executeErr := rootCmd.Execute() // Should not panic, should print message
		if executeErr != nil {
			// Depending on cobra's behavior for RunE errors, Execute might return an error.
			// For simple fmt.Println in Run, it might not.
			// If the command calls log.Fatal, the test will exit here.
			// The current listCmd.Run prints and returns, so Execute() shouldn't error out here.
			t.Logf("listCmd.Execute() returned error (may be expected for some error messages): %v", executeErr)
		}
		result := output.String()

		// Message from list.go: fmt.Println("Configuration directory " + filepath.Join("~", configSubPath) + " not found. Please create it first.")
		expectedMsg := "Configuration directory " + filepath.Join("~", ".config", "gitconfigs") + " not found. Please create it first."
		if !strings.Contains(result, expectedMsg) {
			t.Errorf("Output does not indicate config directory not found. Expected\n'%s'. Got:\n'%s'", expectedMsg, result)
		}
	})
}
