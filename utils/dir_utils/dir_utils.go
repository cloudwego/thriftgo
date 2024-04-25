package dir_utils

import (
	"os"
	"path/filepath"
)

var globalwd string

func SetGlobalwd(wd string) {
	globalwd = wd
}

func HasGlobalWd() bool {
	return globalwd != ""
}

func Getwd() (string, error) {
	if globalwd == "" {
		return os.Getwd()
	}
	currentwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	wantedwd := globalwd
	return filepath.Rel(currentwd, wantedwd)
}
