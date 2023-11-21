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

package golang

import (
	"fmt"
	"go/format"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cloudwego/thriftgo/generator/golang/streaming"
	"github.com/cloudwego/thriftgo/tool/trimmer/trim"

	ref_tpl "github.com/cloudwego/thriftgo/generator/golang/templates/ref"
	reflection_tpl "github.com/cloudwego/thriftgo/generator/golang/templates/reflection"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang/templates"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
)

// GoBackend generates go codes.
// The zero value of GoBackend is ready for use.
type GoBackend struct {
	err              error
	tpl              *template.Template
	refTpl           *template.Template
	reflectionTpl    *template.Template
	reflectionRefTpl *template.Template
	req              *plugin.Request
	res              *plugin.Response
	log              backend.LogFunc

	utils *CodeUtils
	funcs template.FuncMap
}

// Name implements the Backend interface.
func (g *GoBackend) Name() string {
	return "go"
}

// Lang implements the Backend interface.
func (g *GoBackend) Lang() string {
	return "Go"
}

// Options implements the Backend interface.
func (g *GoBackend) Options() (opts []plugin.Option) {
	for _, p := range allParams {
		opts = append(opts, plugin.Option{
			Name: p.name,
			Desc: p.desc,
		})
	}
	return opts
}

// BuiltinPlugins implements the Backend interface.
func (g *GoBackend) BuiltinPlugins() []*plugin.Desc {
	return nil
}

// GetPlugin implements the Backend interface.
func (g *GoBackend) GetPlugin(desc *plugin.Desc) plugin.Plugin {
	return nil
}

// Generate implements the Backend interface.
func (g *GoBackend) Generate(req *plugin.Request, log backend.LogFunc) *plugin.Response {
	g.req = req
	g.res = plugin.NewResponse()
	g.log = log
	g.prepareUtilities()
	if g.utils.Features().TrimIDL {
		g.log.Warn("You Are Using IDL Trimmer")
		structureTrimmed, fieldTrimmed, err := trim.TrimAST(&trim.TrimASTArg{Ast: req.AST, TrimMethods: nil, Preserve: nil})
		if err != nil {
			g.log.Warn("trim error:", err.Error())
		}
		g.log.Warn(fmt.Sprintf("removed %d unused structures with %d fields", structureTrimmed, fieldTrimmed))
	}
	g.prepareTemplates()
	g.fillRequisitions()
	if !g.utils.Features().ThriftStreaming {
		g.removeStreamingFunctions(req.GetAST())
	}
	g.executeTemplates()
	return g.buildResponse()
}

func (g *GoBackend) prepareUtilities() {
	if g.err != nil {
		return
	}

	g.utils = NewCodeUtils(g.log)
	g.err = g.utils.HandleOptions(g.req.GeneratorParameters)
	if g.err != nil {
		return
	}

	g.funcs = g.utils.BuildFuncMap()
	g.funcs["Version"] = func() string { return g.req.Version }
}

func (g *GoBackend) prepareTemplates() {
	if g.err != nil {
		return
	}

	all := template.New("thrift").Funcs(g.funcs)
	tpls := templates.Templates()

	if name := g.utils.Template(); name != defaultTemplate {
		tpls = g.utils.alternative[name]
	}
	for _, tpl := range tpls {
		all = template.Must(all.Parse(tpl))
	}
	g.tpl = all

	g.refTpl = template.Must(template.New("thrift-ref").Funcs(g.funcs).Parse(ref_tpl.File))
	g.reflectionTpl = template.Must(template.New("thrift-reflection").Funcs(g.funcs).Parse(reflection_tpl.File))
	g.reflectionRefTpl = template.Must(template.New("thrift-reflection-util").Funcs(g.funcs).Parse(reflection_tpl.FileRef))
}

func (g *GoBackend) fillRequisitions() {
	if g.err != nil {
		return
	}
}

