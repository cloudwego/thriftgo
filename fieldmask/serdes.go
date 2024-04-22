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
	"bytes"
	"encoding/json"
	"errors"
	"sort"
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

// MarshalJSON marshals the fieldmask into json.
//
// For example:
//   - pathes `[]string{"$.Extra[0].List", "$.Extra[*].Set", "$.Meta.F2{0}", "$.Meta.F2{*}.Addr"}` will produces:
//   - `{"path":"$","type":"Struct","children":[{"path":6,"type":"List","children":[{"path":"*","type":"Struct","children":[{"path":4,"type":"List"}]}]},{"path":256,"type":"Struct","children":[{"path":2,"type":"IntMap","children":[{"path":"*","type":"Struct","children":[{"path":0,"type":"Scalar"}]}]}]}]}`
//
// For details:
//   - `path` is the path segment of current fieldmask layer
//   - `type` is the `FieldMaskType` of the fieldmask
//     -`children` is the chidlren of subsequent pathes
//   - each fieldmask always starts with root path "$"
//   - path "*" indicates all subsequent path of the fieldmask shares the same sub fieldmask
func (fm *FieldMask) MarshalJSON() ([]byte, error) {
	if fm == nil {
		return []byte("null"), nil
	}
	buf := bytesPool.Get().(*[]byte)

	err := fm.marshalBegin(buf)
	if err != nil {
		(*buf) = (*buf)[:0]
		bytesPool.Put(buf)
		return nil, err
	}

	ret := make([]byte, len(*buf))
	copy(ret, *buf)
	(*buf) = (*buf)[:0]
	bytesPool.Put(buf)
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
	write(buf, `","is_black":`)
	write(buf, strconv.FormatBool(self.isBlack))
	return self.marshalRec(buf)
}

type ivalue struct {
	id int
	fm *FieldMask
}

type isorter []ivalue

func (self isorter) Len() int {
	return len(self)
}

func (self isorter) Less(i, j int) bool {
	return self[i].id < self[j].id
}

func (self isorter) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type svalue struct {
	id string
	fm *FieldMask
}

type ssorter []svalue

func (self ssorter) Len() int {
	return len(self)
}

func (self ssorter) Less(i, j int) bool {
	return self[i].id < self[j].id
}

