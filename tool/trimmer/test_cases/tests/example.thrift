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

namespace go hello.cloudwego.team

include "base.thrift"

struct MyReq{
    1:required string name,
    2:required string id,
    255: optional base.Base base
}

struct MyResp{
    1:required string text
    2:required base.BaseResp baseResp
    3:required list<MyResp> myselfs
}

service greet {
	MyResp Hello(1:required MyReq aareqaaa)
}