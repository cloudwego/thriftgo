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

type fieldID int32

const _MaxFieldIDHead = 63

type fieldMap struct {
	head [_MaxFieldIDHead + 1]*FieldMask
	tail map[fieldID]*FieldMask
}

func makeFieldMaskMap() fieldMap {
	return fieldMap{
		tail: make(map[fieldID]*FieldMask),
	}
}

func (fm *fieldMap) Reset() {
	if fm == nil {
		return
	}
	for _, v := range fm.tail {
		v.reset()
	}
	// memclrNoHeapPointers(unsafe.Pointer(&fm.head), 8*(_MaxFieldIDHead+1))
	for _, v := range fm.head {
		v.reset()
	}
}

// func (self *fieldMap) Reset() {
// 	if self == nil {
// 		return
// 	}
// 	self.tail = self.tail[:0]
// }

func (self *fieldMap) SetIfNotExist(f fieldID, ft FieldMaskType, black bool) (s *FieldMask) {
	if f <= _MaxFieldIDHead {
		s = self.head[f]
		if s == nil {
			fm := newFieldMask(ft, black)
			self.head[f] = &fm
			return &fm
		}

	} else {
		s = self.tail[f]
		if s == nil {
			fm := newFieldMask(ft, black)
			self.tail[f] = &fm
			return &fm
		}
	}
	if s.typ == 0 {
		s.assign(ft, black)
	}
	return s
}

func (self *fieldMap) Get(f fieldID) (ret *FieldMask) {
	if f <= _MaxFieldIDHead {
		ret = self.head[f]
	} else {
		ret = self.tail[f]
	}
	if ret.Exist() {
		return ret
	}
	return nil
}

// setFieldID ensure a fieldmask slot for f
func (self *FieldMask) setFieldID(f fieldID, ft FieldMaskType) *FieldMask {
	if self.fdMask == nil {
		// println("new fdmask")
		m := makeFieldMaskMap()
		self.fdMask = &m
	}
	return self.fdMask.SetIfNotExist(fieldID(f), ft, self.isBlack)
}

// type fieldMaskBitmap []byte

// const _BucketBit = 8

// func (self *fieldMaskBitmap) Set(f fieldID) {
// 	b := int(f / _BucketBit)
// 	i := int(f % _BucketBit)
// 	c := cap(*self)
// 	if c <= b+1 {
// 		tmp := make([]byte, len(*self), (c + b + 1))
// 		copy(tmp, *self)
// 		*self = tmp
// 	}
// 	if len(*self) <= b {
// 		*self = (*self)[:b+1]
// 	}
// 	(*self)[b] |= byte(1 << i)
// }

// func (self *fieldMaskBitmap) Get(f fieldID) bool {
// 	b := int(f / _BucketBit)
// 	if len(*self) <= b {
// 		return false
// 	}
// 	i := int(f % _BucketBit)
// 	return ((*self)[b] & byte(1<<i)) != 0
// }

func (self *FieldMask) setInt(v int, ft FieldMaskType, cap int) *FieldMask {
	if self.intMask == nil {
		// println("new intMask")
		self.intMask = make(intMap, cap)
	}
	return self.intMask.SetIfNotExist(v, ft, self.isBlack)
}

type intMap map[int]*FieldMask

func (im intMap) Reset() {
	for _, v := range im {
		v.reset()
	}
}

func (im intMap) Get(i int) (ret *FieldMask) {
	ret = im[i]
	if ret.Exist() {
		return ret
	}
	return nil
}

func (im intMap) SetIfNotExist(i int, ft FieldMaskType, black bool) *FieldMask {
	s := im[i]
	if s == nil {
		fm := newFieldMask(ft, black)
		im[i] = &fm
		return &fm
	}
	if s.typ == 0 {
		s.assign(ft, black)
	}
	return s
}

func (im intMap) Unset(i int) {
	delete(im, i)
}

func (self *FieldMask) setStr(v string, ft FieldMaskType, cap int) *FieldMask {
	if self.strMask == nil {
		// println("new setStr")
		self.strMask = make(strMap, cap)
	}
	return self.strMask.SetIfNotExist(v, ft, self.isBlack)
}

type strMap map[string]*FieldMask

func (sm strMap) Reset() {
	for _, v := range sm {
		v.reset()
	}
}

func (im strMap) Get(i string) (ret *FieldMask) {
	ret = im[i]
	if ret.Exist() {
		return ret
	}
	return nil
}

func (im strMap) SetIfNotExist(i string, ft FieldMaskType, black bool) *FieldMask {
	s := im[i]
	if s == nil {
		fm := newFieldMask(ft, black)
		im[i] = &fm
		return &fm
	}
	if s.typ == 0 {
		s.assign(ft, black)
	}
	return s
}

func (im strMap) Unset(i string) {
	delete(im, i)
}

func (self *FieldMask) setAll(ft FieldMaskType) *FieldMask {
	if self.all == nil {
		fm := newFieldMask(ft, self.isBlack)
		self.all = &fm
	} else if self.all.typ == 0 {
		self.all.assign(ft, self.isBlack)
	}
	return self.all
}

func newFieldMask(ft FieldMaskType, black bool) FieldMask {
	return FieldMask{
		typ:     ft,
		isAll:   false,
		isBlack: black,
	}
}

func (self *FieldMask) assign(ft FieldMaskType, black bool) {
	self.typ = ft
	self.isAll = false
	self.isBlack = black
}
