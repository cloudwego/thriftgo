// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// create directory-tree before dump
func createDirTree(sourceDir, destinationDir string) {
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path[len(sourceDir)-1] != filepath.Separator {
				path = path + string(filepath.Separator)
			}
			newDir := filepath.Join(destinationDir, path[len(sourceDir):])
			err := os.MkdirAll(newDir, os.ModePerm)
			if err != nil {
				return errors.New("create dir tree error:" + err.Error())
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
	files, err := os.ReadDir(source)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			removeEmptyDir(source + string(filepath.Separator) + file.Name())
		}
	}
	empty, err := isDirectoryEmpty(source)
	if empty || err != nil {
		_ = os.RemoveAll(source)
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

	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	return false, err
}
