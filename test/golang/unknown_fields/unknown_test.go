// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unknowntest

import (
	"reflect"
	"testing"

	"github.com/apache/thrift/lib/go/thrift"

	ext "example.com/test/gen-ext/unknown" // generated with `keep_unknown_fields`
	neu "example.com/test/gen-new/unknown"
	old "example.com/test/gen-old/unknown"
)

func Encode(obj thrift.TStruct) ([]byte, error) {
	buf := thrift.NewTMemoryBuffer()
	bin := thrift.NewTBinaryProtocolTransport(buf)

	if err := obj.Write(bin); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(obj thrift.TStruct, data []byte) error {
	buf := thrift.NewTMemoryBuffer()
	bin := thrift.NewTBinaryProtocolTransport(buf)
	buf.Write(data)

	return obj.Read(bin)
}

func CreateOldObjectsWithData() (objs []thrift.TStruct) {
	a := &old.Empty{}
	b := &old.Struct{
		Bool: true,
		Byte: 123,
		I16:  &(&struct{ x int16 }{16}).x,
		Str:  &(&struct{ x string }{"string"}).x,
	}
	c := &old.Union{Double: &(&struct{ x float64 }{1.23}).x}
	d := &old.Exception{Str: "string"}
	e := &old.Merged{S: b, U: c, E: d}

	return append(objs, a, b, c, d, e)
}

func CreateNewObjectsWithData() (objs []thrift.TStruct) {
	a := &neu.Empty{
		Str: &(&struct{ x string }{"empty"}).x,
		I16: &(&struct{ x int16 }{16}).x,
	}
	b := &neu.Struct{
		Bool:     true,
		Byte:     123,
		I16:      &(&struct{ x int16 }{16}).x,
		I32:      &(&struct{ x int32 }{32}).x,
		Str2Str:  map[string]string{"abc": "123", "xyz": "!@#"},
		NotEmpty: a,
		Bin:      []byte("asdf"),
		Str:      &(&struct{ x string }{"string"}).x,
		Strs:     []string{"123", "456", "789"},
	}
	c := &neu.Union{Double: &(&struct{ x float64 }{1.23}).x}
	d := &neu.Exception{Str: "string", Str2: "string2"}
	e := &neu.Merged{S: b, U: c, E: d, Ns: b}

	return append(objs, a, b, c, d, e)
}

func CreateNewObjects() (objs []thrift.TStruct) {
	return append(objs,
		neu.NewEmpty(),
		neu.NewStruct(),
		neu.NewUnion(),
		neu.NewException(),
		neu.NewMerged(),
	)
}

func CreateOldObjects() (objs []thrift.TStruct) {
	return append(objs,
		old.NewEmpty(),
		old.NewStruct(),
		old.NewUnion(),
		old.NewException(),
		old.NewMerged(),
	)
}

func CreateExtObjects() (objs []thrift.TStruct) {
	return append(objs,
		ext.NewEmpty(),
		ext.NewStruct(),
		ext.NewUnion(),
		ext.NewException(),
		ext.NewMerged(),
	)
}

// TestEncodeDecode ensures that types generated from the old and new IDLs
// codec can interoperate correctly.
func TestEncodeDecode(t *testing.T) {
	old1 := CreateOldObjectsWithData()
	old2 := CreateExtObjects()
	old3 := CreateOldObjects()
	for i := range old1 {
		data, err := Encode(old1[i])
		if err != nil {
			t.Fatalf("Encode failed: obj[%T] err[%v]", old1[i], err)
		}
		err = Decode(old2[i], data)
		if err != nil {
			t.Fatalf("Decode failed: obj[%T] err[%v]", old1[i], err)
		}
		data, err = Encode(old2[i])
		if err != nil {
			t.Fatalf("Encode failed: obj[%T] err[%v]", old2[i], err)
		}
		err = Decode(old3[i], data)
		if err != nil {
			t.Fatalf("Decode failed: obj[%T] err[%v]", old3[i], err)
		}
		if !reflect.DeepEqual(old1[i], old3[i]) {
			t.Fatalf("No equal: obj1[%+v] obj2[%+v]", old1[i], old2[i])
		}
	}
}

// TestTransfer ensures that types generated from the old IDL with `keep_unknown_fields` can carry new fields for the new IDL.
func TestTransfer(t *testing.T) {
	new1 := CreateNewObjectsWithData()
	new2 := CreateNewObjects()
	old1 := CreateExtObjects()
	old2 := CreateOldObjects()
	for i := range new1 {
		data, err := Encode(new1[i])
		if err != nil {
			t.Fatalf("Encode failed: obj[%T] err[%v]", new1[i], err)
		}

		err = Decode(old1[i], data)
		if err != nil {
			t.Fatalf("Decode failed: obj[%T] err[%v]", new1[i], err)
		}

		data, err = Encode(old1[i])
		if err != nil {
			t.Fatalf("Encode failed: obj[%T] err[%v]", old1[i], err)
		}

		err = Decode(new2[i], data)
		if err != nil {
			t.Fatalf("Decode failed: obj[%T] err[%v]", old1[i], err)
		}

		err = Decode(old2[i], data)
		if err != nil {
			t.Fatalf("Decode failed: obj[%T] err[%v]", old1[i], err)
		}

		if !reflect.DeepEqual(new1[i], new2[i]) {
			t.Fatalf("No equal: obj1[%+v] obj2[%+v]", new1[i], new2[i])
		}
	}
}
