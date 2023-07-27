// Copyright 2023 CloudWeGo Authors
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
	"os"

	"github.com/cloudwego/thriftgo/generator"
)

// Version of trimmer tool.
const Version = "0.0.1"

var (
	a Arguments
	g generator.Generator
)

func check(err error) {
	if err != nil {
		println(err.Error())
		os.Exit(2)
	}
}

func main() {
	// you can execute "go install" to install this tool and execute "trimmer" or "trimmer -version"
	// todo finish your own arg parser
	println("IDL TRIMMER.....")
	check(a.Parse(os.Args))
	if a.AskVersion {
		println("thriftgo trimmer tool ", Version)
		os.Exit(0)
	}

	println("todo.....")
	// read file and trim and do output
	os.Exit(0)
}
