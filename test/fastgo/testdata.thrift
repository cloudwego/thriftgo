# Copyright 2024 CloudWeGo Authors
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

namespace go testdata

enum Numberz
{
  TEN = 10
}

typedef i64 UserID

struct Msg
{
  1: string message;
  2: i32 type;
}

struct TestTypes {
  1: required bool FBool;
  2: byte FByte;
  3: i8 I8;
  4: i16 I16;
  5: i32 I32;
  6: i64 I64;
  7: double Double;
  8: string String;
  9: binary Binary;
  10: Numberz Enum;
  11: UserID UID;
  12: Msg S;
  20: required map<i32, i32> M0;
  21: map<i32, string> M1;
  22: map<i32, Msg> M2;
  23: map<string, Msg> M3;
  30: required list<i32> L0;
  31: list<string> L1;
  32: list<Msg> L2;
  40: required set<i32> S0;
  41: set<string> S1;
  50: list<map<i32, i32>> LM;
  60: map<i32, list<i32>> ML;
}

struct TestTypesOptional {
  1: optional bool FBool;
  2: optional byte FByte;
  3: optional i8 I8;
  4: optional i16 I16;
  5: optional i32 I32;
  6: optional i64 I64;
  7: optional double Double;
  8: optional string String;
  9: optional binary Binary;
  10: optional Numberz Enum;
  11: optional UserID UID;
  12: optional Msg S;
  20: optional map<i32, i32> M0;
  21: optional map<i32, string> M1;
  22: optional map<i32, Msg> M2;
  23: optional map<string, Msg> M3;
  30: optional list<i32> L0;
  31: optional list<string> L1;
  32: optional list<Msg> L2;
  40: optional set<i32> S0;
  41: optional set<string> S1;
  50: optional list<map<i32, i32>> LM;
  60: optional map<i32, list<i32>> ML;
}

struct TestTypesWithDefault {
  1: optional bool FBool = true;
  2: optional byte FByte = 2;
  3: optional i8 I8 = 3;
  4: optional i16 I6 = 4;
  5: optional i32 I32 = 5;
  6: optional i64 I64 = 6;
  7: optional double Double = 7;
  8: optional string String = "8";
  9: optional binary Binary = "8";
  10: optional Numberz Enum = 10;
  11: optional UserID UID = 11;
  30: optional list<i32> L0 = [ 30 ];
  40: optional set<i32> S0 = [ 40 ];
}
