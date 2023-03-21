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

package reflection

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudwego/thriftgo/parser"
)

type FileDescriptor struct {
	Filename   string               `thrift:"Filename,1" json:"Filename"`
	IncludeMap map[string]string    `thrift:"IncludeMap,2" json:"Include"`
	Typedefs   []*parser.Typedef    `thrift:"Typedefs,5" json:"Typedefs"`
	Constants  []*parser.Constant   `thrift:"Constants,6" json:"Constants"`
	Enums      []*parser.Enum       `thrift:"Enums,7" json:"Enums"`
	Structs    []*parser.StructLike `thrift:"Structs,8" json:"Structs"`
	Unions     []*parser.StructLike `thrift:"Unions,9" json:"Unions"`
	Exceptions []*parser.StructLike `thrift:"Exceptions,10" json:"Exceptions"`
	Services   []*parser.Service    `thrift:"Services,11" json:"Services"`
}

func JsonEncode(f *FileDescriptor) ([]byte, error) {
	data, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func JsonDecode(data []byte) (*FileDescriptor, error) {
	f := &FileDescriptor{}
	err := json.Unmarshal(data, f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func Encode(p *parser.Thrift) string {
	// alias prefix -> path
	includeMap := map[string]string{}
	for _, inc := range p.Includes {
		arr := strings.Split(strings.TrimSuffix(inc.Path, ".thrift"), "/")
		alias := arr[len(arr)-1]
		includeMap[alias] = inc.Path
	}
	f := &FileDescriptor{
		Filename:   p.Filename,
		IncludeMap: includeMap,
		Typedefs:   p.Typedefs,
		Constants:  p.Constants,
		Enums:      p.Enums,
		Structs:    p.Structs,
		Unions:     p.Unions,
		Exceptions: p.Exceptions,
		Services:   p.Services,
	}
	bytes, _ := JsonEncode(f)
	bytes, _ = doGzip(bytes)
	byteStr := fmt.Sprintf("% x", bytes)
	byteStr = strings.ReplaceAll(byteStr, " ", ",0x")
	byteStr = "[]byte" + "{\n0x" + byteStr + "}"
	return byteStr
}

func Decode(data []byte) *FileDescriptor {
	data, _ = unGzip(data)
	f, _ := JsonDecode(data)
	return f
}

func doGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	compressedData, err := ioutil.ReadAll(&buffer)
	return compressedData, nil
}

func unGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(data)
	reader, _ := gzip.NewReader(&buffer)
	defer reader.Close()
	undatas, _ := ioutil.ReadAll(reader)
	return undatas, nil
}
