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

package plugin

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/thriftgo/parser"
)

// MaxExecutionTime is a timeout for executing external plugins.
var MaxExecutionTime time.Duration

// InsertionPointFormat is the format for insertion points.
const InsertionPointFormat = "@@thriftgo_insertion_point(%s)"

// InsertionPoint returns a new insertion point.
func InsertionPoint(names ...string) string {
	return fmt.Sprintf(InsertionPointFormat, strings.Join(names, "."))
}

// Option is used to describes an option for a plugin or a generator backend.
type Option struct {
	Name string
	Desc string
}

// Desc can be used to describes the interface of a plugin or a generator backend.
type Desc struct {
	Name    string
	Options []Option
}

// Pack packs Option into strings.
func Pack(opts []Option) (ss []string) {
	for _, o := range opts {
		ss = append(ss, o.Name+"="+o.Desc)
	}
	return
}

// ParseCompactArguments parses a compact form option into arguments.
// A compact form option is like:
//
//	name:key1=val1,key2,key3=val3
//
// This function barely checks the validity of the string, so the user should
// always provide a valid input.
func ParseCompactArguments(str string) (*Desc, error) {
	if str == "" {
		return nil, errors.New("ParseArguments: empty string")
	}
	desc := new(Desc)
	args := strings.SplitN(str, ":", 2)
	desc.Name = args[0]
	if len(args) > 1 {
		for _, a := range strings.Split(args[1], ",") {
			var opt Option
			kv := strings.SplitN(a, "=", 2)
			switch len(kv) {
			case 2:
				opt.Desc = kv[1]
				fallthrough
			case 1:
				opt.Name = kv[0]
				desc.Options = append(desc.Options, opt)
			}
		}
	}
	return desc, nil
}

// BuildErrorResponse creates a plugin response with a error message.
func BuildErrorResponse(errMsg string, warnings ...string) *Response {
	return &Response{
		Error:    &errMsg,
		Warnings: warnings,
	}
}

// Plugin takes a request and builds a response to generate codes.
type Plugin interface {
	// Name returns the name of the plugin.
	Name() string

	// Execute invokes the plugin and waits for response.
	Execute(req *Request) (res *Response)
}

// Lookup searches for PATH to find a plugin that match the description.
func Lookup(arg string) (Plugin, error) {
	parts := strings.SplitN(arg, "=", 2)

	var name, full string
	switch len(parts) {
	case 0:
		return nil, fmt.Errorf("invalid plugin name: %s", arg)
	case 1:
		name, full = arg, "thrift-gen-"+arg
	case 2:
		name, full = parts[0], parts[1]
	}

	path, err := exec.LookPath(full)
	if err != nil {
		return nil, err
	}
	return &external{name: name, full: full, path: path}, nil
}

type external struct {
	name string
	full string
	path string
}

// Name implements the Plugin interface.
func (e *external) Name() string {
	return e.name
}

// Execute implements the Plugin interface.
func (e *external) Execute(req *Request) (res *Response) {
	trailer := supportDataTrailer(readPluginThriftGoVersion(e.path))
	if trailer && enableCompressThriftInclude {
		m := map[string]*parser.Thrift{}
		compressThriftInclude(req.AST, m)
		defer decompressThriftInclude(req.AST, m) // revert
	}
	data, err := MarshalRequest(req)
	if err != nil {
		err = fmt.Errorf("failed to marshal request: %w", err)
		return BuildErrorResponse(err.Error())
	}
	if trailer && enableCompressThriftInclude {
		data = appendDataTrailer(data, featureCompressInclude)
	}

	ctx := context.Background()

	if MaxExecutionTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, MaxExecutionTime)
		defer cancel()
	}

	stdin := bytes.NewReader(data)
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, e.path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, &stdout, &stderr

	if err = cmd.Run(); err != nil {
		warns := []string{
			"stdout:\n" + stdout.String(),
			"stderr:\n" + stderr.String(),
		}
		err = fmt.Errorf("execute plugin '%s' failed: %w", e.Name(), err)
		return BuildErrorResponse(err.Error(), warns...)
	}

	res, err = UnmarshalResponse(stdout.Bytes())
	if err != nil {
		err = fmt.Errorf(
			"failed to unmarshal plugin response: %w\nstdout:\n%s\nstderr:\n%s",
			err, stdout.String(), stderr.String())
		return BuildErrorResponse(err.Error())
	}
	if warn := stderr.String(); len(warn) > 0 {
		res.Warnings = append(res.Warnings, e.Name()+" stderr:\n"+warn)
	}
	return res
}

