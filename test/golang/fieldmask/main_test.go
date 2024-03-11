// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fieldmask

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/thriftgo/fieldmask"

	hbase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-halfway/base"
	nbase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-new/base"
	obase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-old/base"
	zbase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-zero/base"
	"github.com/stretchr/testify/require"
)

var fieldmaskCache sync.Map

func initFielMask() {
	// new a obj to get its TypeDescriptor
	obj := nbase.NewBase()

	// construct a fieldmask with TypeDescriptor and thrift pathes
	fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(),
		"$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}",
		"$.Meta.List[1]", "$.Meta.Set[0].id", "$.Meta.Set[1].name")
	if err != nil {
		panic(err)
	}

	// cache it for future usage of nbase.Base
	fieldmaskCache.Store("Mask1ForBase", fm)
}

func TestFieldMask_Write(t *testing.T) {
	initFielMask()
	// biz logic: handle and get final response object
	obj := SampleNewBase()

	// Load fieldmask from cache
	fm, _ := fieldmaskCache.Load("Mask1ForBase")
	if fm != nil {
		// load ok, set fieldmask onto the object using codegen API
		obj.Set_FieldMask(fm.(*fieldmask.FieldMask))
	}

	// return obj

	// prepare buffer
	buf := thrift.NewTMemoryBufferLen(1024)
	prot := thrift.NewTBinaryProtocol(buf, true, true)
	if err := obj.Write(prot); err != nil {
		t.Fatal(err)
	}

	// validate output
	obj2 := nbase.NewBase()
	err := obj2.Read(prot)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, obj.Addr, obj2.Addr)
	require.Equal(t, obj.LogID, obj2.LogID)
	require.Equal(t, "", obj2.Caller)
	require.Equal(t, "", obj2.TrafficEnv.Name)
	require.Equal(t, false, obj2.TrafficEnv.Open)
	require.Equal(t, "", obj2.TrafficEnv.Env)
	require.Equal(t, obj.TrafficEnv.Code, obj2.TrafficEnv.Code)
	require.Equal(t, obj.Meta.IntMap[1].ID, obj2.Meta.IntMap[1].ID)
	require.Equal(t, (*nbase.Val)(nil), obj2.Meta.IntMap[0])
	require.Equal(t, obj.Meta.StrMap["1234"].ID, obj2.Meta.StrMap["1234"].ID)
	require.Equal(t, (*nbase.Val)(nil), obj2.Meta.StrMap["abcd"])
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, "b", obj2.Meta.List[0].ID)
	require.Equal(t, "b", obj2.Meta.List[0].Name)
	require.Equal(t, 2, len(obj2.Meta.Set))
	require.Equal(t, "a", obj2.Meta.Set[0].ID)
	require.Equal(t, "", obj2.Meta.Set[0].Name)
	require.Equal(t, "", obj2.Meta.Set[1].ID)
	require.Equal(t, "b", obj2.Meta.Set[1].Name)
}

func TestFieldMask_Read(t *testing.T) {
	initFielMask()
	obj := SampleNewBase()
	buf := thrift.NewTMemoryBufferLen(1024)
	prot := thrift.NewTBinaryProtocol(buf, true, true)

	if err := obj.Write(prot); err != nil {
		t.Fatal(err)
	}

	obj2 := nbase.NewBase()
	fm, _ := fieldmaskCache.Load("Mask1ForBase")
	if fm != nil {
		obj2.Set_FieldMask(fm.(*fieldmask.FieldMask))
	}

	err := obj2.Read(prot)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, "", obj2.Addr)
	require.Equal(t, obj.LogID, obj2.LogID)
	require.Equal(t, "", obj2.Caller)
	require.Equal(t, "", obj2.TrafficEnv.Name)
	require.Equal(t, false, obj2.TrafficEnv.Open)
	require.Equal(t, "", obj2.TrafficEnv.Env)
	require.Equal(t, obj.TrafficEnv.Code, obj2.TrafficEnv.Code)
	require.Equal(t, obj.Meta.IntMap[1].ID, obj2.Meta.IntMap[1].ID)
	require.Equal(t, (*nbase.Val)(nil), obj2.Meta.IntMap[0])
	require.Equal(t, obj.Meta.StrMap["1234"].ID, obj2.Meta.StrMap["1234"].ID)
	require.Equal(t, (*nbase.Val)(nil), obj2.Meta.StrMap["abcd"])
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, "b", obj2.Meta.List[0].ID)
	require.Equal(t, "b", obj2.Meta.List[0].Name)
	require.Equal(t, 2, len(obj2.Meta.Set))
	require.Equal(t, "a", obj2.Meta.Set[0].ID)
	require.Equal(t, "", obj2.Meta.Set[0].Name)
	require.Equal(t, "", obj2.Meta.Set[1].ID)
	require.Equal(t, "b", obj2.Meta.Set[1].Name)
}

