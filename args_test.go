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
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestArgs(t *testing.T) {
	t.Run("--version", func(t *testing.T) {
		var a Arguments
		test.Assert(t, a.Parse([]string{"bin", "-version"}) == nil)
		test.Assert(t, a.AskVersion)
	})
	t.Run("--version", func(t *testing.T) {
		var a Arguments
		test.Assert(t, a.Parse([]string{"bin", "--version"}) == nil)
		test.Assert(t, a.AskVersion)
	})
	t.Run("recurse", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--recurse"})
			test.Assert(t, err != nil, err)
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--r"})
			test.Assert(t, err != nil, err)
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--recurse", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Recursive)
			test.Assert(t, a.IDL == "idl-path")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--r", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Recursive)
			test.Assert(t, a.IDL == "idl-path")
		}
	})
	t.Run("verbose", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--verbose", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Verbose)
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--v", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Verbose)
		}
	})
	t.Run("verbose", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--quiet", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Quiet)
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--q", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Quiet)
		}
	})
	t.Run("out", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--out", "./out", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.OutputPath == "./out")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--o", "./out", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.OutputPath == "./out")
		}
	})
	t.Run("include", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--include", "a", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Includes.String() == "[a]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--include", "a", "--include", "b", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Includes.String() == "[a b]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--i", "a", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Includes.String() == "[a]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--i", "a", "--i", "b", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Includes.String() == "[a b]")
		}
	})
	t.Run("langs", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--gen", "a", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Langs.String() == "[a]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--gen", "a", "--gen", "b", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Langs.String() == "[a b]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--g", "a", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Langs.String() == "[a]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--g", "a", "--g", "b", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Langs.String() == "[a b]")
		}
	})
	t.Run("include", func(t *testing.T) {
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--plugin", "a", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Plugins.String() == "[a]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--plugin", "a", "--plugin", "b", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Plugins.String() == "[a b]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--p", "a", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Plugins.String() == "[a]")
		}
		{
			var a Arguments
			err := a.Parse([]string{"bin", "--p", "a", "--p", "b", "idl-path"})
			test.Assert(t, err == nil, err)
			test.Assert(t, a.Plugins.String() == "[a b]")
		}
	})
	t.Run("all", func(t *testing.T) {
		var a Arguments
		err := a.Parse([]string{"bin", "--recurse", "--g", "a", "--g", "b", "--out", "./out", "--include", "a", "--include", "b", "--verbose", "--plugin", "a", "--plugin", "b", "--quiet", "idl-path"})
		test.Assert(t, err == nil, err)
		test.Assert(t, a.Recursive)
		test.Assert(t, a.Verbose)
		test.Assert(t, a.IDL == "idl-path")
		test.Assert(t, a.Quiet)
		test.Assert(t, a.OutputPath == "./out")
		test.Assert(t, a.Includes.String() == "[a b]")
		test.Assert(t, a.Plugins.String() == "[a b]")
		test.Assert(t, a.Langs.String() == "[a b]")
	})
}
