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

	"github.com/cloudwego/thriftgo/plugin"
	"github.com/cloudwego/thriftgo/version"
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
	AskVersion bool
	OutputFile string
	IDL        string
	Recurse    string
	Methods    StringSlice
	Force      bool
}

// BuildFlags initializes command line flags.
func (a *Arguments) BuildFlags() *flag.FlagSet {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	f.BoolVar(&a.AskVersion, "version", false, "")

	f.StringVar(&a.OutputFile, "o", "", "")
	f.StringVar(&a.OutputFile, "out", "", "")

	f.StringVar(&a.Recurse, "r", "", "")
	f.StringVar(&a.Recurse, "recurse", "", "")

	f.Var(&a.Methods, "m", "")
	f.Var(&a.Methods, "method", "")

	f.BoolVar(&a.Force, "f", false, "")
	f.BoolVar(&a.Force, "force", false, "")

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
	println(`Usage: trimmer [options] file
Options:
  --version			Print the compiler version and exit.
  -h, --help			Print help message and exit.
  -o, --out	[file/dir]	Specify the output IDL file/dir.
  -r, --recurse	[dir]		Specify a root dir and dump the included IDL recursively beneath the given root. -o should be set as a directory.
  -m, --method [service.method] Only keep the specified methods and their dependents. Accept multiple -m.
  -f, --force	use force trimming, ignore @preserve comments
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
