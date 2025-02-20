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

namespace * parser

typedef list<Annotation> Annotations

typedef list<StructuredAnnotation> StructuredAnnotations

typedef ConstStructValue StructuredAnnotation 

typedef set<ExceptionLabel> ExceptionMeta

typedef set<FunctionLabel> FunctionMeta

enum FunctionLabel {
    Oneway
    Readonly
    Idempotent
}

enum ExceptionLabel {
    Safe
    Unsafe
    Transient
    Permanent
    Stateful
    Server
    Client
}

enum Category {
    Constant
    Bool
    Byte // I8
    I16
    I32
    I64
    Double
    String
    Binary
    Map
    List
    Set
    Enum
    Struct
    Union
    Exception
    Typedef
    Service
    Sink
    Stream
}

struct Reference {
    1: string Name // The name of the referenced type with out IDL name prefix.
    2: i32 Index   // The index of the included IDL that contains the referenced type
}

struct Annotation {
    1: string Key
    2: list<string> Values
}

struct Type {
    1: string Name                           // base types | container types | identifier | selector
    2: optional Type KeyType                 // if Name is 'map'
    3: optional Type ValueType               // if Name is 'map', 'list', or 'set'
    4: string CppType                        // map, set, list
    5: Annotations Annotations
    6: Category Category                     // the **final** category resolved
    7: optional Reference Reference          // when Name is an identifier referring to an external type
    8: optional bool IsTypedef               // whether this type is a typedef
    9: optional list<Field> KeyTypeThrows    // for sink<KeyType throws ..., ValType>
    10: optional list<Field> ValueTypeThrows // for sink<KeyType, ValType throws ...>
}

struct Namespace {
    1: string Language
    2: string Name
    3: Annotations Annotations
}

struct Typedef {
    1: optional Type Type
    2: string Alias
    3: Annotations Annotations
    4: string ReservedComments
    5: StructuredAnnotations StructuredAnnotations
}

struct EnumValue {
    1: string Name
    2: i64 Value
    3: Annotations Annotations
    4: string ReservedComments
    5: StructuredAnnotations StructuredAnnotations
}

struct Enum {
    1: string Name
    2: list<EnumValue> Values
    3: Annotations Annotations
    4: string ReservedComments
    5: StructuredAnnotations StructuredAnnotations
}

enum ConstType {
    ConstDouble
    ConstInt
    ConstLiteral
    ConstIdentifier
    ConstList
    ConstMap
    ConstStruct
}

// ConstValueExtra provides extra information when the Type of a ConstValue is ConstIdentifier.
struct ConstValueExtra {
    1: bool IsEnum    // whether the value resolves to an enum value
    2: i32 Index = -1 // the include index if Index > -1
    3: string Name    // the name of the value without selector
    4: string Sel     // the selector
}

struct ConstValue {
    1: ConstType Type
    2: optional ConstTypedValue TypedValue
    3: optional ConstValueExtra Extra
}

union ConstTypedValue {
    1: double Double
    2: i64 Int
    3: string Literal
    4: string Identifier
    5: list<ConstValue> List
    6: list<MapConstValue> Map
    7: ConstStructValue Struct
}

struct MapConstValue {
    1: optional ConstValue Key
    2: optional ConstValue Value
}

struct ConstStructValue {
    1: string Identifier
    2: list<StructConstValue> Values
}

struct StructConstValue {
    1: string Key
    2: ConstValue Value
}

struct Constant {
    1: string Name
    2: optional Type Type
    3: optional ConstValue Value
    4: Annotations Annotations
    5: string ReservedComments
    6: StructuredAnnotations StructuredAnnotations
}

enum FieldType {
    Default
    Required
    Optional
}

struct Field {
    1: i32 ID
    2: string Name
    3: FieldType Requiredness
    4: Type Type
    5: optional ConstValue Default // ConstValue
    6: Annotations Annotations
    7: string ReservedComments
    8: StructuredAnnotations StructuredAnnotations
}

struct StructLike {
    1: string Category // "struct", "union" or "exception"
    2: string Name
    3: list<Field> Fields
    4: Annotations Annotations
    5: string ReservedComments
    6: StructuredAnnotations StructuredAnnotations
    7: ExceptionMeta ExceptionMeta
}

struct Function {
    1: string Name
    2: bool Oneway
    3: bool Void
    4: optional Type FunctionType
    5: list<Field> Arguments
    6: list<Field> Throws
    7: Annotations Annotations
    8: string ReservedComments
    9: StructuredAnnotations StructuredAnnotations
    10: FunctionMeta FunctionMeta
    11: ReturnClause ReturnClause
}

struct Service {
    1: string Name
    2: string Extends
    3: list<Function> Functions
    4: Annotations Annotations

    // If Extends is not empty and it references to a service defined in an
    // included IDL, then Reference will be set.
    5: optional Reference Reference

    6: string ReservedComments
    7: StructuredAnnotations StructuredAnnotations
    8: optional list<Performs> Performs
}

struct Include {
    1: string Path               // The path literal in the include statement.
    2: optional Thrift Reference // The parsed AST of the included IDL.
    3: optional bool Used        // If this include is used in the IDL
}

struct Performs {
    1: string Interaction
    2: string ReservedComments
}

struct ReturnClause {
    1: Type Type0
    2: optional Type Type1
    3: optional Type Type2
}

struct Package {
    1: string Name
    2: StructuredAnnotations StructuredAnnotations
    3: string ReservedComments
}

struct Interaction {
    1: string Name
    2: list<Function> Functions
    3: Annotations Annotations
    4: string ReservedComments
    5: StructuredAnnotations StructuredAnnotations
}

// Thrift is the AST of the current IDL with symbols sorted.
struct Thrift {
    1: string Filename            // A valid path of current thrift IDL.
    2: list<Include> Includes     // Direct dependencies.
    3: list<string> CppIncludes
    4: list<Namespace> Namespaces
    5: list<Typedef> Typedefs
    6: list<Constant> Constants
    7: list<Enum> Enums
    8: list<StructLike> Structs
    9: list<StructLike> Unions
    10: list<StructLike> Exceptions
    11: list<Service> Services

    // Name2Category keeps a mapping for all global names with their **direct** category.
    12: map<string, Category> Name2Category

    13: optional Package Package
    14: list<string> HsIncludes
    15: list<Interaction> Interactions
}
