package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	confPath, err := homedir.Expand("~/.config/gitconfigs")
	if err != nil {
		log.Panic(err)
	}

	log.SetFlags(log.Lshortfile)
	// hash: filename
	configs := make(map[string]string)

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		err = os.Mkdir(confPath, os.ModeDir|0700)
		if err != nil {
			log.Println(err)
		}
	}

	// List ~/.gitconfigs folder
	err = filepath.WalkDir(confPath+"/", func(path string, d fs.DirEntry, e error) error {
		if d.IsDir() {
			return nil
		}

		configs[hash(path)] = filepath.Base(path)

		return nil
	})
	if err != nil {
		log.Println(err)
	}

	// Check current gitconfig is exist in configs
	gitConfig, _ := homedir.Expand("~/.gitconfig")

	// If gitconfig file is not exist create empty file
	if _, err := os.Stat(gitConfig); os.IsNotExist(err) {
		f, e := os.Create(gitConfig)
		if e != nil {
			panic(e)
		}
		defer f.Close()
		f.Write([]byte(""))
	}
	gitConfigHash := hash(gitConfig)
	if _, ok := configs[gitConfigHash]; !ok {
		err := os.Link(gitConfig, confPath+"/Old Config")
		if err != nil {
			log.Panic(err)
		}
	}

	//	log.Println(configs)
	newConfig := ""
	if len(os.Args) > 1 && len(os.Args[1]) != 0 {
		newConfig = os.Args[1]
	} else if len(configs) >= 1 {
		// List git configs
		var profiles []string
		var pos int
		i := 0
		for hash, val := range configs {
			// Find current config index
			if hash == gitConfigHash {
				pos = i
			}
			profiles = append(profiles, val)
			i++
		}

		prompt := promptui.Select{
			Label: "Select Git Config (Current: " + configs[gitConfigHash] + ")",
			Items: profiles,
			// Change cursor to current config
			CursorPos:    pos,
			HideSelected: true,
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		newConfig = result
	}

	// Remove file for link new one
	err = os.Remove(gitConfig)
	if err != nil {
		log.Panic(err)
	}

	// Symbolic link to "~/.gitconfig"
	err = os.Symlink(confPath+"/"+newConfig, gitConfig)
	if err != nil {
		log.Panic(err)
	}
	color.HiBlue("Switched to profile %q", newConfig)
}

func hash(path string) string {
	f, _ := os.Open(path)
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}