func TestMaskRequired(t *testing.T) {
	fm, err := fieldmask.NewFieldMask(nbase.NewBaseResp().GetTypeDescriptor(), "$.F1", "$.F8", "$.R12.Env")
	if err != nil {
		t.Fatal(err)
	}
	j, err := fm.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	println(string(j))
	nf, ex := fm.Field(111)
	if !ex {
		t.Fatal(nf)
	}

	t.Run("read", func(t *testing.T) {
		obj := nbase.NewBaseResp()
		obj.F1 = map[nbase.Str]nbase.Str{"a": "b"}
		obj.F8 = map[float64][]nbase.Str{1.0: {"a"}}
		obj.R12 = nbase.NewTrafficEnv()
		obj.R12.Name = "a"
		obj.R12.Env = "a"
		buf := thrift.NewTMemoryBufferLen(1024)
		prot := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(prot); err != nil {
			t.Fatal(err)
		}
		obj2 := nbase.NewBaseResp()
		obj2.Set_FieldMask(fm)
		if err := obj2.Read(prot); err != nil {
			t.Fatal(err)
		}
		require.Equal(t, obj.F1, obj2.F1)
		require.Equal(t, obj.F8, obj2.F8)
		require.Equal(t, "", obj2.R12.Name)
		require.Equal(t, obj.R12.Env, obj2.R12.Env)
	})

	t.Run("current-write", func(t *testing.T) {
		obj := nbase.NewBaseResp()
		obj.StatusCode = 1
		obj.R3 = true
		obj.R4 = 1
		obj.R5 = 1
		obj.R6 = 1
		obj.R7 = 1
		obj.R8 = "R8"
		obj.R9 = nbase.Ex_B
		v := nbase.NewVal()
		v.ID = "a"
		obj.R10 = []*nbase.Val{v}
		obj.R11 = []*nbase.Val{v}
		obj.R12 = nbase.NewTrafficEnv()
		obj.R12.Name = "a"
		obj.R12.Env = "a"
		obj.R13 = map[string]*nbase.Key{"a": v}
		obj.F8 = map[float64][]nbase.Str{1.0: {"a"}}
		obj.F1 = map[nbase.Str]nbase.Str{"a": "b"}

		obj.Set_FieldMask(fm)
		buf := thrift.NewTMemoryBufferLen(1024)
		prot := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(prot); err != nil {
			t.Fatal(err)
		}

		obj2 := nbase.NewBaseResp()
		if err := obj2.Read(prot); err != nil {
			t.Fatal(err)
		}

		require.Equal(t, obj.F1, obj2.F1)
		require.Equal(t, obj.F8, obj2.F8)
		require.Equal(t, obj.StatusCode, obj2.StatusCode)
		require.Equal(t, obj.R3, obj2.R3)
		require.Equal(t, obj.R4, obj2.R4)
		require.Equal(t, obj.R5, obj2.R5)
		require.Equal(t, obj.R6, obj2.R6)
		require.Equal(t, obj.R7, obj2.R7)
		require.Equal(t, obj.R8, obj2.R8)
		require.Equal(t, obj.R9, obj2.R9)
		require.Equal(t, obj.R10, obj2.R10)
		require.Equal(t, obj.R11, obj2.R11)
		require.Equal(t, "", obj2.R12.Name)
		require.Equal(t, obj.R12.Env, obj2.R12.Env)
		require.Equal(t, obj.R13, obj2.R13)
	})

	t.Run("zero-required-write", func(t *testing.T) {
		fm, err := fieldmask.NewFieldMask(nbase.NewBaseResp().GetTypeDescriptor(), "$.F1", "$.F8", "$.R12")
		if err != nil {
			t.Fatal(err)
		}
		obj := zbase.NewBaseResp()
		obj.F1 = map[zbase.Str]zbase.Str{"a": "b"}
		obj.F8 = map[float64][]zbase.Str{1.0: {"a"}}
		obj.StatusCode = 1
		obj.R3 = true
		obj.R4 = 1
		obj.R5 = 1
		obj.R6 = 1
		obj.R7 = 1
		obj.R8 = "R8"
		obj.R9 = zbase.Ex_B
		v := zbase.NewVal()
		v.ID = "a"
		obj.R10 = []*zbase.Val{v}
		obj.R11 = []*zbase.Val{v}
		obj.R12 = zbase.NewTrafficEnv()
		obj.R13 = map[string]*zbase.Key{"a": v}

		obj.Set_FieldMask(fm)
		buf := thrift.NewTMemoryBufferLen(1024)
		prot := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(prot); err != nil {
			t.Fatal(err)
		}
		obj2 := zbase.NewBaseResp()
		if err := obj2.Read(prot); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%#v\n", obj2)

		require.Equal(t, obj.F1, obj2.F1)
		require.Equal(t, obj.F8, obj2.F8)
		require.Equal(t, int32(0), obj2.StatusCode)
		require.Equal(t, false, obj2.R3)
		require.Equal(t, int8(0), obj2.R4)
		require.Equal(t, int16(0), obj2.R5)
		require.Equal(t, int64(0), obj2.R6)
		require.Equal(t, float64(0), obj2.R7)
		require.Equal(t, "", obj2.R8)
		require.Equal(t, zbase.Ex(0), obj2.R9)
		require.Equal(t, []*zbase.Val{}, obj2.R10)
		require.Equal(t, []*zbase.Val{}, obj2.R11)
		obj.R12.Set_FieldMask(nil)
		require.Equal(t, obj.R12, obj2.R12)
		require.Equal(t, map[string]*zbase.Key{}, obj2.R13)
	})
}

