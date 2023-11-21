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
	"testing"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/thriftgo/fieldmask"
	"github.com/cloudwego/thriftgo/internal/test_util"
	"github.com/cloudwego/thriftgo/plugin"
	nbase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-new/base"
	obase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-old/base"
	"github.com/stretchr/testify/require"
)

func TestGen(t *testing.T) {
	g, r := test_util.GenerateGolang("a.thrift", "gen-old/", nil, nil)
	if err := g.Persist(r); err != nil {
		panic(err)
	}
	g, r = test_util.GenerateGolang("a.thrift", "gen-new/", []plugin.Option{
		{"with_field_mask", ""},
		{"with_reflection", ""},
	}, nil)
	if err := g.Persist(r); err != nil {
		panic(err)
	}
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
	v1 := nbase.NewVal()
	v1.ID = "b"
	obj.Meta.List = []*nbase.Val{v0, v1}
	obj.Meta.Set = []*nbase.Val{v0, v1}
	obj.Extra = nbase.NewExtraInfo()
	obj.TrafficEnv = nbase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
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
	v1 := obase.NewVal()
	v1.ID = "b"
	obj.Meta.List = []*obase.Val{v0, v1}
	obj.Meta.Set = []*obase.Val{v0, v1}
	obj.Extra = obase.NewExtraInfo()
	obj.TrafficEnv = obase.NewTrafficEnv()
	obj.TrafficEnv.Code = 1
	obj.TrafficEnv.Env = "abcd"
	obj.TrafficEnv.Name = "abcd"
	obj.TrafficEnv.Open = true
	return obj
}

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

		fm, err := fieldmask.GetFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < b.N; i++ {
			obj.SetFieldMask(fm)
			if err := obj.Write(t); err != nil {
				b.Fatal(err)
			}
			buf.Reset()
		}
		fm.Recycle()
	})
}

func BenchmarkReadWithFieldMask(b *testing.B) {
	b.Run("old", func(b *testing.B) {
		obj := SampleOldBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(t); err != nil {
			b.Fatal(err)
		}
		data := []byte(buf.String())
		obj = obase.NewBase()

		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("new", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(t); err != nil {
			b.Fatal(err)
		}
		data := []byte(buf.String())
		obj = nbase.NewBase()

		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("new-mask-half", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)
		if err := obj.Write(t); err != nil {
			b.Fatal(err)
		}
		data := []byte(buf.String())
		obj = nbase.NewBase()

		fm, err := fieldmask.GetFieldMask(obj.GetTypeDescriptor(), "$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
		if err != nil {
			b.Fatal(err)
		}

		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			obj.SetFieldMask(fm)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}

		fm.Recycle()
	})
}

func TestFieldmaskWrite(t *testing.T) {
	obj := SampleNewBase()
	buf := thrift.NewTMemoryBufferLen(1024)
	prot := thrift.NewTBinaryProtocol(buf, true, true)

	fm, err := fieldmask.GetFieldMask(obj.GetTypeDescriptor(),
		"$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
	if err != nil {
		t.Fatal(err)
	}
	obj.SetFieldMask(fm)
	if err := obj.Write(prot); err != nil {
		t.Fatal(err)
	}

	obj2 := nbase.NewBase()
	err = obj2.Read(prot)
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
	require.Equal(t, "b", obj2.Meta.List[0].ID)
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, "b", obj2.Meta.Set[0].ID)
	require.Equal(t, 1, len(obj2.Meta.Set))
	fm.Recycle()
}

func TestFieldmaskRead(t *testing.T) {
	obj := SampleNewBase()
	buf := thrift.NewTMemoryBufferLen(1024)
	prot := thrift.NewTBinaryProtocol(buf, true, true)

	fm, err := fieldmask.GetFieldMask(obj.GetTypeDescriptor(),
		"$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
	if err != nil {
		t.Fatal(err)
	}

	if err := obj.Write(prot); err != nil {
		t.Fatal(err)
	}

	obj2 := nbase.NewBase()
	obj2.SetFieldMask(fm)
	err = obj2.Read(prot)
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
	require.Equal(t, "b", obj2.Meta.List[0].ID)
	require.Equal(t, 1, len(obj2.Meta.List))
	require.Equal(t, "b", obj2.Meta.Set[0].ID)
	require.Equal(t, 1, len(obj2.Meta.Set))
	fm.Recycle()
}
