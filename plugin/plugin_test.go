// Copyright 2022 CloudWeGo Authors
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

package plugin

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/cloudwego/thriftgo/parser"
)

func TestReadPluginThriftGoVersion(t *testing.T) {
	// Test with current executable
	self, err := exec.LookPath(os.Args[0])
	if err != nil {
		t.Fatalf("Failed to find current executable: %v", err)
	}

	// Since we're testing with the current executable, we should get a valid version
	// or an empty string if the binary wasn't built with module information
	version := readPluginThriftGoVersion(self)

	// We can't assert the exact version, but we can check the format or emptiness
	if version != "" {
		// If we got a version, it should follow semver format or be a pseudo-version
		// Just do a basic check that it's not completely invalid
		if !strings.HasPrefix(version, "v") && !strings.Contains(version, "-") {
			t.Errorf("Unexpected version format: %s", version)
		}
	}

	// Test with a non-existent file
	nonExistentVersion := readPluginThriftGoVersion("non-existent-file")
	if nonExistentVersion != "" {
		t.Errorf("Expected empty version for non-existent file, got: %s", nonExistentVersion)
	}
}

func TestAppendDataTrailer(t *testing.T) {
	data := make([]byte, 100)
	if hasDataTrailerFeature(data, featureCompressInclude) {
		t.Fatal("should false")
	}
	data = appendDataTrailer(data, featureCompressInclude)
	if !hasDataTrailerFeature(data, featureCompressInclude) {
		t.Fatal("should true")
	}
}

func TestSupportDataTrailer(t *testing.T) {
	// Test positive cases
	for _, v := range []string{"v0.4.2-abc", "v0.4.2", "v0.10.0", "v1.0.0"} {
		if !supportDataTrailer(v) {
			t.Errorf("expected true for version %s", v)
		}
	}
	// Test negative cases
	for _, v := range []string{"v0.1.7", "v0.4.1", "", "0.4.2", "v0.4", "v0.4.1-abc", "invalid", "v1.2.3.4"} {
		if supportDataTrailer(v) {
			t.Errorf("expected false for versions %s", v)
		}
	}
}

func TestCompressThriftInclude(t *testing.T) {
	z := &parser.Thrift{Filename: "z.thrift"}
	y := &parser.Thrift{Filename: "y.thrift"}
	y.Includes = []*parser.Include{{Path: "z.thrift", Reference: z}}

	x := &parser.Thrift{Filename: "x.thrift"}
	x.Includes = []*parser.Include{{Path: "z.thrift", Reference: z}, {Path: "y.thrift", Reference: y}}

	compressThriftInclude(x, nil)

	// Reference z.thrift should be compressed
	if fn := y.Includes[0].Reference.Filename; fn != refFilenamePrefix+"z.thrift" {
		t.Fatal(fn)
	}
	if x.Includes[0].Reference == y.Includes[0].Reference {
		t.Fatal("must not same")
	}

	decompressThriftInclude(x, nil)
	if fn := y.Includes[0].Reference.Filename; fn != "z.thrift" {
		t.Fatal(fn)
	}
	if x.Includes[0].Reference != y.Includes[0].Reference {
		t.Fatal("must same")
	}
}
