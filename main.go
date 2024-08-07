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
	"fmt"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/cloudwego/thriftgo/args"
	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/fastgo"
	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/sdk"
)

var (
	a args.Arguments
	g generator.Generator
)

var debugMode bool

func init() {
	_ = g.RegisterBackend(new(golang.GoBackend))
	_ = g.RegisterBackend(new(fastgo.FastGoBackend))
	// export THRIFTGO_DEBUG=1
	debugMode = os.Getenv("THRIFTGO_DEBUG") == "1"
}

func check(err error) {
	if err != nil {
		if err.Error() != "flag: help requested" {
			println(err.Error())
		}
		os.Exit(2)
	}
}

func main() {
	if debugMode {
		f, _ := os.Create("thriftgo-cpu.pprof")
		defer f.Close()
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		startTime := time.Now()
		defer func() {
			fmt.Printf("Cost: %s\n", time.Since(startTime))
		}()
	}

	defer handlePanic()

	check(sdk.InvokeThriftgo(nil, os.Args...))
}

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Println("Recovered from panic:")
		fmt.Println(r)
		debug.PrintStack()
	}
}
