namespace go option_test
include "annotations/entity/entity.thrift"
include "annotations/validation/validation.thrift"

struct IDCard{
    1:required string number
    2:required i8 age
}

struct Person{
    1:required string name (option = "entity.person_field_info='the name of this person'")
    2:required IDCard id
}(
    option = 'entity.person_basic_info = {
            valuei8:8
            valuei16:16
            valuei32:32
            valuei64:64
            valuestring:\'example@email.com\'
            valuebyte: 1
            valuebinary: 12
            valuedouble:3.14159
            valuebool: true
    }'
    option = 'entity.person_struct_info = {
            valuestruct:{email:"empty email"}
            valueteststruct:{
             name: "lee"
             innerStruct:{
                 email:"no email"
              }
            }
            valueenum: B
            valuestructtypedef:{email:"empty email"}
            valuebasictypedef: "hello there"
    }'
    option = 'entity.person_container_info = {
            valuemap:{"hey1":"value1"}
            valuelist:["list1","list2"]
            valueset:["list3","list4"]
            valuelistset:[[a,b,c],[d,e,f]]
            valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]
            valuemapstruct:{k1:{email:e1} k2:{email:e2}}
    }'
    option = 'validation.person_string_info = hello'
    option = 'validation.person_map_info = {"hey1":"value1"}'
    option = 'validation.person_enum_info = XXL'
    option = 'validation.person_basic_typedef_info = "hello there"'
    option = 'validation.person_struct_typedef_info = {name:"empty name"}'
    option = 'validation.person_struct_default_value_info = {v1:"v1 string"}'
)

enum MyEnum{
    A
    (
        option = 'validation.enum_value_info = {
            name: EnumValueInfoName
            number: 222
        }'
    )
    B
}(
    option = 'validation.enum_info = {
        name: EnumInfoName
        number: 333
    }'
)

service MyService{
    string M1()
    (
        option = 'validation.method_info = {
            name: MethodInfoName
            number: 555
        }'
    )
    string M2()
    (
        option = 'validation.method_info = {
            name: MethodInfoName
            number: 444
        }'
    )
}(
    option = 'validation.svc_info = {
        name: ServiceInfoName
        number: 666
    }'

)


