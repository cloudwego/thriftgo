// Copyright 2024 CloudWeGo Authors
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

namespace go multiple.include.common

include "../common1.thrift"
include "../common2.thrift"

struct Test3Struct1 {
    1: required common1.Common1Enum1 enumField1
    2: required common1.Common1Struct1 structField1
    3: required common2.Common2Enum1 enumField2
    4: required common2.Common2Struct1 structField2
}

service Test3 {
    string Process(1: Test3Struct1 req)
}