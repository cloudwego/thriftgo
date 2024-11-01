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
  1: bool B0;
  2: required bool B1;
  3: optional bool B2;
  4: optional bool B3 = true;

  11: byte Byte0;
  12: required byte Byte1;
  13: optional byte Byte2;
  14: optional byte Byte3= 1;

  21: i8 I800;
  22: required i8 I801;
  23: optional i8 I802;
  24: optional i8 I803 = 2;

  31: i16 I160;
  32: required i16 I161;
  33: optional i16 I162;
  34: optional i16 I163= 3;

  41: i32 I320;
  42: required i32 I321;
  43: optional i32 I322;
  44: optional i32 I323 = 4;

  51: double Dbl0;
  52: required double Dbl1;
  53: optional double Dbl2;
  54: optional double Dbl3= 5;

  61: string Str0;
  62: required string Str1;
  63: optional string Str2;
  64: optional string Str3 = "6";

  71: binary Bin0;
  72: required binary Bin1;
  73: optional binary Bin2;
  74: optional binary Bin3 = "7";

  81: Numberz Num0;
  82: required Numberz Num1;
  83: optional Numberz Num2;
  84: optional Numberz Num3 = 10;

  91: UserID UID0;
  92: required UserID UID1;
  93: optional UserID UID2;
  94: optional UserID UID3 = 9

  101: Msg Msg0;
  102: required Msg Msg1;
  103: optional Msg Msg2;

  111: map<i32, string> Map111;
  112: required map<i32, string> Map112;
  113: optional map<i32, string> Map113;

  121: map<i32, i32> Map121;
  122: required map<i32, i32> Map122;
  123: optional map<i32, i32> Map123;

  131: map<string, Msg> Map131;
  132: required map<string, Msg> Map132;
  133: optional map<string, Msg> Map133;

  141: list<i32> List141;
  142: required list<i32> List142;
  143: optional list<i32> List143;

  151: list<string> List151;
  152: required list<string> List152;
  153: optional list<string> List153;

  161: list<Msg> List161;
  162: required list<Msg> List162;
  163: optional list<Msg> List163;

  171: set<i32> Set171;
  172: required set<i32> Set172;
  173: optional set<i32> Set173;

  181: set<string> Set181;
  182: required set<string> Set182;
  183: optional set<string> Set183;

  191: list<map<i32, i32>> Mix191;
  192: required list<map<i32, i32>> Mix192;
  193: optional list<map<i32, i32>> Mix193;

  201: map<i32, list<i32>> Mix201;
  202: required map<i32, list<i32>> Mix202;
  203: optional map<i32, list<i32>> Mix203;
}
