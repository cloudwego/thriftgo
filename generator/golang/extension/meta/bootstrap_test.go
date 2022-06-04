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

package meta

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/cloudwego/thriftgo/pkg/test"
)

func TestRandomValue(t *testing.T) {
	ctx := context.Background()
	mem := new(MemoryTransport)
	bin := NewBinaryProtocol(mem).WithStrictRead().WithStrictWrite()

	log := new(bytes.Buffer)
	debug := NewDebugProtocol(bin).WithLogFunc(
		func(format string, a ...interface{}) {
			fmt.Fprintf(log, format, a...)
			log.WriteByte('\n')
		})

	isUnion := make(map[reflect.Type]bool)
	for rt, st := range structs {
		if st.Category == "union" {
			isUnion[rt] = true
		}
	}

	for rt, st := range structs {
		t.Logf("testing registered type: %s", rt)
		v := st.newFunc.Call(nil)[0].Interface()
		test.ThriftRandomFill(v, isUnion)

		log.Reset()
		mem.Reset()
		sv, err := AsStruct(v)
		test.Assert(t, err == nil && sv != nil, sv, err)

		err = sv.Write(ctx, debug)
		test.Assert(t, err == nil, err)
		bytes0 := mem.Bytes()
		logs0 := log.String()
		test.Assert(t, len(bytes0) > 0)

		w := st.newFunc.Call(nil)[0].Interface()
		test.ThriftRandomFill(w, isUnion)

		u := st.newFunc.Call(nil)[0].Interface()
		su, err := AsStruct(u)
		test.Assert(t, err == nil && su != nil, su, err)

		err = su.Read(ctx, bin)
		test.Assert(t, err == nil)

		log.Reset()
		mem.Reset()
		err = su.Write(ctx, debug)
		test.Assert(t, err == nil, err)
		bytes1 := mem.Bytes()
		logs1 := log.String()

		if !reflect.DeepEqual(bytes0, bytes1) {
			dumpDetail("a.txt", v, logs0, bytes0)
			dumpDetail("b.txt", u, logs1, bytes1)
			t.Fatal("inconsistent read/write")
		}
	}
}

func dumpDetail(fn string, obj interface{}, log string, bites []byte) {
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	test.DeepPrint(f, obj)
	f.Write([]byte{'\n'})
	fmt.Fprintf(f, "%s\n", log)
	f.Write([]byte{'\n'})
	fmt.Fprintf(f, "%#v\n", bites)
}
