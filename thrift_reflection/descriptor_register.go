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
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"sync"

	"github.com/cloudwego/thriftgo/parser"
)

type GlobalDescriptor struct {
	uuid     string
	globalFD map[string]*FileDescriptor

	structDes2goType  map[*StructDescriptor]reflect.Type
	enumDes2goType    map[*EnumDescriptor]reflect.Type
	typedefDes2goType map[*TypedefDescriptor]reflect.Type

	goType2StructDes  map[reflect.Type]*StructDescriptor
	goType2EnumDes    map[reflect.Type]*EnumDescriptor
	goType2TypedefDes map[reflect.Type]*TypedefDescriptor
}

var defaultGlobalDescriptor = &GlobalDescriptor{
	uuid:              DEFAULT_GLOBAL_DESCRIPTOR_UUID,
	globalFD:          map[string]*FileDescriptor{},
	structDes2goType:  map[*StructDescriptor]reflect.Type{},
	enumDes2goType:    map[*EnumDescriptor]reflect.Type{},
	typedefDes2goType: map[*TypedefDescriptor]reflect.Type{},
	goType2StructDes:  map[reflect.Type]*StructDescriptor{},
	goType2EnumDes:    map[reflect.Type]*EnumDescriptor{},
	goType2TypedefDes: map[reflect.Type]*TypedefDescriptor{},
}

var globalDescriptorMap = map[string]*GlobalDescriptor{
	DEFAULT_GLOBAL_DESCRIPTOR_UUID: defaultGlobalDescriptor,
}
var lock sync.RWMutex

const (
	DEFAULT_GLOBAL_DESCRIPTOR_UUID = "default"
	GLOBAL_UUID_EXTRA_KEY          = "global_descriptor_uuid"
)

func GetGlobalDescriptor(v interface{ GetExtra() map[string]string }) *GlobalDescriptor {
	uuid := v.GetExtra()[GLOBAL_UUID_EXTRA_KEY]
	if uuid == "" {
		return defaultGlobalDescriptor
	}
	lock.RLock()
	defer lock.RUnlock()
	return globalDescriptorMap[uuid]
}

func addExtraToDescriptor(uuid string, v interface {
	GetExtra() map[string]string
	setExtra(m map[string]string)
}) {
	if v == nil || isNil(v) {
		return
	}
	if v.GetExtra() == nil {
		v.setExtra(map[string]string{})
	}
	v.GetExtra()[GLOBAL_UUID_EXTRA_KEY] = uuid
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}

	vi := reflect.ValueOf(i)

	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}

	return false
}

func addExtraToTypeDescriptor(uuid string, td *TypeDescriptor) {
	if td == nil {
		return
	}
	if td.GetExtra() == nil {
		td.setExtra(map[string]string{})
	}
	td.GetExtra()[GLOBAL_UUID_EXTRA_KEY] = uuid
	addExtraToTypeDescriptor(uuid, td.KeyType)
	addExtraToTypeDescriptor(uuid, td.ValueType)
}

func registerGlobalUUID(fd *FileDescriptor, uuid string) {
	addExtraToDescriptor(uuid, fd)

	for _, service := range fd.Services {
		addExtraToDescriptor(uuid, service)
		for _, method := range service.Methods {
			addExtraToDescriptor(uuid, method)
			addExtraToTypeDescriptor(uuid, method.Response)
			for _, arg := range method.Args {
				addExtraToDescriptor(uuid, arg)
				addExtraToTypeDescriptor(uuid, arg.Type)
			}
			for _, e := range method.ThrowExceptions {
				addExtraToDescriptor(uuid, e)
				addExtraToTypeDescriptor(uuid, e.Type)
			}
		}
	}
	structs := []*StructDescriptor{}
	structs = append(structs, fd.Structs...)
	structs = append(structs, fd.Unions...)
	structs = append(structs, fd.Exceptions...)
	for _, strct := range structs {
		addExtraToDescriptor(uuid, strct)
		for _, f := range strct.Fields {
			addExtraToDescriptor(uuid, f)
			addExtraToTypeDescriptor(uuid, f.Type)
			addExtraToDescriptor(uuid, f.DefaultValue)
		}
	}

	for _, enum := range fd.Enums {
		addExtraToDescriptor(uuid, enum)
		for _, ev := range enum.Values {
			addExtraToDescriptor(uuid, ev)
		}
	}

	for _, typedef := range fd.Typedefs {
		addExtraToDescriptor(uuid, typedef)
		addExtraToTypeDescriptor(uuid, typedef.Type)
	}
	for _, c := range fd.Consts {
		addExtraToDescriptor(uuid, c)
		addExtraToDescriptor(uuid, c.Value)
	}
}

func (gd *GlobalDescriptor) checkDuplicateAndRegister(f *FileDescriptor, currentGoPkgPath string) {
	f.setGoPkgPath(currentGoPkgPath)
	previous, ok := gd.globalFD[f.Filepath]
	if !ok {
		gd.globalFD[f.Filepath] = f
		return
	}

	// just check the content of thrift file
	newFD := *f
	newFD.Extra = nil
	newPrevFD := *previous
	newPrevFD.Extra = nil
	if reflect.DeepEqual(newFD, newPrevFD) {
		return
	}
	panicString := fmt.Sprintf("thrift reflection: file '%s' is already registered\n"+
		"\tpreviously from: '%s'\n"+
		"\tcurrently from '%s'\n"+
		"To solve this, you need to remove one of the idl above.", f.Filepath, previous.getGoPkgPath(), f.getGoPkgPath())
	panic(panicString)
}

func BuildFileDescriptor(builder *FileDescriptorBuilder) *FileDescriptor {
	fd := MustUnmarshal(builder.Bytes)
	defaultGlobalDescriptor.checkDuplicateAndRegister(fd, builder.GoPackagePath)
	defaultGlobalDescriptor.registerGoTypes(fd, builder.GoTypes)
	return fd
}

type FileDescriptorBuilder struct {
	Bytes         []byte
	GoTypes       []interface{}
	GoPackagePath string
}

func (gd *GlobalDescriptor) ShowRegisterInfo() map[string]string {
	info := map[string]string{}
	for _, fd := range gd.globalFD {
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

func RegisterAST(ast *parser.Thrift) (*GlobalDescriptor, *FileDescriptor) {
	gd := &GlobalDescriptor{globalFD: map[string]*FileDescriptor{}}
	gd.uuid = generateShortUUID()
	fd := doRegisterAST(ast, gd.globalFD, gd.uuid)
	lock.Lock()
	defer lock.Unlock()
	globalDescriptorMap[gd.uuid] = gd
	return gd, fd
}

func generateShortUUID() string {
	uuid := make([]byte, 4)
	_, _ = rand.Read(uuid)
	// 将随机生成的字节转换为十六进制字符串
	shortUUID := hex.EncodeToString(uuid)
	return shortUUID
}

func doRegisterAST(ast *parser.Thrift, globalFD map[string]*FileDescriptor, globalUUID string) *FileDescriptor {
	fd, ok := globalFD[ast.Filename]
	if ok {
		return fd
	}
	fd = GetFileDescriptor(ast)
	globalFD[fd.Filepath] = fd
	for _, inc := range ast.Includes {
		doRegisterAST(inc.GetReference(), globalFD, globalUUID)
	}
	registerGlobalUUID(fd, globalUUID)
	return fd
}
