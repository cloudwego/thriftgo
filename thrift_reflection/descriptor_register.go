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

package thrift_reflection

import (
	"fmt"

	"github.com/cloudwego/thriftgo/parser"
)

var globalFD = map[string]*FileDescriptor{}

func checkDuplicateAndRegister(f *FileDescriptor, currentGoPkgPath string) {
	f.setGoPkgPath(currentGoPkgPath)
	if previous, ok := globalFD[f.Filepath]; ok {
		panicString := fmt.Sprintf("thrift reflection: file '%s' is already registered\n"+
			"\tpreviously from: '%s'\n"+
			"\tcurrently from '%s'\n"+
			"To solve this, you need to remove one of the idl above.", f.Filepath, previous.getGoPkgPath(), f.getGoPkgPath())
		panic(panicString)
	}
	globalFD[f.Filepath] = f
}

func BuildFileDescriptor(builder *FileDescriptorBuilder) *FileDescriptor {
	fd := MustUnmarshal(builder.Bytes)
	checkDuplicateAndRegister(fd, builder.GoPackagePath)
	registerGoTypes(fd, builder.GoTypes)
	return fd
}

type FileDescriptorBuilder struct {
	Bytes         []byte
	GoTypes       []interface{}
	GoPackagePath string
}

func ShowRegisterInfo() map[string]string {
	info := map[string]string{}
	for _, fd := range globalFD {
		info[fd.Filepath] = fd.getGoPkgPath()
	}
	return info
}

func (f *FileDescriptor) isGoPkgPathSet() bool {
	return f.Extra != nil && f.Extra["GoPkgPath"] != ""
}

func (f *FileDescriptor) getGoPkgPath() string {
	if f.Extra == nil {
		return ""
	}
	return f.Extra["GoPkgPath"]
}

func (f *FileDescriptor) setGoPkgPath(path string) {
	if f.Extra == nil {
		f.Extra = map[string]string{}
	}
	f.Extra["GoPkgPath"] = path
}

func RegisterAST(ast *parser.Thrift) *FileDescriptor {
	fd, ok := globalFD[ast.Filename]
	if ok {
		return fd
	}
	fd = GetFileDescriptor(ast)
	globalFD[fd.Filepath] = fd
	for _, inc := range ast.Includes {
		RegisterAST(inc.GetReference())
	}
	return fd
}
