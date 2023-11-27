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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/cloudwego/thriftgo/internal/utils"
)

var bytesPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 4096)
		return &b
	},
}

func (fm *FieldMask) MarshalJSON() ([]byte, error) {
	buf := bytesPool.Get().(*[]byte)

	err := fm.marshalBegin(buf)
	if err != nil {
		(*buf) = (*buf)[:0]
		bytesPool.Put(buf)
		return nil, err
	}

	ret := make([]byte, len(*buf))
	copy(ret, *buf)
	return ret, nil
}

func write(buf *[]byte, str string) {
	*buf = append(*buf, str...)
}

func (self *FieldMask) marshalBegin(buf *[]byte) error {
	if self == nil {
		write(buf, "{}")
		return nil
	}
	write(buf, `{"path":"$","type":"`)
	out, _ := self.typ.MarshalText()
	*buf = append(*buf, out...)
	write(buf, `"`)
	return self.marshalRec(buf)
}

func (self *FieldMask) marshalRec(buf *[]byte) error {
	if self.typ == FtScalar || (self.isAll && self.all == nil) {
		write(buf, "}")
		return nil
	}

	var start bool
	var writer = func(path string, f *FieldMask) (bool, error) {
		if !f.Exist() {
			return true, nil
		}
		if start {
			write(buf, ",")
		}

		// write path
		write(buf, `{"path":`)
		write(buf, path)
		write(buf, ",")

		// write type
		write(buf, `"type":"`)
		typ, _ := f.typ.MarshalText()
		*buf = append(*buf, typ...)
		write(buf, `"`)

		if err := f.marshalRec(buf); err != nil {
			return false, err
		}

		start = true
		return true, nil
	}

	// write children
	write(buf, `,"children":[`)

	if self.All() {
		_, err := writer(jsonPathAny, self.all)
		if err != nil {
			return err
		}

	} else if self.typ == FtStruct {
		for id, f := range self.fdMask.head {
			cont, err := writer(strconv.Itoa(id), f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}
		for id, f := range self.fdMask.tail {
			cont, err := writer(strconv.Itoa(int(id)), f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == FtList || self.typ == FtIntMap {
		for k, f := range self.intMask {
			cont, err := writer(strconv.Itoa(k), f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == FtStrMap {
		for k, f := range self.strMask {
			cont, err := writer(strconv.Quote(k), f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}
	} else {
		return errors.New("invalid fieldmask type")
	}

	write(buf, "]}")
	return nil
}

type shadowFieldMask struct {
	Path    interface{}       `json:"path"`
	Type    FieldMaskType     `json:"type"`
	Chilren []shadowFieldMask `json:"children"`
}

func (self *FieldMask) UnmarshalJSON(in []byte) error {
	if self == nil {
		return errors.New("nil memory address")
	}
	var s = new(shadowFieldMask)
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
	// spew.Dump(s)
	if s.Path != jsonPathRoot {
		return errors.New("fieldmask must begin with root path '$'")
	}
	return self.fromShadow(s)
}

func (self *FieldMask) fromShadow(s *shadowFieldMask) error {
	if s == nil || s.Type == FtInvalid {
		return errors.New("invalid fieldmask type")
	}
	self.typ = s.Type

	if len(s.Chilren) == 0 {
		self.isAll = true
		return nil
	}

	if s.Type == FtStruct {
		for _, n := range s.Chilren {
			if is, err := self.checkAll(&n); err != nil {
				return err
			} else if is {
				return nil
			}
			id, ok := n.Path.(float64)
			if !ok {
				return fmt.Errorf("expect number but got %#v", n.Path)
			}
			if self.fdMask == nil {
			}
			next := self.setFieldID(fieldID(id), n.Type, len(s.Chilren))
			if err := next.fromShadow(&n); err != nil {
				return err
			}
		}

	} else if s.Type == FtList || s.Type == FtIntMap {
		for _, n := range s.Chilren {
			if is, err := self.checkAll(&n); err != nil {
				return err
			} else if is {
				return nil
			}
			id, ok := n.Path.(float64)
			if !ok {
				return fmt.Errorf("expect number but got %#v", n.Path)
			}
			next := self.setInt(int(id), n.Type, len(s.Chilren))
			if err := next.fromShadow(&n); err != nil {
				return err
			}
		}

	} else if s.Type == FtStrMap {
		for _, n := range s.Chilren {
			if is, err := self.checkAll(&n); err != nil {
				return err
			} else if is {
				return nil
			}
			id, ok := n.Path.(string)
			if !ok {
				return fmt.Errorf("expect string but got %#v", n.Path)
			}
			next := self.setStr(id, n.Type, len(s.Chilren))
			if err := next.fromShadow(&n); err != nil {
				return err
			}
		}
	}

	return nil
}

func (self *FieldMask) checkAll(s *shadowFieldMask) (bool, error) {
	if s.Path == "*" {
		self.isAll = true
		self.all = &FieldMask{}
		return true, self.all.fromShadow(s)
	}
	return false, nil
}

var (
	fm2json sync.Map
	json2fm sync.Map
)

func Marshal(fm *FieldMask) ([]byte, error) {
	// fast-path: load from cache
	if j, ok := fm2json.Load(fm); ok {
		return j.([]byte), nil
	}
	// slow-path: marshal from object
	nj, err := fm.MarshalJSON()
	if err != nil {
		return nil, err
	}
	fm2json.Store(fm, nj)
	return nj, nil
}

func Unmarshal(data []byte) (*FieldMask, error) {
	// fast-path: load from cache
	sd := utils.B2S(data)
	if fm, ok := json2fm.Load(sd); ok {
		return fm.(*FieldMask), nil
	}
	// slow-path: unmarshal from json
	var fm = new(FieldMask)
	err := fm.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}
	json2fm.Store(string(data), fm)
	return fm, nil
}
