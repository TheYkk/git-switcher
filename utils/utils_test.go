package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:    "simple content",
			content: "[user]\n\tname = test\n\temail = test@example.com",
		},
		{
			name:    "empty content",
			content: "",
		},
		{
			name:    "special characters",
			content: "[user]\n\tname = test-user_123\n\temail = test+tag@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "test.txt")

			// Write content to file
			Write(tempFile, []byte(tt.content))

			// Calculate hash
			result := Hash(tempFile)

			// Calculate expected hash manually
			h := md5.New()
			h.Write([]byte(tt.content))
			expected := hex.EncodeToString(h.Sum(nil))

			if result != expected {
				t.Errorf("Hash() = %v, want %v", result, expected)
			}

			// Verify hash is consistent
			result2 := Hash(tempFile)
			if result != result2 {
				t.Errorf("Hash() is not consistent: %v != %v", result, result2)
			}
		})
	}
}

func TestHashNonExistentFile(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Hash() should panic when file doesn't exist")
		}
	}()

	Hash("/nonexistent/file/path")
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{
			name:    "simple text",
			content: []byte("hello world"),
		},
		{
			name:    "git config",
			content: []byte("[user]\n\tname = test\n\temail = test@example.com"),
		},
		{
			name:    "empty content",
			content: []byte(""),
		},
		{
			name:    "binary content",
			content: []byte{0, 1, 2, 3, 255, 254, 253},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "test.txt")

			// Write content
			Write(tempFile, tt.content)

			// Read back and verify
			result, err := os.ReadFile(tempFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			if string(result) != string(tt.content) {
				t.Errorf("Write() content mismatch, got %v, want %v", result, tt.content)
			}
		})
	}
}

func TestWriteOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")

	// Write initial content
	Write(tempFile, []byte("initial content"))

	// Verify initial content
	result, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(result) != "initial content" {
		t.Errorf("Initial write failed, got %v, want %v", string(result), "initial content")
	}

	// Overwrite with new content
	Write(tempFile, []byte("new content"))

	// Verify new content
	result, err = os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read file after overwrite: %v", err)
	}
	if string(result) != "new content" {
		t.Errorf("Overwrite failed, got %v, want %v", string(result), "new content")
	}
}

func TestWriteInvalidPath(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Write() should panic when path is invalid")
		}
	}()

	// Try to write to an invalid path (directory that doesn't exist)
	Write("/nonexistent/directory/file.txt", []byte("content"))
} 