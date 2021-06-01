namespace * parser

enum FieldType {
    Default
    Required
    Optional
}

enum ConstType {
    ConstDouble
    ConstInt
    ConstLiteral
    ConstIdentifier
    ConstList
    ConstMap
}

struct Type {
    1: string Name,             // base types | container types | identifier | selector
    2: optional Type KeyType,   // if Name is 'map'
    3: optional Type ValueType, // if Name is 'map', 'list', or 'set'
    4: string CppType,          // map, set, list
    5: map<string, string> Annotations
}

struct Namespace {
    1: string Language,
    2: string Name,
    3: map<string, string> Annotations
}

struct Typedef {
    1: optional Type Type,
    2: string Alias
    3: map<string, string> Annotations
}

struct EnumValue {
    1: string Name,
    2: i64 Value
    3: map<string, string> Annotations
}

struct Enum {
    1: string Name,
    2: list<EnumValue> Values
    3: map<string, string> Annotations
}

struct ConstValue {
    1: ConstType Type,
    2: optional ConstTypedValue TypedValue
}

union ConstTypedValue {
    1: double Double,
    2: i64 Int,
    3: string Literal,
    4: string Identifier,
    5: list<ConstValue> List,
    6: list<MapConstValue> Map
}

struct MapConstValue {
    1: optional ConstValue Key,
    2: optional ConstValue Value
}

struct Constant {
    1: string Name,
    2: optional Type Type,
    3: optional ConstValue Value
    4: map<string, string> Annotations
}

struct Field {
    1: i32 ID,
    2: string Name,
    3: FieldType Requiredness,
    4: Type Type,
    5: optional ConstValue Default // ConstValue
    6: map<string, string> Annotations
}

struct StructLike {
    1: string Category, // "struct", "union" or "exception"
    2: string Name,
    3: list<Field> Fields,
    4: map<string, string> Annotations
}

struct Function {
    1: string Name,
    2: bool Oneway,
    3: bool Void,
    4: optional Type FunctionType,
    5: list<Field> Arguments,
    6: list<Field> Throws
    7: map<string, string> Annotations
}

struct Service {
    1: string Name,
    2: string Extends,
    3: list<Function> Functions
    4: map<string, string> Annotations
}

struct Include {
    1: string Path,              // The path literal in the include statement
    2: optional Thrift Reference // The parsed AST
}

struct Thrift {
    1: string Filename,            // A valid path of current thrift IDL
    2: list<Include> Includes,     // Direct dependencies.
    3: list<string> CppIncludes,
    4: list<Namespace> Namespaces,
    5: list<Typedef> Typedefs,
    6: list<Constant> Constants,
    7: list<Enum> Enums,
    8: list<StructLike> Structs,
    9: list<StructLike> Unions,
    10: list<StructLike> Exceptions,
    11: list<Service> Services
}
