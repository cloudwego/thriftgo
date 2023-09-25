package main

import (
	"github.com/cloudwego/thriftgo/pkg/test"
	"os"
	"path/filepath"
	"testing"
)

func TestDirTree(t *testing.T) {
	_ = os.RemoveAll("trimmer_test")
	createDirTree("test_cases", "trimmer_test")
	fileCount, dirCount, err := countFilesAndSubdirectories("trimmer_test")
	test.Assert(t, err == nil)
	test.Assert(t, fileCount == 0)
	test.Assert(t, dirCount == 3)
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
