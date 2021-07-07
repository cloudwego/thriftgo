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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/plugin"
)

// LangSpec is the parameter to specify which language to generate codes for and
// what plugins should be used.
type LangSpec struct {
	Language    string
	Options     []plugin.Option
	UsedPlugins []*plugin.Desc
}

// Arguments contains arguments for generator's Generate method.
type Arguments struct {
	Out *LangSpec
	Req *plugin.Request
	Log backend.LogFunc
}

// Generator controls the code generation.
// The zero value of Generator is ready for use.
type Generator struct {
	backends []backend.Backend
	plugins  []plugin.Plugin
	files    *FileManager
	log      backend.LogFunc
	pp       backend.PostProcessor
}

// Name returns "thriftgo".
func (g *Generator) Name() string {
	return "thriftgo"
}

// RegisterBackend adds a backend to the generator.
func (g *Generator) RegisterBackend(b backend.Backend) error {
	if l := b.Lang(); g.GetBackend(l) != nil {
		return fmt.Errorf("backend for language '%s' already exist", l)
	}
	g.backends = append(g.backends, b)
	return nil
}

// GetBackend returns a backend of the language.
func (g *Generator) GetBackend(name string) backend.Backend {
	for _, x := range g.backends {
		if x.Name() == name {
			return x
		}
	}
	return nil
}

// AllBackend returns all registered backends.
func (g *Generator) AllBackend() []backend.Backend {
	return g.backends
}

func (g *Generator) validateRequest(req *plugin.Request) error {
	// TODO(lushaojie): validate request
	return nil
}

func (g *Generator) preparePlugins(be backend.Backend, pds []*plugin.Desc) error {
	for _, d := range pds {
		// TODO(lushaojie): check d

		if p := be.GetPlugin(d); p != nil {
			g.plugins = append(g.plugins, p)
			continue
		}
		p, err := plugin.Lookup(d.Name)
		if err != nil {
			return err
		}

		g.plugins = append(g.plugins, p)
	}
	return nil
}

// Generate generates codes for the target language and executes plugins specified.
func (g *Generator) Generate(args *Arguments) (res *plugin.Response) {
	out, req, log := args.Out, args.Req, args.Log

	log.Info(fmt.Sprintf(`Generating: "%s"`, out.Language))
	if err := g.validateRequest(req); err != nil {
		return plugin.BuildErrorResponse(err.Error())
	}

	g.files = NewFileManager(log)
	g.log = log

	be := g.GetBackend(out.Language)
	if be == nil {
		err := fmt.Sprintf("No generator for language '%s'.", out.Language)
		return plugin.BuildErrorResponse(err)
	}
	if pp, ok := be.(backend.PostProcessor); ok {
		g.pp = pp
	}

	if err := g.preparePlugins(be, out.UsedPlugins); err != nil {
		return plugin.BuildErrorResponse(err.Error())
	}

	req.GeneratorParameters = plugin.Pack(out.Options)
	res = be.Generate(req, log)
	log.MultiWarn(res.Warnings)
	if res.GetError() != "" {
		return res
	}
	log.Info("Got", len(res.Contents), "contents")

	if err := g.files.Feed(g.Name(), res.Contents); err != nil {
		return plugin.BuildErrorResponse(err.Error())
	}

	for i, p := range g.plugins {
		log.Info(fmt.Sprintf(`Run plugin "%s"`, p.Name()))

		req.PluginParameters = plugin.Pack(out.UsedPlugins[i].Options)
		extra := p.Execute(req)
		log.MultiWarn(extra.Warnings)

		if err := extra.GetError(); err != "" {
			return plugin.BuildErrorResponse(err)
		}
		if err := g.files.Feed(p.Name(), extra.Contents); err != nil {
			return plugin.BuildErrorResponse(err.Error())
		}
	}

	res = g.files.BuildResponse()
	return res
}

// Persist writes generated files into the disk. Each files in the Contents
// slice must have a legal name.
func (g *Generator) Persist(res *plugin.Response) error {
	if err := res.GetError(); err != "" {
		return errors.New(err)
	}
	for i, c := range res.Contents {
		full := c.GetName()
		if full == "" {
			return fmt.Errorf("file name not found for the %dth generated item", i)
		}

		content := []byte(c.Content)
		if g.pp != nil {
			processed, err := g.pp.PostProcess(full, content)
			if err != nil {
				return err
			}
			content = processed
		}

		g.log.Info("Write", full)
		path := filepath.Dir(full)
		if err := os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create path '%s': %w", path, err)
		}
		if err := ioutil.WriteFile(full, content, 0644); err != nil {
			return fmt.Errorf("failed to write file '%s': %w", full, err)
		}
	}
	return nil
}
