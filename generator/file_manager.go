// Copyright 2021 CloudWeGo Authors
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

FileLoop:
	for i := 0; i < len(files); i++ {
		f := files[i]
		if !f.IsSetName() {
			if last == "" {
				return fmt.Errorf("[%s] attended to append but no target file found", src)
			}
			fm.patch[last] = append(fm.patch[last], f)
			continue
		}
		name := f.GetName()
		if idx, ok := fm.index[name]; !ok {
			fm.index[name] = len(fm.files)
			fm.files = append(fm.files, f)
		} else {
			if f.GetInsertionPoint() != "" {
				// FIXME: when the target file is renamed due to name collision, the patch may be invalid.
				fm.patch[name] = append(fm.patch[name], f)
			} else {
				fst := idx
				ext := filepath.Ext(name)
				pth := strings.TrimSuffix(name, ext)
				cnt := 1

				var renamed string
				for {
					if fm.files[idx].Content == f.Content { // duplicate content
						fm.log.Info(fmt.Sprintf("[%s] discard generated file '%s': size %d", src, name, len(f.Content)))
						for j := i + 1; j < len(files) && !files[j].IsSetName(); j++ {
							fm.log.Info("discard patch @", files[j].GetInsertionPoint())
							i++
						}
						continue FileLoop
					}
					renamed = fmt.Sprintf("%s_%d%s", pth, cnt, ext)
					if cnt > fm.count[name] {
						break
					} else {
						idx = fm.index[renamed]
						cnt++
					}
				}

				fm.log.Warn(fmt.Sprintf("[%s] file names conflict: '%s' (%d <> %d)", src, name, len(fm.files[fst].Content), len(f.Content)))
				fm.index[renamed] = len(fm.files)
				fm.files = append(fm.files, f)
				fm.count[name]++
				f.Name = &renamed
				name = renamed // propagate the new name to last
			}
		}
		last = name
	}
	return nil
}

type insertionPointReplacer struct {
	m map[string]string
}

var insertReg = regexp.MustCompile(fmt.Sprintf(plugin.InsertionPointFormat, `\([$.0-9a-zA-Z_]*\)`))

func newInsertionPointReplacer(content string) *insertionPointReplacer {
	kk := insertReg.FindAllString(content, -1) // all insertion points
	m := make(map[string]string, len(kk))
	for _, k := range kk {
		m[k] = ""
	}
	return &insertionPointReplacer{m}
}

func (p *insertionPointReplacer) Add(x, content string) {
	v := p.m[x]
	if v == "" {
		p.m[x] = content
	} else {
		// allow multiple content for a single insertion point
		p.m[x] += content
	}
}

func (p *insertionPointReplacer) Replace(content string) string {
	oldnews := make([]string, 0, 2*len(p.m))
	for k, v := range p.m {
		oldnews = append(oldnews, k, v)
	}
	return strings.NewReplacer(oldnews...).Replace(content)
}

// BuildResponse creates a plugin.Response containing all files that the
// FileManager manages.  All insertion points will be removed after the response
// is built.
func (fm *FileManager) BuildResponse() *plugin.Response {
	res := plugin.NewResponse()
	for _, f := range fm.files {
		x := newInsertionPointReplacer(f.Content)
		for _, p := range fm.patch[f.GetName()] {
			x.Add(plugin.InsertionPoint(p.GetInsertionPoint()), p.Content)
		}
		g := &plugin.Generated{
			Name:    f.Name,
			Content: x.Replace(f.Content),
		}
		res.Contents = append(res.Contents, g)
	}
	return res
}
