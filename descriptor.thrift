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

namespace go thrift_reflection

struct TypeDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:optional TypeDescriptor key_type   // key type for map container
    4:optional TypeDescriptor value_type // value type for map container, or the type for list and set
    5:optional map<string,string> extra  // extra info
}

struct ConstDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:required TypeDescriptor type
    4:required ConstValueDescriptor value
    5:required map<string,list<string>> annotations
    6:required string comments
    7:optional map<string,string> extra  // extra info
}

enum ConstValueType{
    DOUBLE
    INT
    STRING
    BOOL
    LIST
    MAP
    IDENTIFIER
}

struct ConstValueDescriptor{
    1:required ConstValueType type
    2:required double value_double // for double
    3:required i64 value_int   // for i8 i16 i32 i64 byte binary
    4:required string value_string // for string
    5:required bool value_bool // for bool
    6:optional list<ConstValueDescriptor> value_list // for list set
    7:optional map<ConstValueDescriptor,ConstValueDescriptor> value_map // for map
    8:required string value_identifier // for identifier, such as another constant's name
}

struct TypedefDescriptor{
    1:required string filepath // the name of idl file
    2:required TypeDescriptor type
    3:required string alias
    4:required map<string,list<string>> annotations
    5:required string comments
    6:optional map<string,string> extra // extra
}

struct EnumDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:required list<EnumValueDescriptor> values
    4:required map<string,list<string>> annotations
    5:required string comments  
    6:optional map<string,string> extra  // extra info
}

struct EnumValueDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:required i64 value
    4:required map<string,list<string>> annotations
    5:required string comments  
    6:optional map<string,string> extra  // extra info
}

struct FieldDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:required TypeDescriptor type
    4:required string requiredness  // required, optional, or default
    5:required i32 id // field id
    6:optional ConstValueDescriptor default_value
    7:required map<string,list<string>> annotations
    8:required string comments
    9:optional map<string,string> extra  // extra info
}

struct StructDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:required list<FieldDescriptor> fields
    4:required map<string,list<string>> annotations
    5:required string comments  
    6:optional map<string,string> extra  // extra info
}

struct MethodDescriptor{
    1:required string filepath // the name of idl file
    2:required string name
    3:optional TypeDescriptor response // response, if it's oneway method, this should be nil
    4:required list<FieldDescriptor> args
    5:required map<string,list<string>> annotations
    6:required string comments  
    7:required list<FieldDescriptor> throw_exceptions
    8:required bool is_oneway
    9:optional map<string,string> extra  // extra info
}

struct ServiceDescriptor{
    1:required string filepath // the name of idl file
    2:required string name 
    3:required list<MethodDescriptor> methods  
    4:required map<string,list<string>> annotations 
    5:required string comments  
    6:optional map<string,string> extra  // extra info
}

struct FileDescriptor{
    1:required string filepath // the path of idl file, eg: xx/idl/entity.thrift
    2:required map<string,string> includes  // include IDL, key is alias and value is filepath, eg: entity -> xx/idl/entity.thrift
    3:required map<string,string> namespaces // namespace, key is language and value is namespace for this language, eg: go -> xxx ; java -> xxx
    4:required list<ServiceDescriptor> services
    5:required list<StructDescriptor> structs
    6:required list<StructDescriptor> exceptions
    7:required list<EnumDescriptor> enums
    8:required list<TypedefDescriptor> typedefs
    9:required list<StructDescriptor> unions
    10:required list<ConstDescriptor> consts
    11:optional map<string,string> extra  // extra info
}