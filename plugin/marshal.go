// Copyright 2021 CloudWeGo Authors
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

package plugin

import "github.com/cloudwego/thriftgo/generator/golang/extension/meta"

// MarshalRequest encodes a request with binary protocol.
func MarshalRequest(req *Request) ([]byte, error) {
	return meta.Marshal(req)
}

// UnmarshalRequest decodes a request with binary protocol.
func UnmarshalRequest(bs []byte) (*Request, error) {
	req := NewRequest()
	if err := meta.Unmarshal(bs, req); err != nil {
		return nil, err
	}
	return req, nil
}

// MarshalResponse encodes a response with binary protocol.
func MarshalResponse(res *Response) ([]byte, error) {
	return meta.Marshal(res)
}

// UnmarshalResponse decodes a response with binary protocol.
func UnmarshalResponse(bs []byte) (*Response, error) {
	res := NewResponse()
	if err := meta.Unmarshal(bs, res); err != nil {
		return nil, err
	}
	return res, nil
}
