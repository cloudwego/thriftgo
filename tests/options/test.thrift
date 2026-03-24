// Copyright 2025 CloudWeGo Authors
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

namespace go tests

// Typedefs for alias testing
typedef bool AliasBool
typedef i32 AliasI32
typedef string AliasString
typedef Enum AliasEnum

// Enum for enum-related options
enum Enum {
    Unknown = 0,
    Value1 = 1,
    Value2 = 2,
    Value3 = 3,
}

enum Color {
    Red = 1,
    Green = 2,
    Blue = 3,
}

// Exception for deep_equal / setter coverage
exception MyException {
    1: i32 code
    2: string message
}

// Simple struct (no self-reference, safe for raw_struct template)
struct Inner {
    1: required string Key
    2: optional string Value
    3: i32 Num
}

// Struct with various field types to exercise all options
struct MyStruct {
    // Required field
    1: required string Name
    // Optional field (omitempty_for_optional)
    2: optional string Nickname
    // Default field
    3: i32 Age
    // Enum field
    4: Enum Status
    // Container fields (value_type_in_container, validate_set)
    5: list<Inner> Children
    6: map<string, Inner> Props
    7: set<string> Tags
    // Binary
    8: binary Data
    // Bool field
    9: bool Active
    // Double
    10: double Score
    // Map with enum key
    11: map<Enum, string> EnumMap
    // Set of i32 (validate_set)
    12: set<i32> Numbers
    // Optional struct (nil_safe)
    13: optional Inner Extra
}

// Struct with annotations for nested struct and field mask
struct AnnotatedStruct {
    1: string ID (go.tag = 'json:"id" yaml:"id"')
    2: Inner Nested (thrift.nested = "true")
    3: optional string Label
}

// Service for processor / streaming coverage
service MyService {
    MyStruct GetStruct(1: string id, 2: i32 limit)
    void UpdateStruct(1: MyStruct s)
    list<MyStruct> ListStructs(1: set<string> ids)
    Enum GetStatus(1: string id)
    oneway void Ping()
}
