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

package generator_test

import (
	"testing"

	"github.com/cloudwego/thriftgo/generator"
	"github.com/cloudwego/thriftgo/generator/backend"
	"github.com/cloudwego/thriftgo/pkg/test"
	"github.com/cloudwego/thriftgo/plugin"
)

func pstr(s string) *string {
	return &s
}

func TestFileManagerEmpty(t *testing.T) {
	fm := generator.NewFileManager(backend.DummyLogFunc())

	resp := fm.BuildResponse()
	test.Assert(t, resp != nil)
	test.Assert(t, !resp.IsSetError())
	test.Assert(t, len(resp.Contents) == 0)
	test.Assert(t, len(resp.Warnings) == 0)
}

func TestFileManagerInsert(t *testing.T) {
	fm := generator.NewFileManager(backend.DummyLogFunc())

	pos2 := plugin.InsertionPoint("2nd")
	pos3 := plugin.InsertionPoint("3rd")
	fs := []*plugin.Generated{
		{
			Content: "first file",
			Name:    pstr("first"),
		},
		{
			Content: "second file begin\n" + pos2 + "\nsecond file end",
			Name:    pstr("second"),
		},
		{
			Content: "third file\n" + pos3,
			Name:    pstr("third"),
		},
		{
			Content:        "patch to third",
			InsertionPoint: pstr("3rd"),
		},
		{
			Content:        "patch to second",
			Name:           pstr("second"),
			InsertionPoint: pstr("2nd"),
		},
	}
	err := fm.Feed("test", fs)
	test.Assert(t, err == nil)

	resp := fm.BuildResponse()
	test.Assert(t, resp != nil)
	test.Assert(t, !resp.IsSetError())
	test.Assert(t, len(resp.Contents) == 3)
	test.Assert(t, len(resp.Warnings) == 0)
	test.Assert(t, resp.Contents[0].GetName() == "first")
	test.Assert(t, resp.Contents[1].GetName() == "second")
	test.Assert(t, resp.Contents[2].GetName() == "third")
	for i := range resp.Contents {
		test.Assert(t, !resp.Contents[i].IsSetInsertionPoint())
	}
	test.Assert(t, resp.Contents[0].Content == "first file")
	test.Assert(t, resp.Contents[1].Content == "second file begin\npatch to second\nsecond file end")
	test.Assert(t, resp.Contents[2].Content == "third file\npatch to third")
}

func TestFileManagerRename(t *testing.T) {
	fm := generator.NewFileManager(backend.DummyLogFunc())

	fs := []*plugin.Generated{
		{
			Content: "first file",
			Name:    pstr("first"),
		},
		{
			Content: "second file",
			Name:    pstr("second"),
		},
		{
			Content: "another second file",
			Name:    pstr("second"),
		},
		{
			Content: "another second file",
			Name:    pstr("second"),
		},
		{
			Content: "third file",
			Name:    pstr("third"),
		},
	}
	err := fm.Feed("test", fs)
	test.Assert(t, err == nil)

	resp := fm.BuildResponse()
	test.Assert(t, resp != nil)
	test.Assert(t, !resp.IsSetError())
	test.Assert(t, len(resp.Contents) == 4)
	test.Assert(t, len(resp.Warnings) == 0)
	test.Assert(t, resp.Contents[0].GetName() == "first")
	test.Assert(t, resp.Contents[1].GetName() == "second")
	test.Assert(t, resp.Contents[2].GetName() == "second_1")
	test.Assert(t, resp.Contents[3].GetName() == "third")
	for i := range resp.Contents {
		test.Assert(t, !resp.Contents[i].IsSetInsertionPoint())
	}
	test.Assert(t, resp.Contents[0].Content == "first file")
	test.Assert(t, resp.Contents[1].Content == "second file")
	test.Assert(t, resp.Contents[2].Content == "another second file")
	test.Assert(t, resp.Contents[3].Content == "third file")
}
