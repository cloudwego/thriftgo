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

namespace go option_gen.annotation.entity
include "entity_struct.thrift"

struct _StructOptions {
      1:required PersonBasicInfo person_basic_info
      2:required PersonStructInfo person_struct_info
      3:required PersonContainerInfo person_container_info
}

struct _FieldOptions {
      1:required string person_field_info
}

struct PersonBasicInfo{
    1:required i8 valuei8
    2:required i16 valuei16;
    3:required i32 valuei32;
    4:required i64 valuei64;
    5:required string valuestring;
    6:required byte valuebyte;
    7:required double valuedouble;
    8:required binary valuebinary;
    9:required bool valuebool;
}

struct PersonContainerInfo{
    1:required map<string,string> valuemap;
    2:required list<string> valuelist;
    3:required set<string> valueset;
    4:required list<set<string>> valuelistset;
    5:required list<set<entity_struct.InnerStruct>> valuelistsetstruct;
    6:required map<string,entity_struct.InnerStruct> valuemapstruct;
}

struct PersonStructInfo{
    1:required TestStruct valueteststruct
    2:required entity_struct.InnerStruct valuestruct
    3:required TestEnum valueenum
    4:required TestStructTypedef valuestructtypedef
    5:required TestBasicTypedef valuebasictypedef
}

enum TestEnum{
    A
    B
    C
}

typedef entity_struct.InnerStruct TestStructTypedef
typedef string TestBasicTypedef

struct TestStruct{
    1:required string name
    2:required string age
    3:required entity_struct.InnerStruct innerStruct
}