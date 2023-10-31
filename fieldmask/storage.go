/*
 * Copyright 2023 ByteDance Inc.
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

package fieldmask

import "github.com/cloudwego/thriftgo/thrift_reflection"

type fieldID int32

const _MaxFieldIDHead = 128

type fieldMap struct {
	head [_MaxFieldIDHead + 1]FieldMask
	tail []FieldMask
}

func makeFieldMaskMap(st *thrift_reflection.StructDescriptor) fieldMap {
	max := 0
	count := 0
	for _, f := range st.GetFields() {
		if max < int(f.GetID()) {
			max = int(f.GetID())
			count = 0
		} else {
			count += 1
		}
	}
	return fieldMap{
		tail: make([]FieldMask, 0, count),
	}
}

func (self *fieldMap) IsInitialized() bool {
	return self != nil && self.tail != nil
}

func (self *fieldMap) Reset() {
	if self == nil {
		return
	}
	self.tail = self.tail[:0]
}

func (self *fieldMap) GetOrAlloc(f fieldID) *FieldMask {
	if f <= _MaxFieldIDHead {
		return &self.head[f]
	} else {
		if int(f) >= cap(self.tail) {
			tmp := make([]FieldMask, len(self.tail), int(f)+cap(self.tail)>>1+1)
			copy(tmp, self.tail)
			self.tail = tmp
		}
		if int(f) >= len(self.tail) {
			self.tail = self.tail[:f+1]
		}
		return &self.tail[f]
	}
}

func (self *fieldMap) Get(f fieldID) *FieldMask {
	if f <= _MaxFieldIDHead {
		return &self.head[f]
	} else {
		if int(f) >= len(self.tail) {
			return nil
		}
		return &self.tail[f]
	}
}

type fieldMaskBitmap []byte

const _BucketBit = 8

func (self *fieldMaskBitmap) Set(f fieldID) {
	b := int(f / _BucketBit)
	i := int(f % _BucketBit)
	c := cap(*self)
	if c <= b+1 {
		tmp := make([]byte, len(*self), (c + b + 1))
		copy(tmp, *self)
		*self = tmp
	}
	if len(*self) <= b {
		*self = (*self)[:b+1]
	}
	(*self)[b] |= byte(1 << i)
}

func (self *fieldMaskBitmap) Get(f fieldID) bool {
	b := int(f / _BucketBit)
	if len(*self) <= b {
		return false
	}
	i := int(f % _BucketBit)
	return ((*self)[b] & byte(1<<i)) != 0
}

type intMap map[int]bool

func (im intMap) Get(i int) bool {
	return im[i]
}

func (im intMap) Set(i int) {
	im[i] = true
}

func (im intMap) Unset(i int) {
	delete(im, i)
}

type strMap map[string]bool

func (im strMap) Get(i string) bool {
	return im[i]
}

func (im strMap) Set(i string) {
	im[i] = true
}

func (im strMap) Unset(i string) {
	delete(im, i)
}
