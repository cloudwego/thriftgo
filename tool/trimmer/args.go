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
	"os"
	"strings"
	"time"

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
	println("Version:", Version)
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
