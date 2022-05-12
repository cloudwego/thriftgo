// Copyright 2022 CloudWeGo Authors
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

include "reflect.thrift"
include "context.thrift"
include "thrift.thrift"
include "strings.thrift"
include "fmt.thrift"

struct X {
    1: reflect.X X1
    2: context.X X2
    3: thrift.X X3
    4: strings.X X4
    5: fmt.X X5
}

service XXX {
    X Method(1: X req)
}

