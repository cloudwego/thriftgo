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

	err := fm.marshal(buf)
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

func (self *FieldMask) marshal(buf *[]byte) error {
	if self == nil {
		write(buf, "{}")
		return nil
	}

	// write path
	if self.path == "" {
		return errors.New("unknown path for fieldmask")
	}
	write(buf, `{"path":`)
	write(buf, self.path)
	write(buf, ",")

	// write type
	write(buf, `"type":`)
	write(buf, self.typ.String())
	if self.typ == ftScalar {
		write(buf, "}")
		return nil
	}
	write(buf, ",")

	// write children
	write(buf, `"children":[`)
	var start bool
	var writer = func(f *FieldMask) (bool, error) {
		if !f.Exist() {
			return true, nil
		}
		if start {
			write(buf, ",")
		}
		if err := f.marshal(buf); err != nil {
			return false, err
		}
		start = true
		return true, nil
	}

	if self.All() {
		_, err := writer(self.all)
		if err != nil {
			return err
		}
	} else if self.typ == ftStruct {
		for _, f := range self.fdMask.head {
			cont, err := writer(f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}
		for _, f := range self.fdMask.tail {
			cont, err := writer(f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == ftArray || self.typ == ftIntMap {
		for _, f := range self.intMask {
			cont, err := writer(f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}

	} else if self.typ == ftStrMap {
		for _, f := range self.strMask {
			cont, err := writer(f)
			if err != nil {
				return err
			}
			if !cont {
				break
			}
		}
	}

	write(buf, "]}")
	return nil
}

// func (self *FieldMask)
