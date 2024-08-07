/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fastgo

import (
	"sort"

	"github.com/cloudwego/thriftgo/generator/golang"
	"github.com/cloudwego/thriftgo/parser"
)

func isContainerType(f *parser.Type) bool {
	switch f.Category {
	case parser.Category_Map,
		parser.Category_List,
		parser.Category_Set:
		return true
	case parser.Category_Binary:
		return true // []byte, a byte list
	}
	return false
}

func varnameVal(pointer bool, varname string) string {
	if pointer {
		return "*" + varname
	}
	return varname
}

func varnamePtr(pointer bool, varname string) string {
	if pointer {
		return varname
	}
	return "&" + varname
}

// getSortedFields returns fields sorted by field id.
// we don't want to see code changes due to field order.
func getSortedFields(s *golang.StructLike) []*golang.Field {
	ff := append([]*golang.Field(nil), s.Fields()...)
	sort.Slice(ff, func(i, j int) bool { return ff[i].ID < ff[j].ID })
	return ff
}
