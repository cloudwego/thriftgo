namespace go codegen.annotation.entity

struct _StructOptions {
      1:required PersonInfo MyStructOption
}

struct _FieldOptions {
      1:required string MyDescriptor

}

struct PersonInfo{
    1:required i8 valuei8
    2:required i16 valuei16;
    3:required i32 valuei32;
    4:required i64 valuei64;
    5:required string valuestring;
    6:required byte valuebyte;
    7:required double valuedouble;
    8:required binary valuebinary;
    9:required bool valuebool;

    11:required map<string,string> valuemap;
    12:required list<string> valuelist;
    13:required set<string> valueset;
    14:required list<set<string>> valuelistset;
    15:required list<set<InnerStruct>> valuelistsetstruct;
    16:required map<string,InnerStruct> valuemapStruct;

    21:required TestStruct valuestruct
}

struct TestStruct{
    1:required string name
    2:required string age
    3:required InnerStruct is
}

struct InnerStruct{
    1:required string email
}