func (self ssorter) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *FieldMask) marshalRec(buf *[]byte) error {
	if self.All() && self.all == nil {
		write(buf, "}")
		return nil
	}

	var start bool
	writer := func(path json.RawMessage, f *FieldMask) (bool, error) {
		if !f.Exist() {
			return true, nil
		}
		if start {
			write(buf, `,`)
		}

		// write path
		write(buf, `{"path":`)
		write(buf, string(path))
		write(buf, `,`)

		// write type
		write(buf, `"type":"`)
		typ, _ := f.typ.MarshalText()
		*buf = append(*buf, typ...)

		// write is_black
		write(buf, `","is_black":`)
		write(buf, strconv.FormatBool(f.isBlack))

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
		fds := make(isorter, 0, len(self.fdMask.tail)*2)
		for id, f := range self.fdMask.head {
			if !f.Exist() {
				continue
			}
			fds = append(fds, ivalue{id, f})
		}
		for id, f := range self.fdMask.tail {
			if !f.Exist() {
				continue
			}
			fds = append(fds, ivalue{int(id), f})
		}
		sort.Stable(fds)
		for _, v := range fds {
			cont, err := writer(json.RawMessage(strconv.Itoa(v.id)), v.fm)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == FtList || self.typ == FtIntMap {
		fds := make(isorter, 0, len(self.intMask))
		for k, f := range self.intMask {
			if !f.Exist() {
				continue
			}
			fds = append(fds, ivalue{int(k), f})
		}
		sort.Stable(fds)
		for _, v := range fds {
			cont, err := writer(json.RawMessage(strconv.Itoa(v.id)), v.fm)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == FtStrMap {
		fds := make(ssorter, 0, len(self.strMask))
		for k, f := range self.strMask {
			if !f.Exist() {
				continue
			}
			fds = append(fds, svalue{k, f})
		}
		sort.Stable(fds)
		for _, v := range fds {
			cont, err := writer(json.RawMessage(strconv.Quote(v.id)), v.fm)
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

// FieldMaskTransfer is the data struct being used to transfer and construct a fieldmask
type FieldMaskTransfer struct {
	Path     json.RawMessage     `json:"path"` // NOTICE: must be float64 or string
	Type     FieldMaskType       `json:"type"`
	IsBlack  bool                `json:"is_black"`
	Children []FieldMaskTransfer `json:"children"`
}

// UnmarshalJSON unmarshal the fieldmask from json.
//
//	Input JSON **MUST** be according to the schema of `FieldMask.MarshalJSON()`
func (self *FieldMask) UnmarshalJSON(in []byte) error {
	if self == nil {
		return errors.New("nil memory address")
	}
	s := new(FieldMaskTransfer)
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
	if s == nil {
		self = nil
		return nil
	}
	// spew.Dump(s)
	if !bytes.Equal(s.Path, jsonPathRoot) {
		return errors.New("fieldmask must begin with root path '$'")
	}
	return self.TransferFrom(s)
}

// TransferTo transfer FieldMaskTransfer to a FieldMask
func (self *FieldMaskTransfer) TransferTo() (*FieldMask, error) {
	fm := new(FieldMask)
	err := fm.TransferFrom(self)
	return fm, err
}

// TransferFrom transfroms a FieldMaskTransfer to the FieldMask
func (self *FieldMask) TransferFrom(s *FieldMaskTransfer) error {
	if s == nil || s.Type == FtInvalid {
		return errors.New("invalid fieldmask type")
	}
	self.typ = s.Type
	self.isBlack = s.IsBlack

	if len(s.Children) == 0 {
		self.isAll = true
		return nil
	}

	if s.Type == FtScalar {
		is, err := self.checkAll(&s.Children[0])
		if err != nil {
			return err
		}
		if !is {
			return errors.New("expect * for the child")
		}
		return nil
	} else if s.Type == FtStruct {
		for _, n := range s.Children {
			if is, err := self.checkAll(&n); err != nil {
				return err
			} else if is {
				return nil
			}
			var id fieldID
			if err := json.Unmarshal(n.Path, &id); err != nil {
				return err
			}
			next := self.setFieldID(id, n.Type)
			if err := next.TransferFrom(&n); err != nil {
				return err
			}
		}
	} else if s.Type == FtList || s.Type == FtIntMap {
		for _, n := range s.Children {
			if is, err := self.checkAll(&n); err != nil {
				return err
			} else if is {
				return nil
			}
			var id int
			if err := json.Unmarshal(n.Path, &id); err != nil {
				return err
			}
			next := self.setInt(int(id), n.Type, len(s.Children))
			if err := next.TransferFrom(&n); err != nil {
				return err
			}
		}
	} else if s.Type == FtStrMap {
		for _, n := range s.Children {
			if is, err := self.checkAll(&n); err != nil {
				return err
			} else if is {
				return nil
			}
			var id string
			if err := json.Unmarshal(n.Path, &id); err != nil {
				return err
			}
			next := self.setStr(id, n.Type, len(s.Children))
			if err := next.TransferFrom(&n); err != nil {
				return err
			}
		}
	}

	return nil
}

func (self *FieldMask) checkAll(s *FieldMaskTransfer) (bool, error) {
	if bytes.Equal(s.Path, jsonPathAny) {
		self.isAll = true
		self.all = &FieldMask{}
		return true, self.all.TransferFrom(s)
	}
	return false, nil
}

var (
	fm2json sync.Map
	json2fm sync.Map
)

// Marshal serializes a fieldmask into bytes.
//
// Notice: This API uses cache to accelerate processing,
// at the cost of increasing memory usage
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

// Marshal deserializes a fieldmask from bytes.
//
// Notice: This API uses cache to accelerate processing,
// at the cost of increasing memory usage
func Unmarshal(data []byte) (*FieldMask, error) {
	// fast-path: load from cache
	if fm, ok := json2fm.Load(utils.B2S(data)); ok {
		return fm.(*FieldMask), nil
	}
	// slow-path: unmarshal from json
	fm := new(FieldMask)
	err := fm.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}
	json2fm.Store(string(data), fm)
	return fm, nil
}
