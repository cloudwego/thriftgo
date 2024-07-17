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

namespace go tests.extend.common

struct Common3Struct1 {
    1: required string stringField
}

struct Commmon3Struct2 {
    1: required string stringField
}

service Common3 {
    string ProcessCommon3(1: Common3Struct1 req)
    string Echo(1: Commmon3Struct2 req)
}