// Copyright 2021 CloudWeGo
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

package generator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/plugin"
)

// FileManager manages in-memory files that used during the code generation process.
type FileManager struct {
	files []*plugin.Generated
	patch map[string][]*plugin.Generated
	index map[string]int
	count map[string]int
	log   backend.LogFunc
}

// NewFileManager creates a new FileManager.
func NewFileManager(log backend.LogFunc) *FileManager {
	return &FileManager{
		patch: make(map[string][]*plugin.Generated),
		index: make(map[string]int),
		count: make(map[string]int),
		log:   log,
	}
}

// Feed adds files to the FileManager.
func (fm *FileManager) Feed(src string, files []*plugin.Generated) error {
	var last string

	for _, f := range files {
		if f.IsSetName() {
			name := f.GetName()
			if idx, ok := fm.index[name]; ok {
				if f.GetInsertionPoint() != "" {
					fm.patch[name] = append(fm.patch[name], f)
				} else {
					a, b := len(fm.files[idx].Content), len(f.Content)
					if a == b {
						fm.log.Warn(fmt.Sprintf("[%s] replaces generated file '%s': %d -> %d", src, name, a, b))
						fm.files[idx] = f
					} else {
						fm.log.Warn(fmt.Sprintf("[%s] file names conflict: '%s' (%d <> %d)", src, name, a, b))
						fm.count[name]++

						ext := filepath.Ext(name)
						pth := strings.TrimSuffix(name, ext)
						name2 := fmt.Sprintf("%s_%d%s", pth, fm.count[name], ext)
						f.Name = &name2

						fm.index[name2] = len(fm.files)
						fm.files = append(fm.files, f)
					}
				}
			} else {
				fm.index[name] = len(fm.files)
				fm.files = append(fm.files, f)
			}

			last = name
		} else {
			if last == "" {
				return fmt.Errorf("[%s] attended to append but no target file found", src)
			}
			fm.patch[last] = append(fm.patch[last], f)
			continue
		}
	}
	return nil
}

var (
	ptn = fmt.Sprintf(plugin.InsertionPointFormat, `\([.0-9a-zA-Z_]*\)`)
	reg = regexp.MustCompile(ptn)
)

func stripInsertionPoint(content string) string {
	return reg.ReplaceAllLiteralString(content, "")
}

// BuildResponse creates a plugin.Response containing all files that the
// FileManager manages.  All insertion points will be removed after the response
// is built.
func (fm *FileManager) BuildResponse() *plugin.Response {
	res := plugin.NewResponse()
	for _, f := range fm.files {
		content := f.Content
		for _, p := range fm.patch[f.GetName()] {
			pos := fmt.Sprintf(plugin.InsertionPointFormat, p.GetInsertionPoint())
			txt := p.Content + pos
			content = strings.Replace(content, pos, txt, -1)
		}

		g := &plugin.Generated{
			Name:    f.Name,
			Content: stripInsertionPoint(content),
		}
		res.Contents = append(res.Contents, g)
	}
	return res
}
