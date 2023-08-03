// Copyright 2023 CloudWeGo Authors
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

package thrift_reflection

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"

	"github.com/cloudwego/thriftgo/generator/golang/extension/meta"
)

func (fd *FileDescriptor) Marshal() ([]byte, error) {
	bytes, err := meta.Marshal(fd)
	if err != nil {
		return nil, err
	}
	return doGzip(bytes)
}

func Unmarshal(bytes []byte) (*FileDescriptor, error) {
	bytes, err := doUnzip(bytes)
	if err != nil {
		return nil, err
	}
	fd := NewFileDescriptor()
	if err = meta.Unmarshal(bytes, fd); err != nil {
		return nil, err
	}
	return fd, nil
}

func MustUnmarshal(bytes []byte) *FileDescriptor {
	bytes, err := doUnzip(bytes)
	if err != nil {
	}
	fd := NewFileDescriptor()
	if err = meta.Unmarshal(bytes, fd); err != nil {
		panic(err)
	}
	return fd
}

func doGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return ioutil.ReadAll(&buffer)
}

func doUnzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(data)
	reader, err := gzip.NewReader(&buffer)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}