func TestSetMaskHalfway(t *testing.T) {
	obj := hbase.NewBase()
	obj.Extra = hbase.NewExtraInfo()
	obj.Extra.F1 = map[string]string{"a": "b"}
	obj.Extra.F8 = map[int64][]*hbase.Key{1: {hbase.NewKey()}}

	fm, err := fieldmask.NewFieldMask(obj.Extra.GetTypeDescriptor(), "$.F1")
	if err != nil {
		t.Fatal(err)
	}
	obj.Extra.Set_FieldMask(fm)
	buf := thrift.NewTMemoryBufferLen(1024)
	prot := thrift.NewTBinaryProtocol(buf, true, true)
	if err := obj.Write(prot); err != nil {
		t.Fatal(err)
	}

	obj2 := hbase.NewBase()
	if err := obj2.Read(prot); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, obj.Extra.F1, obj2.Extra.F1)
	require.Equal(t, map[int64][]*hbase.Key(nil), obj2.Extra.F8)

	fm, err = fieldmask.NewFieldMask(obj.Extra.GetTypeDescriptor(), "$.F8")
	if err != nil {
		t.Fatal(err)
	}
	obj.Extra.Set_FieldMask(fm)
	if err := obj.Write(prot); err != nil {
		t.Fatal(err)
	}

	obj2 = hbase.NewBase()
	if err := obj2.Read(prot); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, map[string]string(nil), obj2.Extra.F1)
	require.Equal(t, obj.Extra.F8, obj2.Extra.F8)
}

func SampleNewBase() *nbase.Base {
	obj := nbase.NewBase()
	obj.Addr = "abcd"
	obj.Caller = "abcd"
	obj.LogID = "abcd"
	obj.Meta = nbase.NewMetaInfo()
	obj.Meta.StrMap = map[string]*nbase.Val{
		"abcd": nbase.NewVal(),
		"1234": nbase.NewVal(),
	}
	obj.Meta.IntMap = map[int64]*nbase.Val{
		1: nbase.NewVal(),
		2: nbase.NewVal(),
	}
	v0 := nbase.NewVal()
	v0.ID = "a"
	v0.Name = "a"
	v1 := nbase.NewVal()
	v1.ID = "b"
	v1.Name = "b"
	obj.Meta.List = []*nbase.Val{v0, v1}
	// v0 = nbase.NewVal()
	// v0.ID = "a"
	// v0.Name = "a"
	// v1 = nbase.NewVal()
	// v1.ID = "b"
	// v1.Name = "b"
	obj.Meta.Set = []*nbase.Val{v0, v1}
	// obj.Extra = nbase.NewExtraInfo()
	obj.TrafficEnv = nbase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	obj.Meta.Base = nbase.NewBase()
	return obj
}

