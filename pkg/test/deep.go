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
	"io"
	"reflect"
	"sort"
	"strings"
)

// DeepPrint prints the given object with its fields recursively.
// Note that when a pointer is encountered more than once during
// printing, its subsequent occurrences after the first time will
// not be print in detail.
func DeepPrint(out io.Writer, x interface{}) {
	visited := make(map[reflect.Value]bool)
	v := reflect.ValueOf(x)
	deepPrint(out, v, visited, 0)
}

func deepPrint(out io.Writer, v reflect.Value, visited map[reflect.Value]bool, depth int) {
	switch k := v.Kind(); k {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Uintptr, reflect.UnsafePointer:
		fmt.Fprintf(out, "%s(%#v)", v.Type(), v.Interface())
	case reflect.Func, reflect.Interface, reflect.Chan:
		fmt.Fprintf(out, "%s(%v)", v.Type(), v.Interface())
	case reflect.Ptr:
		if v.IsNil() {
			fmt.Fprintf(out, "(%s)(nil)", v.Type())
			return
		}
		if visited[v] {
			fmt.Fprintf(out, "(%s)(%p:visited)", v.Type(), v.Interface())
			return
		}
		visited[v] = true
		fmt.Fprintf(out, "/* %p */&", v.Interface())
		deepPrint(out, v.Elem(), visited, depth)
	case reflect.Array, reflect.Slice:
		if k == reflect.Array {
			fmt.Fprintf(out, "[%d]%s{", v.Len(), v.Type().Elem())
		} else {
			fmt.Fprintf(out, "[]%s{", v.Type().Elem())
		}
		if v.Len() > 0 {
			fmt.Fprint(out, "\n")
			for i := 0; i < v.Len(); i++ {
				fmt.Fprint(out, strings.Repeat("\t", depth+1))
				deepPrint(out, v.Index(i), visited, depth+1)
				fmt.Fprint(out, ",\n")
			}
			fmt.Fprint(out, strings.Repeat("\t", depth))
		}
		fmt.Fprint(out, "}")
	case reflect.Map:
		keys := v.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		fmt.Fprintf(out, "map[%s]%s{", v.Type().Key(), v.Type().Elem())
		if v.Len() > 0 {
			fmt.Fprint(out, "\n")
			for _, k := range keys {
				fmt.Fprint(out, strings.Repeat("\t", depth+1))
				deepPrint(out, v.MapIndex(k), visited, depth+1)
				fmt.Fprint(out, ",\n")
			}
			fmt.Fprint(out, strings.Repeat("\t", depth))
		}
		fmt.Fprintf(out, "}")
	case reflect.Struct:
		t := v.Type()
		fmt.Fprintf(out, "%s{", t)
		if t.NumField() > 0 {
			fmt.Fprintf(out, "\n")
			for i := 0; i < t.NumField(); i++ {
				fmt.Fprint(out, strings.Repeat("\t", depth+1))
				ft := t.Field(i)
				if ft.Anonymous {
					fmt.Fprintf(out, "%s: ", t.Field(i).Type)
				} else {
					fmt.Fprintf(out, "%s: ", t.Field(i).Name)
				}
				deepPrint(out, v.Field(i), visited, depth+1)
				fmt.Fprint(out, ",\n")
			}
			fmt.Fprint(out, strings.Repeat("\t", depth))
		}
		fmt.Fprintf(out, "}")
	default:
		fmt.Fprintf(out, "<unsupported kind: %v>", k)
	}
}
