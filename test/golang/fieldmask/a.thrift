#! /bin/bash -e

# Copyright 2022 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

namespace go base

struct TrafficEnv {
	0: string Name = "",
	1: bool Open = false,
	2: string Env = "",
	256: i64 Code,
}

struct Base {
	0: string Addr = "",
	1: string LogID = "",
	2: string Caller = "",
	5: optional TrafficEnv TrafficEnv,
	255: optional ExtraInfo Extra,
	256: MetaInfo Meta,
}

struct ExtraInfo {
	1: map<string, string> F1
	2: map<i64, string> F2,
	3: list<string> F3
	4: set<string> F4,
	5: map<double, Val> F5
	6: map<Int, Key> F6
	7: map<Int, map<Int, Key>> F7
	8: map<Int, list<Key>> F8
	9: map<Int, list<map<Int, Key>>> F9
	10: map<Val, Key> F10
}

struct MetaInfo {
	1: map<Int, Val> IntMap,
	2: map<Str, Key> StrMap,
	3: list<Key> List,
	4: set<Val> Set,
	11: map<Int, list<Str>> MapList
	12: list<map<Int, list<Str>>> ListMapList
	255: Base Base,
}

typedef Val Key 

struct Val {
	1: string id
}

typedef double Float

typedef i64 Int

typedef string Str

enum Ex {
	A = 1,
	B = 2,
	C = 3
}

struct BaseResp {
	1: string StatusMessage = "",
	2: i32 StatusCode = 0,
	3: optional map<string, string> Extra,

	4: map<Str, Str> F1
	5: map<Int, string> F2,
	6: list<string> F3
	7: set<string> F4,
	8: map<Float, Val> F5
	9: map<double, string> F6
	10: map<Ex, string> F7
	11: map<double, list<Str>> F8
	12: list<map<Float, list<Str>>> F9
}
