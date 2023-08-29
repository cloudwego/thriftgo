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
	"os"
	"testing"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/thriftgo/fieldmask"
	"github.com/cloudwego/thriftgo/plugin"
	nbase "github.com/cloudwego/thriftgo/test/golang/fieldmask/output/new/base"
	obase "github.com/cloudwego/thriftgo/test/golang/fieldmask/output/old/base"
	"github.com/cloudwego/thriftgo/test/golang/test_util"
)

func TestMain(m *testing.M) {
	g, r := test_util.GenerateGolang("a.thrift", "output/old/", nil, nil)
	if err := g.Persist(r); err != nil {
		panic(err)
	}
	g, r = test_util.GenerateGolang("a.thrift", "output/new/", []plugin.Option{
		{"with_field_mask", ""},
		{"with_reflection", ""},
	}, nil)
	if err := g.Persist(r); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func SampleNewBase() *nbase.Base {
	obj := nbase.NewBase()
	obj.Addr = "abcd"
	obj.Caller = "abcd"
	obj.LogID = "abcd"
	obj.Meta = nbase.NewMetaInfo()
	obj.Meta.PersistentKVS = map[string]string{
		"abcd": "abcd",
	}
	obj.Meta.TransientKVS = map[string]string{
		"abcd": "abcd",
	}
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
	obj.Meta.PersistentKVS = map[string]string{
		"abcd": "abcd",
	}
	obj.Meta.TransientKVS = map[string]string{
		"abcd": "abcd",
	}
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
			// if err := obj.Read(t); err != nil {
			// 	b.Fatal(err)
			// }
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
			// if err := obj.Read(t); err != nil {
			// 	b.Fatal(err)
			// }
			buf.Reset()
		}
	})

	b.Run("new-mask-half", func(b *testing.B) {
		obj := SampleNewBase()
		buf := thrift.NewTMemoryBufferLen(1024)
		t := thrift.NewTBinaryProtocol(buf, true, true)

		fm := fieldmask.NewFieldMaskFromNames(obj.GetDescriptor(), "Addr", "LogID", "Meta.PersistentKVS", "TrafficEnv.Code", "TrafficEnv.Env")
		obj.SetFieldMask(fm)

		for i := 0; i < b.N; i++ {
			if err := obj.Write(t); err != nil {
				b.Fatal(err)
			}
			// if err := obj.Read(t); err != nil {
			// 	b.Fatal(err)
			// }
			buf.Reset()
		}
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
		data := []byte(string(buf.Bytes()))
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
		data := []byte(string(buf.Bytes()))
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
		data := []byte(string(buf.Bytes()))
		obj = nbase.NewBase()

		fm := fieldmask.NewFieldMaskFromNames(obj.GetDescriptor(), "Addr", "LogID", "Meta.PersistentKVS", "TrafficEnv.Code", "TrafficEnv.Env")
		obj.SetFieldMask(fm)

		for i := 0; i < b.N; i++ {
			buf.Reset()
			buf.Write(data)
			if err := obj.Read(t); err != nil {
				b.Fatal(err)
			}
		}
	})
}
