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
	"runtime"

	"github.com/cloudwego/gopkg/unsafex"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/utils/dir_utils"
)

// LangSpec is the parameter to specify which language to generate codes for and
// what plugins should be used.
type LangSpec struct {
	Language    string
	Options     []plugin.Option
	UsedPlugins []*plugin.Desc
	SDKPlugins  []plugin.SDKPlugin
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

	if len(out.SDKPlugins) > 0 {
		for _, sdk := range out.SDKPlugins {
			req.PluginParameters = sdk.GetPluginParameters()
			extra := sdk.Invoke(req)
			if err := extra.GetError(); err != "" {
				return plugin.BuildErrorResponse(err)
			}
			if err := g.files.Feed(sdk.GetName(), extra.Contents); err != nil {
				return plugin.BuildErrorResponse(err.Error())
			}
		}
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
	p := newAsyncPostProcess(g.pp)
	for i, c := range res.Contents {
		full := c.GetName()
		if full == "" {
			return fmt.Errorf("file name not found for the %dth generated item", i)
		}
		if !filepath.IsAbs(full) && dir_utils.HasGlobalWd() {
			wd, err := dir_utils.Getwd()
			if err != nil {
				return err
			}
			full = filepath.Join(wd, full)
		}
		p.Add(full, c.Content)
	}
	return p.OnFinished(func(path string, content []byte) error {
		g.log.Info("Write", path)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create path '%s': %w", dir, err)
		}
		if err := ioutil.WriteFile(path, content, 0o644); err != nil {
			return fmt.Errorf("failed to write file '%s': %w", path, err)
		}
		return nil
	})
}

type asyncPostProcessJob struct {
	Path    string
	Content string
}

type asyncPostProcess struct {
	pp         backend.PostProcessor
	jobs       []asyncPostProcessJob
	rets       chan error
	processing chan struct{}

	concurrency int
}

func newAsyncPostProcess(pp backend.PostProcessor) *asyncPostProcess {
	return &asyncPostProcess{
		pp:          pp,
		concurrency: runtime.GOMAXPROCS(0),
	}
}

func (p *asyncPostProcess) Add(path, content string) {
	p.jobs = append(p.jobs, asyncPostProcessJob{Path: path, Content: content})
}

func (p *asyncPostProcess) wait() error {
	for len(p.processing) > 0 {
		if err := <-p.rets; err != nil {
			return err
		}
	}
	return nil
}

func (p *asyncPostProcess) OnFinished(f func(path string, content []byte) error) error {
	p.rets = make(chan error, len(p.jobs))
	if p.concurrency == 0 {
		p.concurrency = 1
	}
	p.processing = make(chan struct{}, p.concurrency)
	for i := 0; i < len(p.jobs); {
		select {
		case err := <-p.rets:
			if err != nil {
				_ = p.wait()
				return err
			}
			continue // retry

		case p.processing <- struct{}{}: // processing++
			j := p.jobs[i]
			go func(path string, content []byte) {
				defer func() { <-p.processing }() // processing--
				var err error
				if p.pp != nil {
					content, err = p.pp.PostProcess(path, content)
				}
				if err == nil {
					err = f(path, content)
				}
				p.rets <- err
			}(j.Path, unsafex.StringToBinary(j.Content))
		}
		i++ // next job
	}
	return p.wait()
}
