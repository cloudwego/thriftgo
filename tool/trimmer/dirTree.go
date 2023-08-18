package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// create directory-tree before dump
func createDirTree(sourceDir string, destinationDir string) {
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			newDir := filepath.Join(destinationDir, path[len(sourceDir):])
			err := os.MkdirAll(newDir, os.ModePerm)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("manage output error: %v\n", err)
		os.Exit(2)
	}
}

// remove empty directory of output dir-tree
func removeEmptyDir(source string) {
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			empty, err := isDirectoryEmpty(path)
			if err != nil {
				return err
			}
			if empty {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func isDirectoryEmpty(path string) (bool, error) {
	dir, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	_, err = dir.Readdirnames(1)
	if err == nil {
		return false, nil
	}

	if len(err.Error()) > len("EOF") {
		return false, err
	}
	return true, nil
}
