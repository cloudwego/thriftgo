namespace go codegen.test.simple
include "annotations/entity/entity.thrift"
include "annotations/validation/validation.thrift"

enum MyEnum{
    A  (option = "IsOdd=true")
    B  (option = "IsOdd=false")
}(option = "EnumDesc='Hello This Is Enum'")

struct _EnumOptions {
      1:required string EnumDesc
}

struct _EnumValueOptions {
      1:required bool IsOdd
}

struct _StructOptions {
      1:required BasicStruct basicStruct
      2:required StructStruct structStruct
      3:required ContainerStruct containerStruct
}

struct BasicStruct{
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


struct StructStruct{
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

struct ContainerStruct {
    11:required map<string,string> valuemap;
    12:required list<string> valuelist;
    13:required set<InnerStruct> valueset;
    14:required list<set<string>> valuelistset;
    15:required list<set<InnerStruct>> valuelistsetstruct;
    16:required map<string,InnerStruct> valuemapStruct;
}

     // what's this comment?




     //   service definition line 1
     //   service definition line 2
struct Person{
    1:required string name (option = "entity.MyDescriptor='the name of this person'")
    2:required i8 age (option = "validation.MyAnotherDescriptor='the name of this person'")
}(
    option = 'containerStruct = {
        valuemap:{k1:v1 k2:v2 k3:v3}
        valuelist:[a,b,c,d]
        valueset:[{email:e1},{email:e2}]
        valuelistset:[[a,b,c],[d,e,f]]
        valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]
        valuemapStruct:{k1:{email:e1} k2:{email:e2}}
    }'
    option = 'entity.MyStructOption = {
            valuei8:20
            valuei16:20
            valuei32:20
            valuei64:20
            valuestring:\'example@email.com\'
            valuebyte: 1
            valuedouble:3.1415926

            valuemap:{"hey1":"value1"}
            valuelist:["list1","list2"]
            valueset:["list3","list4"]
            valuebool:true
        valuestruct:{
            name: "lee"
            is:{
                email:"no email"
            }
        }
    }'
)