type SDKPlugin interface {
	Invoke(req *Request) (res *Response)
	GetName() string
	GetPluginParameters() []string
}

const thriftgoPackage = "github.com/cloudwego/thriftgo"

func readPluginThriftGoVersion(name string) string {
	bi, err := buildinfo.ReadFile(name)
	if err != nil {
		return ""
	}
	for _, d := range bi.Deps {
		if d.Path == thriftgoPackage {
			return d.Version
		}
	}
	return ""
}

const pluginDataTrailer = "\xffTHRIFTGO_TRAILER_V1\xff"

const (
	featureCompressInclude uint8 = 1 << 0
)

// NOTE: remove the env once the feature is considered to be stable.
// In the feature, we should remove the whole plugin mode instead of using this
var enableCompressThriftInclude = (os.Getenv("THRIFTGO_PLUGIN_COMPRESS_INCLUDE") == "1")

func appendDataTrailer(data []byte, feature uint8) []byte {
	data = append(data, feature)
	return append(data, pluginDataTrailer...)
}

func hasDataTrailerFeature(data []byte, feature uint8) bool {
	if len(data) < len(pluginDataTrailer)+1 || !bytes.HasSuffix(data, []byte(pluginDataTrailer)) {
		return false
	}
	return data[len(data)-1-len(pluginDataTrailer)]&feature == feature
}

// supportDataTrailer returns true if v >= v0.4.2-xx
//
// v0.4.2 should be the version we're going to release for this feature
func supportDataTrailer(v string) bool {
	v, _, _ = strings.Cut(v, "-")
	if len(v) < 6 || v[0] != 'v' { // len(v0.0.0) == 6
		return false
	}
	ss := strings.Split(v[1:], ".")
	if len(ss) != 3 {
		return false
	}
	major, _ := strconv.Atoi(ss[0])
	if major > 0 {
		return true // 1.x.x
	}
	minor, _ := strconv.Atoi(ss[1])
	if minor != 4 {
		return minor > 4 // > 0.4.x or < 0.4.x
	}
	patch, _ := strconv.Atoi(ss[2])
	return patch >= 2
}

const refFilenamePrefix = "THRIFGO_REF:"

// compressThriftInclude compresses duplicated includes and keep the 1st one
//
// includes can be considered as DAG,
// then we can simply keep the 1st one and trim others with refFilenamePrefix prefix
func compressThriftInclude(p *parser.Thrift, m map[string]*parser.Thrift) {
	if m == nil {
		m = map[string]*parser.Thrift{}
	}
	for _, incl := range p.Includes {
		if m[incl.Reference.Filename] != nil {
			// visited, only keep the filename for mapping
			incl.Reference = &parser.Thrift{Filename: refFilenamePrefix + incl.Reference.Filename}
		} else {
			// mark it's visited
			m[incl.Reference.Filename] = incl.Reference
			compressThriftInclude(incl.Reference, m)
		}
	}
}

func decompressThriftInclude(p *parser.Thrift, m map[string]*parser.Thrift) {
	if m == nil {
		m = map[string]*parser.Thrift{}
		collectThriftInclude(p, m)
	}
	for _, incl := range p.Includes {
		if fn := strings.TrimPrefix(incl.Reference.Filename, refFilenamePrefix); fn != incl.Reference.Filename {
			incl.Reference = m[fn]
			if incl.Reference == nil {
				panic("not found ref: " + fn)
			}
		}
		decompressThriftInclude(incl.Reference, m)
	}
}

func collectThriftInclude(p *parser.Thrift, m map[string]*parser.Thrift) {
	for _, incl := range p.Includes {
		if !strings.HasPrefix(incl.Reference.Filename, refFilenamePrefix) {
			m[incl.Reference.Filename] = incl.Reference
			collectThriftInclude(incl.Reference, m)
		}
	}
}
