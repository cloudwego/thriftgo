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
	"sync"

	"github.com/cloudwego/thriftgo/thrift_reflection"
)

const (
	_CAP_FM_MAP     = 16
	_CAP_FM_SUB_MAP = 8
)

var fieldmasks = struct {
	mux sync.RWMutex
	m   map[*thrift_reflection.StructDescriptor]map[uint64]*FieldMask
}{
	m: make(map[*thrift_reflection.StructDescriptor]map[uint64]*FieldMask, _CAP_FM_MAP),
}

func RegisterFieldMask(id uint64, desc *thrift_reflection.StructDescriptor, fm *FieldMask) {
	fieldmasks.mux.Lock()
	m := fieldmasks.m[desc]
	if m == nil {
		m = make(map[uint64]*FieldMask, _CAP_FM_SUB_MAP)
		fieldmasks.m[desc] = m
	}
	m[id] = fm
	fieldmasks.mux.Unlock()
}

func GetFieldMask(id uint64, desc *thrift_reflection.StructDescriptor) (ret *FieldMask) {
	fieldmasks.mux.RLock()
	m := fieldmasks.m[desc]
	if m != nil {
		ret = m[id]
	}
	fieldmasks.mux.RUnlock()
	return
}
