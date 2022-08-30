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

package test

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"
)

// ThriftRandomFill fills a thrift object with random values.
// The isUnion argument may used to specify union types which require
// exactly one field to be set.
func ThriftRandomFill(obj interface{}, isUnion map[reflect.Type]bool) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%+v", x)
			}
		}
	}()
	env := environment{
		marked:  make(map[reflect.Type]bool),
		counts:  make(map[reflect.Kind]int),
		isUnion: isUnion,
	}
	value := reflect.ValueOf(obj)
	randomFill(value, env, 0, false)
	return
}

const maxDepth = 64

type environment struct {
	marked  map[reflect.Type]bool
	counts  map[reflect.Kind]int
	isUnion map[reflect.Type]bool
}

func randomFill(x reflect.Value, env environment, depth int, required bool) {
	if depth >= maxDepth {
		panic("exceed max depth")
	}

	var v reflect.Value
	switch k := x.Kind(); k {
	case reflect.Bool:
		b := true
		v = reflect.ValueOf(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Uint8:
		env.counts[k] = (env.counts[k] + 1) % 256
		v = reflect.ValueOf(env.counts[k])
	case reflect.Map:
		t := x.Type()
		if x.IsNil() {
			m := reflect.MakeMap(t)
			x.Set(m)
		}
		if x.Len() == 0 {
			k := reflect.New(t.Key()).Elem()
			v := reflect.New(t.Elem()).Elem()
			randomFill(k, env, depth+1, true)
			randomFill(v, env, depth+1, true)
			x.SetMapIndex(k, v)
		}
		return
	case reflect.Ptr:
		if x.IsNil() {
			rt := x.Type().Elem()
			e := reflect.New(rt)
			x.Set(e)
		}
		randomFill(x.Elem(), env, depth+1, required)
		return
	case reflect.Slice:
		if x.IsNil() || x.Len() == 0 {
			s := reflect.MakeSlice(x.Type(), 1, 1)
			x.Set(s)
		}
		randomFill(x.Index(0), env, depth+1, true)
		return
	case reflect.String:
		env.counts[k] = (env.counts[k] + 1) % 256
		s := fmt.Sprint(env.counts[k])
		v = reflect.ValueOf(s)
	case reflect.Struct:
		rt := x.Type()
		if env.marked[rt] && !env.isUnion[rt] && !required {
			return
		}
		env.marked[rt] = true
		var exported []int
		for i := 0; i < rt.NumField(); i++ {
			// We can't use this API because it is introduced in go1.17 and we must be compatible with lower version of golang
			// if rt.Field(i).IsExported() {
			if token.IsExported(rt.Field(i).Name) {
				exported = append(exported, i)
			}
		}
		if len(exported) == 0 {
			return
		}
		if env.isUnion[rt] {
			f := x.Field(exported[0])
			randomFill(f, env, depth+1, false)
		} else {
			for _, i := range exported {
				f := x.Field(i)
				tag := rt.Field(i).Tag.Get("thrift")
				randomFill(f, env, depth+1, !strings.Contains(tag, "optional"))
			}
		}
		return
	default:
		panic(fmt.Sprintf("unexpected type: %s", x.Type()))
	}
	x.Set(v.Convert(x.Type()))
	return
}
