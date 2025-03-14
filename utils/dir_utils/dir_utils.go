// Copyright 2024 CloudWeGo Authors
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

func ToAbsolute(k string) (string, error) {
	if !filepath.IsAbs(k) {
		wd, err := Getwd()
		if err != nil {
			return "", err
		}
		k = filepath.Join(wd, k)
		absK, err := filepath.Abs(k)
		if err != nil {
			return "", err
		}
		k = absK
	}
	return k, nil
}

// ToRelative 将绝对路径转换为相对路径
func ToRelative(path string) (string, error) {
	if filepath.IsAbs(path) {
		wd, err := Getwd()
		if err != nil {
			return "", err
		}
		relPath, err := filepath.Rel(wd, path)
		if err != nil {
			return "", err
		}
		path = relPath
	}
	return path, nil
}
