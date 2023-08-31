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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cloudwego/thriftgo/version"

	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/plugin"
)

// StringSlice implements the flag.Value interface on string slices
// to allow a flag to be set multiple times.
type StringSlice []string

func (ss *StringSlice) String() string {
	return fmt.Sprintf("%v", *ss)
}

// Set implements the flag.Value interface.
func (ss *StringSlice) Set(value string) error {
	*ss = append(*ss, value)
	return nil
}

// Arguments contains command line arguments for thriftgo.
type Arguments struct {
	AskVersion      bool
	Recursive       bool
	Verbose         bool
	Quiet           bool
	CheckKeyword    bool
	OutputPath      string
	Includes        StringSlice
	Plugins         StringSlice
	Langs           StringSlice
	IDL             string
	PluginTimeLimit time.Duration
}

// Output returns an output path for generated codes for the target language.
func (a *Arguments) Output(lang string) string {
	if len(a.OutputPath) > 0 {
		return a.OutputPath
	}
	return "./gen-" + lang
}

const WINDOWS_REPLACER = "#$$#"

// UsedPlugins returns a list of plugin.Desc for plugins.
func (a *Arguments) UsedPlugins() (descs []*plugin.Desc, err error) {
	for _, str := range a.Plugins {
		if runtime.GOOS == "windows" {
			// windows should replace :\ because thriftgo will separates args by ":"
			str = strings.ReplaceAll(str, ":\\", WINDOWS_REPLACER)
		}
		desc, err := plugin.ParseCompactArguments(str)
		if err != nil {
			return nil, err
		}
		if runtime.GOOS == "windows" {
			desc.Name = strings.ReplaceAll(desc.Name, WINDOWS_REPLACER, ":\\")
			for i, o := range desc.Options {
				desc.Options[i].Name = strings.ReplaceAll(o.Name, WINDOWS_REPLACER, ":\\")
				desc.Options[i].Desc = strings.ReplaceAll(o.Desc, WINDOWS_REPLACER, ":\\")
			}
		}
		descs = append(descs, desc)
	}
	return
}

// Targets returns a list of generator.LangSpec for target languages.
func (a *Arguments) Targets() (specs []*generator.LangSpec, err error) {
	for _, lang := range a.Langs {
		desc, err := plugin.ParseCompactArguments(lang)
		if err != nil {
			return nil, err
		}
		opts, err := a.checkOptions(desc.Options)
		if err != nil {
			return nil, err
		}
		desc.Options = opts
		spec := &generator.LangSpec{
			Language: desc.Name,
			Options:  desc.Options,
		}
		specs = append(specs, spec)
	}
	return
}

func (a *Arguments) checkOptions(opts []plugin.Option) ([]plugin.Option, error) {
	params := plugin.Pack(opts)
	cu := golang.NewCodeUtils(backend.DummyLogFunc())
	cu.HandleOptions(params)
	if cu.Features().EnableNestedStruct {
		if cu.Template() != "slim" {
			found := false
			for _, opt := range opts {
				if opt.Name == "template" {
					log.Printf("[WARN] EnableNestedStruct is only available under the \"slim\" template, so adapt the template to \"slim\"")
					opt.Desc = "slim"
					found = true
					break
				}
			}
			if !found {
				log.Printf("[WARN] EnableNestedStruct is only available under the \"slim\" template, so adapt the template to \"slim\"")
				opts = append(opts, plugin.Option{Name: "template", Desc: "slim"})
			}

		}
	}
	return opts, nil
}

// MakeLogFunc creates logging functions according to command line flags.
func (a *Arguments) MakeLogFunc() backend.LogFunc {
	logs := backend.DummyLogFunc()

	if !a.Quiet {
		if a.Verbose {
			logger := log.New(os.Stderr, "[INFO] ", 0)
			logs.Info = func(v ...interface{}) {
				logger.Println(v...)
			}
		}

		logger := log.New(os.Stderr, "[WARN] ", 0)
		logs.Warn = func(v ...interface{}) {
			logger.Println(v...)
		}
		logs.MultiWarn = func(ws []string) {
			for _, w := range ws {
				logger.Println(w)
			}
		}
	}

	return logs
}