func (g *GoBackend) executeTemplates() {
	if g.err != nil {
		return
	}

	processed := make(map[*parser.Thrift]bool)

	var trees chan *parser.Thrift
	if g.req.Recursive {
		trees = g.req.AST.DepthFirstSearch()
	} else {
		trees = make(chan *parser.Thrift, 1)
		trees <- g.req.AST
		close(trees)
	}

	for ast := range trees {
		if processed[ast] {
			continue
		}
		processed[ast] = true
		g.log.Info("Processing", ast.Filename)

		if g.err = g.renderOneFile(ast); g.err != nil {
			break
		}
	}
}

func (g *GoBackend) renderOneFile(ast *parser.Thrift) error {
	keepName := g.utils.Features().KeepCodeRefName
	path := g.utils.CombineOutputPath(g.req.OutputPath, ast)
	filename := filepath.Join(path, g.utils.GetFilename(ast))
	localScope, refScope, err := BuildRefScope(g.utils, ast)
	if err != nil {
		return err
	}
	err = g.renderByTemplate(localScope, g.tpl, filename)
	if err != nil {
		return err
	}
	err = g.renderByTemplate(refScope, g.refTpl, ToRefFilename(keepName, filename))
	if err != nil {
		return err
	}
	if g.utils.Features().WithReflection {
		err = g.renderByTemplate(refScope, g.reflectionRefTpl, ToReflectionRefFilename(keepName, filename))
		if err != nil {
			return err
		}
		return g.renderByTemplate(localScope, g.reflectionTpl, ToReflectionFilename(filename))
	}
	return nil
}

func ToRefFilename(keepName bool, filename string) string {
	if keepName {
		return filename
	}
	return strings.TrimSuffix(filename, ".go") + "-ref.go"
}

func ToReflectionFilename(filename string) string {
	return strings.TrimSuffix(filename, ".go") + "-reflection.go"
}

func ToReflectionRefFilename(keepName bool, filename string) string {
	if keepName {
		return ToReflectionFilename(filename)
	}
	return strings.TrimSuffix(filename, ".go") + "-reflection-ref.go"
}

func (g *GoBackend) renderByTemplate(scope *Scope, executeTpl *template.Template, filename string) error {
	if scope == nil {
		return nil
	}
	var buf strings.Builder
	g.utils.SetRootScope(scope)
	err := executeTpl.ExecuteTemplate(&buf, executeTpl.Name(), scope)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	g.res.Contents = append(g.res.Contents, &plugin.Generated{
		Content: buf.String(),
		Name:    &filename,
	})
	buf.Reset()
	imports, err := scope.ResolveImports()
	if err != nil {
		return err
	}
	err = executeTpl.ExecuteTemplate(&buf, "Imports", imports)
	if err != nil {
		return fmt.Errorf("%s: %w", filename, err)
	}
	point := "imports"
	g.res.Contents = append(g.res.Contents, &plugin.Generated{
		Content:        buf.String(),
		InsertionPoint: &point,
	})
	return nil
}

func (g *GoBackend) buildResponse() *plugin.Response {
	if g.err != nil {
		return plugin.BuildErrorResponse(g.err.Error())
	}
	return g.res
}

// PostProcess implements the backend.PostProcessor interface to do
// source formatting before writing files out.
func (g *GoBackend) PostProcess(path string, content []byte) ([]byte, error) {
	switch filepath.Ext(path) {
	case ".go":
		if formated, err := format.Source(content); err != nil {
			g.log.Warn(fmt.Sprintf("Failed to format %s: %s", path, err.Error()))
		} else {
			content = formated
		}
	}
	return content, nil
}

func (g *GoBackend) removeStreamingFunctions(ast *parser.Thrift) {
	for _, svc := range ast.Services {
		functions := make([]*parser.Function, 0, len(svc.Functions))
		for _, f := range svc.Functions {
			st, err := streaming.ParseStreaming(f)
			if err != nil {
				g.log.Warn(fmt.Sprintf("%s.%s: failed to parse streaming, err = %v", svc.Name, f.Name, err))
				continue
			}
			if st.IsStreaming {
				g.log.Warn(fmt.Sprintf("skip streaming function %s.%s: not supported by your kitex, "+
					"please update your kitex tool to the latest version", svc.Name, f.Name))
				continue
			}
			functions = append(functions, f)
		}
		svc.Functions = functions
	}
}
