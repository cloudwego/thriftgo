/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fastgo

import (
	"bytes"
	"fmt"
	"go/format"
	"path"
	"path/filepath"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
)

// FastGoBackend ...
type FastGoBackend struct {
	golang.GoBackend

	req *plugin.Request
	log backend.LogFunc

	utils *golang.CodeUtils
}

var _ backend.Backend = &FastGoBackend{}

// Name implements the Backend interface.
func (g *FastGoBackend) Name() string { return "fastgo" }

// Lang implements the Backend interface.
func (g *FastGoBackend) Lang() string { return "FastGo" }

// Generate implements the Backend interface.
func (g *FastGoBackend) Generate(req *plugin.Request, log backend.LogFunc) *plugin.Response {
	ret := g.GoBackend.Generate(req, log)
	if ret.Error != nil {
		return ret
	}
	g.req = req
	g.log = log
	g.utils = g.GoBackend.GetCoreUtils()
	var trees chan *parser.Thrift
	if req.Recursive {
		trees = req.AST.DepthFirstSearch()
	} else {
		trees = make(chan *parser.Thrift, 1)
		trees <- req.AST
		close(trees)
	}
	respErr := func(err error) *plugin.Response {
		errstr := err.Error()
		ret.Error = &errstr
		return ret
	}
	processed := make(map[*parser.Thrift]bool)
	for ast := range trees {
		if processed[ast] {
			continue
		}
		processed[ast] = true
		log.Info("Processing", ast.Filename)
		content, err := g.GenerateOne(ast)
		if err != nil {
			return respErr(err)
		}
		ret.Contents = append(ret.Contents, content)
	}
	return ret
}

func (g *FastGoBackend) GenerateOne(ast *parser.Thrift) (*plugin.Generated, error) {
	// the filename should differentiate the default code files,
	// keep same as kitex, coz we're deprecating the old impl of fastcodec.
	// it will overwrites the old k-xxx.go.
	filename := "k-" + g.utils.GetFilename(ast)
	filename = filepath.Join(g.utils.CombineOutputPath(g.req.OutputPath, ast), filename)

	// not generating ref code, see `code_ref` parameter
	scope, _, err := golang.BuildRefScope(g.utils, ast)
	if err != nil {
		return nil, fmt.Errorf("golang.BuildRefScope: %w", err)
	}

	w := newCodewriter()

	// TODO: only supports struct now, other dirty jobs will be done in golang.GoBackend
	for _, s := range scope.Structs() {
		g.generateStruct(w, scope, s)
	}
	for _, s := range scope.Unions() {
		g.generateStruct(w, scope, s)
	}
	for _, s := range scope.Exceptions() {
		g.generateStruct(w, scope, s)
	}
	for _, ss := range scope.Services() {
		for _, f := range ss.Functions() {
			if s := f.ArgType(); s != nil {
				g.generateStruct(w, scope, s)
			}
			if s := f.ResType(); s != nil {
				g.generateStruct(w, scope, s)
			}
		}
	}

	ret := &plugin.Generated{}
	ret.Name = &filename

	// for ret.Content
	c := &bytes.Buffer{}

	// Headers:
	// thriftgo version and package name
	packageName := path.Base(golang.GetImportPath(g.utils, ast))
	fmt.Fprintf(c, "%s\npackage %s\n\n", fixedFileHeader, packageName)

	// Imports
	unusedProtect := false
	for _, incl := range scope.Includes() {
		if incl == nil { // TODO(liyun.339): fix this
			continue
		}
		unusedProtect = true
		w.UsePkg(incl.ImportPath, incl.PackageName)
	}
	if len(w.pkgs) > 0 {
		c.WriteString(w.Imports())
	}
	c.WriteByte('\n')

	// Unused protects
	if unusedProtect {
		fmt.Fprintln(c, "var (")
		for _, incl := range scope.Includes() {
			if incl == nil { // TODO(liyun.339): fix this
				continue
			}
			fmt.Fprintf(c, "_ = %s.KitexUnusedProtection\n", incl.PackageName)
		}
		fmt.Fprintln(c, ")")
	}

	// Methods
	c.Write(w.Bytes())

	ret.Content = g.Format(filename, c.Bytes())
	return ret, nil
}

func (g *FastGoBackend) Format(filename string, content []byte) string {
	if g.utils.Features().NoFmt {
		return string(content)
	}
	if formated, err := format.Source(content); err != nil {
		g.log.Warnf("Failed to format %s: %s", filename, err.Error())
	} else {
		content = formated
	}
	return string(content)
}

func (g *FastGoBackend) generateStruct(w *codewriter, scope *golang.Scope, s *golang.StructLike) {
	// TODO: This method doesn't generate struct definition for now.
	// It only generates a better version of FastRead, FastWrite(Nocopy) methods which originally from Kitex.
	g.genBLength(w, scope, s)
	g.genFastWrite(w, scope, s)
	g.genFastRead(w, scope, s)
}
