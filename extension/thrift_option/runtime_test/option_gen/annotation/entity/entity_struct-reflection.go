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
// Code generated by thriftgo (0.3.2-option-exp). DO NOT EDIT.

package entity

import (
	"reflect"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

// IDL Name: entity_struct
// IDL Path: ../option_idl/annotations/entity/entity_struct.thrift

var file_entity_struct_thrift_go_types = []interface{}{
	(*InnerStruct)(nil), // Struct 0: entity.InnerStruct
}
var file_entity_struct_thrift *thrift_reflection.FileDescriptor
var file_idl_entity_struct_rawDesc = []byte{
	0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xac, 0x90, 0x4b, 0x4e, 0xc3, 0x30,
	0x10, 0x40, 0x5f, 0xf3, 0x69, 0x3, 0xd3, 0x28, 0x17, 0xe0, 0xc, 0xce, 0x8a, 0x43, 0xb0, 0x85,
	0x3, 0x54, 0x11, 0x35, 0xc1, 0x52, 0x19, 0x83, 0x33, 0x5d, 0x70, 0x7b, 0xe4, 0xc4, 0x88, 0x3,
	0x94, 0xd5, 0xf3, 0x47, 0xf3, 0x34, 0x7a, 0xc2, 0xe, 0x78, 0x74, 0x6e, 0x8c, 0x9f, 0x16, 0xa2,
	0x9e, 0xc2, 0xf9, 0x32, 0x4e, 0xaa, 0xd1, 0xa6, 0x7c, 0x5d, 0x46, 0xaf, 0x16, 0xec, 0xbb, 0xe0,
	0xb4, 0x58, 0xba, 0xbe, 0x9a, 0xb3, 0xf7, 0x14, 0xde, 0xac, 0xa7, 0x12, 0x1, 0xe8, 0xa9, 0xd7,
	0x43, 0x56, 0x55, 0x73, 0x4, 0x1e, 0x8a, 0x6d, 0xf6, 0xea, 0xfe, 0x6c, 0x6e, 0xb3, 0xc, 0x34,
	0xc7, 0x3c, 0x36, 0xd0, 0x66, 0xee, 0x6e, 0xd9, 0x41, 0xa8, 0x0, 0x79, 0x52, 0xf5, 0xe9, 0x65,
	0xfd, 0x18, 0xa8, 0xff, 0xc9, 0xda, 0xfa, 0x8f, 0x29, 0x5c, 0x8e, 0xd4, 0xb7, 0xab, 0xf6, 0x8b,
	0xa5, 0xa0, 0x33, 0x42, 0x3, 0x74, 0xcf, 0xfe, 0xeb, 0x1a, 0x92, 0x3f, 0x77, 0xb4, 0x79, 0xd5,
	0x9e, 0x83, 0xc, 0xb9, 0x88, 0xd0, 0xb1, 0x5, 0x6d, 0x7e, 0x1f, 0x5a, 0xb6, 0x54, 0xfb, 0x92,
	0xec, 0x50, 0xd8, 0x15, 0xde, 0x15, 0xde, 0xaf, 0xe4, 0x27, 0x0, 0x0, 0xff, 0xff, 0x88, 0x5f,
	0x20, 0x80, 0xd0, 0x1, 0x0, 0x0,
}

func init() {
	if file_entity_struct_thrift != nil {
		return
	}
	type x struct{}
	builder := &thrift_reflection.FileDescriptorBuilder{
		Bytes:         file_idl_entity_struct_rawDesc,
		GoTypes:       file_entity_struct_thrift_go_types,
		GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
	}
	file_entity_struct_thrift = thrift_reflection.BuildFileDescriptor(builder)
}

func GetFileDescriptorForEntityStruct() *thrift_reflection.FileDescriptor {
	return file_entity_struct_thrift
}
func (p *InnerStruct) GetDescriptor() *thrift_reflection.StructDescriptor {
	return file_entity_struct_thrift.GetStructDescriptor("InnerStruct")
}
