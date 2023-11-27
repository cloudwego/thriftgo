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
	"errors"
	"strconv"
	"sync"
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
	write(buf, self.typ.String())
	write(buf, `"`)
	return self.marshalRec(buf)
}

func (self *FieldMask) marshalRec(buf *[]byte) error {
	if self.typ == FtScalar {
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
		write(buf, f.typ.String())
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

	} else if self.typ == ftStruct {
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

	} else if self.typ == FtList || self.typ == ftIntMap {
		for k, f := range self.intMask {
			cont, err := writer(strconv.Itoa(k), f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == ftStrMap {
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

// func (self *FieldMask)
