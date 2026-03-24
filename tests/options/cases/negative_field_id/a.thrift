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

struct X {
    -1: i32 a
    -2: string b
    -3: binary c
}

service S {
    void m0(-1: i32 r)
    void m1(-1: i32 r1, -2: double r2)
    void m2(-1: string r)
    void m3(-1: binary r)
    void m4(-1: X r)
}


