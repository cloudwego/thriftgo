// Copyright 2024 CloudWeGo Authors
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

package golang_test

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/thriftgo/internal/test_util"
	"github.com/cloudwego/thriftgo/plugin"
)

const testSlimIDL = `
namespace go test_slim

struct User {
    1: required string Name,
    2: optional i32 Age,
    3: list<string> Tags,
    4: map<string, string> Attributes,
}

struct Request {
    1: required User User,
    2: optional string Token,
}
`

func TestSlimTemplateFieldMask(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repoRoot, err := filepath.Abs(filepath.Join(wd, "../.."))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                 string
		opts                 []plugin.Option
		expectMethods        bool
		expectHalfwayMethods bool
	}{
		{
			name: "slim template with field_mask",
			opts: []plugin.Option{
				{Name: "template", Desc: "slim"},
				{Name: "with_field_mask"},
				{Name: "with_reflection"},
			},
			expectMethods:        true,
			expectHalfwayMethods: false,
		},
		{
			name: "slim template with field_mask and halfway",
			opts: []plugin.Option{
				{Name: "template", Desc: "slim"},
				{Name: "with_field_mask"},
				{Name: "with_reflection"},
				{Name: "field_mask_halfway"},
			},
			expectMethods:        true,
			expectHalfwayMethods: true,
		},
		{
			name: "no_default_serdes with field_mask",
			opts: []plugin.Option{
				{Name: "no_default_serdes"},
				{Name: "with_field_mask"},
				{Name: "with_reflection"},
			},
			expectMethods:        true,
			expectHalfwayMethods: false,
		},
		{
			name: "no_default_serdes with field_mask and halfway",
			opts: []plugin.Option{
				{Name: "no_default_serdes"},
				{Name: "with_field_mask"},
				{Name: "with_reflection"},
				{Name: "field_mask_halfway"},
			},
			expectMethods:        true,
			expectHalfwayMethods: true,
		},
		{
			name: "raw_struct template with field_mask",
			opts: []plugin.Option{
				{Name: "template", Desc: "raw_struct"},
				{Name: "with_field_mask"},
				{Name: "with_reflection"},
			},
			expectMethods:        false,
			expectHalfwayMethods: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			idlPath := filepath.Join(tmpDir, "test.thrift")
			if err := os.WriteFile(idlPath, []byte(testSlimIDL), 0o644); err != nil {
				t.Fatal(err)
			}

			outDir := filepath.Join(tmpDir, "gen")
			gen, res := test_util.GenerateGolang(idlPath, outDir, tc.opts, nil)
			if errStr := res.GetError(); errStr != "" {
				t.Fatalf("generation failed: %s", errStr)
			}
			if err := gen.Persist(res); err != nil {
				t.Fatalf("persist failed: %v", err)
			}

			var foundStructFile bool
			err = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
					return err
				}
				if strings.HasSuffix(path, "-reflection.go") {
					return nil
				}

				contentBytes, readErr := os.ReadFile(path)
				if readErr != nil {
					return readErr
				}
				content := string(contentBytes)

				if strings.Contains(content, "type User struct") {
					foundStructFile = true

					if !strings.Contains(content, "_fieldmask *fieldmask.FieldMask") {
						t.Errorf("file %s: expected struct definition to contain '_fieldmask *fieldmask.FieldMask', got:\n%s", path, content)
					}

					if !strings.Contains(content, "github.com/cloudwego/thriftgo/fieldmask") {
						t.Errorf("file %s: expected import of fieldmask package", path)
					}

					if tc.expectMethods {
						if !strings.Contains(content, "func (p *User) Get_FieldMask() *fieldmask.FieldMask") {
							t.Errorf("file %s: missing Get_FieldMask method", path)
						}
						if !strings.Contains(content, "func (p *User) Set_FieldMask(fm *fieldmask.FieldMask)") {
							t.Errorf("file %s: missing Set_FieldMask method", path)
						}
					}
					if tc.expectHalfwayMethods {
						if !strings.Contains(content, "func (p *User) Pass_FieldMask(fm *fieldmask.FieldMask)") {
							t.Errorf("file %s: missing Pass_FieldMask method", path)
						}
					}
				}

				fset := token.NewFileSet()
				if _, parseErr := parser.ParseFile(fset, path, contentBytes, parser.AllErrors); parseErr != nil {
					t.Errorf("file %s failed to parse: %v", path, parseErr)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if !foundStructFile {
				t.Fatalf("did not find struct file in %s", outDir)
			}

			goModContent := "module test_gen\n\ngo 1.20\n\nreplace github.com/cloudwego/thriftgo => " + repoRoot + "\nreplace github.com/apache/thrift => github.com/apache/thrift v0.13.0\n\nrequire (\n\tgithub.com/apache/thrift v0.13.0\n\tgithub.com/cloudwego/thriftgo v0.0.0-00010101000000-000000000000\n)\n\nrequire (\n\tgithub.com/bytedance/gopkg v0.1.4 // indirect\n\tgithub.com/cloudwego/gopkg v0.2.0 // indirect\n\tgithub.com/davecgh/go-spew v1.1.1 // indirect\n\tgithub.com/pmezard/go-difflib v1.0.0 // indirect\n\tgopkg.in/yaml.v3 v3.0.1 // indirect\n)\n"
			if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o644); err != nil {
				t.Fatal(err)
			}

			goSumSrc, err := os.ReadFile(filepath.Join(repoRoot, "tests/fieldmask/go.sum"))
			if err == nil {
				_ = os.WriteFile(filepath.Join(tmpDir, "go.sum"), goSumSrc, 0o644)
			}

			cmd := exec.Command("go", "vet", "./...")
			cmd.Dir = tmpDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go vet failed: %v\nOutput:\n%s", err, string(out))
			}
		})
	}
}