// BuildFlags initializes command line flags.
func (a *Arguments) BuildFlags() *flag.FlagSet {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	f.BoolVar(&a.AskVersion, "version", false, "")

	f.BoolVar(&a.Recursive, "r", false, "")
	f.BoolVar(&a.Recursive, "recurse", false, "")

	f.BoolVar(&a.Verbose, "v", false, "")
	f.BoolVar(&a.Verbose, "verbose", false, "")

	f.BoolVar(&a.Quiet, "q", false, "")
	f.BoolVar(&a.Quiet, "quiet", false, "")

	f.StringVar(&a.OutputPath, "o", "", "")
	f.StringVar(&a.OutputPath, "out", "", "")

	f.Var(&a.Includes, "i", "")
	f.Var(&a.Includes, "include", "")

	f.Var(&a.Langs, "g", "")
	f.Var(&a.Langs, "gen", "")

	f.Var(&a.Plugins, "p", "")
	f.Var(&a.Plugins, "plugin", "")

	f.BoolVar(&a.CheckKeyword, "check-keywords", true, "")

	f.DurationVar(&a.PluginTimeLimit, "plugin-time-limit", time.Minute, "")

	f.Usage = help
	return f
}

// Parse parse command line arguments.
func (a *Arguments) Parse(argv []string) error {
	f := a.BuildFlags()
	if err := f.Parse(argv[1:]); err != nil {
		return err
	}

	if a.AskVersion {
		return nil
	}

	rest := f.Args()
	if len(rest) != 1 {
		return fmt.Errorf("require exactly 1 argument for the IDL parameter, got: %d", len(rest))
	}

	a.IDL = rest[0]
	return nil
}

func help() {
	println("Version:", version.ThriftgoVersion)
	println(`Usage: thriftgo [options] file
Options:
  --version           Print the compiler version and exit.
  -h, --help          Print help message and exit.
  -i, --include dir   Add a search path for includes.
  -o, --out dir	      Set the output location for generated files. Default path is ./gen-*, the code will be genereated at ./gen-*/xxxnamespace.
					  If you don't want the path ends with namespace, you can use {namespace} or {namespaceUnderscore}, such as /gen-*/{namespace}/data
  -r, --recurse       Generate codes for includes recursively.
  -v, --verbose       Output detail logs.
  -q, --quiet         Suppress all warnings and informatic logs.
  -g, --gen STR       Specify the target language.
                      STR has the form language[:key1=val1[,key2[,key3=val3]]].
                      Keys and values are options passed to the backend.
                      Many options will not require values. Boolean options accept
                      "false", "true" and "" (empty is treated as "true").
  -p, --plugin STR    Specify an external plugin to invoke.
                      STR has the form plugin[=path][:key1=val1[,key2[,key3=val3]]].
  --check-keywords    Check if any identifier using a keyword in common languages. 
  --plugin-time-limit Set the execution time limit for plugins. Naturally 0 means no limit.

Available generators (and options):
`)
	// print backend options
	for _, b := range g.AllBackend() {
		name, lang := b.Name(), b.Lang()
		println(fmt.Sprintf("  %s (%s):", name, lang))
		println(align(b.Options()))
	}
	println()
	os.Exit(2)
}

// align the help strings for plugin options.
func align(opts []plugin.Option) string {
	var names, descs, ss []string
	max := 0
	for _, opt := range opts {
		names = append(names, opt.Name)
		descs = append(descs, opt.Desc)
		if max <= len(opt.Name) {
			max = len(opt.Name)
		}
	}

	for i := range names {
		rest := 2 + max - len(names[i])
		ss = append(ss, fmt.Sprintf(
			"    %s:%s%s",
			names[i],
			strings.Repeat(" ", rest),
			descs[i],
		))
	}
	return strings.Join(ss, "\n")
}
