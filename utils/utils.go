package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
)

// Hash generates an MD5 hash of the given file's content.
// It's used to identify git configuration files.
func Hash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		// If the file doesn't exist, we can't hash it.
		// Depending on desired behavior, either panic or return an error/empty string.
		// For this utility, if a path is given, it's expected to be hashable.
		// If it could be a new/empty .gitconfig, os.IsNotExist(err) could be checked.
		log.Panicf("Failed to open file %s for hashing: %v", path, err)
	}
	defer f.Close() // Ensure file is closed
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalf("Failed to copy file content to hasher for %s: %v", path, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Write creates or truncates a file and writes data to it.
func Write(file string, data []byte) {
	f, e := os.Create(file)
	if e != nil {
		log.Panicf("Failed to create file %s: %v", file, e)
	}
	defer f.Close() // Ensure file is closed
	_, err := f.Write(data)
	if err != nil {
		log.Panicf("Failed to write to file %s: %v", file, err)
	}
}