func SampleOldBase() *obase.Base {
	obj := obase.NewBase()
	obj.Addr = "abcd"
	obj.Caller = "abcd"
	obj.LogID = "abcd"
	obj.Meta = obase.NewMetaInfo()
	obj.Meta.StrMap = map[string]*obase.Val{
		"abcd": obase.NewVal(),
		"1234": obase.NewVal(),
	}
	obj.Meta.IntMap = map[int64]*obase.Val{
		1: obase.NewVal(),
		2: obase.NewVal(),
	}
	v0 := obase.NewVal()
	v0.ID = "a"
	v0.Name = "a"
	v1 := obase.NewVal()
	v1.ID = "b"
	v1.Name = "b"
	obj.Meta.List = []*obase.Val{v0, v1}
	obj.Meta.Set = []*obase.Val{v0, v1}
	// obj.Extra = obase.NewExtraInfo()
	obj.TrafficEnv = obase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	obj.Meta.Base = obase.NewBase()
	return obj
}

// func SampleApacheBase() *abase.Base {
// 	obj := abase.NewBase()
// 	obj.Addr = "abcd"
// 	obj.Caller = "abcd"
// 	obj.LogID = "abcd"
// 	obj.Meta = abase.NewMetaInfo()
// 	obj.Meta.StrMap = map[string]*abase.Val{
// 		"abcd": abase.NewVal(),
// 		"1234": abase.NewVal(),
// 	}
// 	obj.Meta.IntMap = map[int64]*abase.Val{
// 		1: abase.NewVal(),
// 		2: abase.NewVal(),
// 	}
// 	v0 := abase.NewVal()
// 	v0.ID = "a"
// 	v1 := abase.NewVal()
// 	v1.ID = "b"
// 	obj.Meta.List = []*abase.Val{v0, v1}
// 	obj.Meta.Set = []*abase.Val{v0, v1}
// 	// obj.Extra = abase.NewExtraInfo()
// 	obj.TrafficEnv = abase.NewTrafficEnv()
// 	obj.TrafficEnv.Code = 1
// 	obj.TrafficEnv.Env = "abcd"
// 	obj.TrafficEnv.Name = "abcd"
// 	obj.TrafficEnv.Open = true
// 	obj.Meta.Base = abase.NewBase()
// 	return obj
// }

func BenchmarkWriteWithFieldMask(b *testing.B) {
	b.Run("old", func(b *testing.B) {
		obj := SampleOldBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)

		for i := 0; i < b.N; i++ {
			if err := obj.Write(t); err != nil {
				b.Fatal(err)
			}
			buf.Reset()
		}
	})

	b.Run("new", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)

		for i := 0; i < b.N; i++ {
			if err := obj.Write(t); err != nil {
				b.Fatal(err)
			}
			buf.Reset()
		}
	})

	b.Run("new-mask-half", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)

		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.TrafficEnv.Name", "$.Meta.IntMap", "$.Meta.List[*]")
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			obj.Set_FieldMask(fm)
			if err := obj.Write(t); err != nil {
				b.Fatal(err)
			}
			buf.Reset()
		}
	})
}

func BenchmarkReadWithFieldMask(b *testing.B) {
	// b.Run("apache", func(b *testing.B) {
	// 	obj := SampleApacheBase()
	// 	buf := thrift.NewTMemoryBufferLen(1024)
	// 	t := thrift.NewTBinaryProtocol(buf, true, true)
	// 	if err := obj.Write(t); err != nil {
	// 		b.Fatal(err)
	// 	}
	// 	data := []byte(buf.String())

	// 	obj = abase.NewBase()
	// 	b.ResetTimer()
	// 	for i := 0; i < b.N; i++ {
	// 		buf.Reset()
	// 		buf.Write(data)
	// 		if err := obj.Read(t); err != nil {
	// 			b.Fatal(err)
	// 		}
	// 	}
	// })
	b.Run("old", func(b *testing.B) {
		obj := SampleOldBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(t); err != nil {
			b.Fatal(err)
		}
		data := []byte(buf.String())

		obj = obase.NewBase()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}
	})
	runtime.GC()
	b.Run("new", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(t); err != nil {
			b.Fatal(err)
		}
		data := []byte(buf.String())
		obj = nbase.NewBase()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}
	})
	runtime.GC()
	b.Run("new-mask-half", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(t); err != nil {
			b.Fatal(err)
		}
		data := []byte(buf.String())
		obj = nbase.NewBase()
		fm, err := fieldmask.NewFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.TrafficEnv.Name", "$.Meta.IntMap", "$.Meta.List[*]")
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			obj.Set_FieldMask(fm)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}
	})
}
