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
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestDirTree(t *testing.T) {
	_ = os.RemoveAll("trimmer_test")
	createDirTree("test_cases", "trimmer_test")
	fileCount, dirCount, err := countFilesAndSubdirectories("trimmer_test")
	test.Assert(t, err == nil)
	test.Assert(t, fileCount == 0)
	test.Assert(t, dirCount == 4)
	removeEmptyDir("trimmer_test")
	_, err = os.ReadDir("trimmer_test")
	test.Assert(t, err != nil)
}

func countFilesAndSubdirectories(dirPath string) (int, int, error) {
	var fileCount, dirCount int
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, 0, err
	}
	for _, file := range files {
		if file.IsDir() {
			dirCount++
			subDirPath := filepath.Join(dirPath, file.Name())
			subFileCount, subDirCount, err := countFilesAndSubdirectories(subDirPath)
			if err != nil {
				return 0, 0, err
			}
			fileCount += subFileCount
			dirCount += subDirCount
		} else {
			fileCount++
		}
	}
	return fileCount, dirCount, nil
}
