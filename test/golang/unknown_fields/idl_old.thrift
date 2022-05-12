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


namespace * unknown

struct Empty {
}

struct Struct {
    1: bool Bool
    2: required byte Byte
    3: optional i16 I16

    100: optional string Str
}

union Union {
    1: i32 I32
    2: double Double
    255: string Str
}

exception Exception {
    1: string Str
}

struct Merged {
    1: required Struct s
    2: required Union u
    3: required Exception e
}

