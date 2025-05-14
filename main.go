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
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"time"

	"github.com/cloudwego/thriftgo/sdk"
)

var debugMode bool

func init() {
	// export THRIFTGO_DEBUG=1
	debugMode = os.Getenv("THRIFTGO_DEBUG") == "1"
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if debugMode {

		f1, err := os.Create("thriftgo-cpu.pprof")
		assert(err)
		assert(pprof.StartCPUProfile(f1))

		f2, err := os.Create("thriftgo-heap.pprof")
		assert(err)

		startTime := time.Now()
		defer func() {
			fmt.Printf("Cost: %s\n", time.Since(startTime))

			pprof.StopCPUProfile()
			assert(pprof.WriteHeapProfile(f2))
			f1.Close()
			f2.Close()
		}()
	}

	defer handlePanic()

	if err := sdk.InvokeThriftgo(nil, os.Args...); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			println(err.Error())
		}
		os.Exit(2)
	}
}

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Println("Recovered from panic:")
		fmt.Println(r)
		debug.PrintStack()
	}
}
