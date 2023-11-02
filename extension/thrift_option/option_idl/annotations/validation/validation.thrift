namespace go option_gen.annotation.validation

struct _StructOptions {
      1:required string person_string_info
      2:required map<string,string> person_map_info
      3:required MyEnum person_enum_info
      4:required MyBasicTypedef person_basic_typedef_info
      5:required MyStructTypedef person_struct_typedef_info
      6:required MyStructWithDefaultVal person_struct_default_value_info
}

struct MyStructWithDefaultVal{
    1:required string v1
    2:required string v2 = "v2"
    3:required i8 v3 = 8
    4:required i16 v4 = 16
    5:required i32 v5 = 32
    6:required i64 v6 = 64
    7:required bool v7 = true
    8:required double v8 = 3.1415926123456
    9:required map<string,string> v9 = {"k1":"v1"}
    10:required list<string> v10 = ["k1","k2"]
    11:required string v11 = HELLO
}

const string HELLO = "hello there"

enum MyEnum{
    X
    XL
    XXL
}

typedef string MyBasicTypedef
typedef TestInfo MyStructTypedef

struct _FieldOptions {
      1:required string card_field_info
}

struct TestInfo{
    1:required string name
    2:required i16 number
}

struct _ServiceOptions {
      1:required TestInfo svc_info
}

struct _MethodOptions {
      1:required TestInfo method_info
}

struct _EnumOptions {
      1:required TestInfo enum_info
}

struct _EnumValueOptions {
      1:required TestInfo enum_value_info
}
