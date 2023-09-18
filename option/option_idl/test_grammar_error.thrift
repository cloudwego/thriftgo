namespace go option_test
include "annotations/entity/entity.thrift"


struct PersonA{
    1:required string name
}(
    // 错误的 option 名称
    option = 'entity.person_xxx_info = {
            valuei16:16
            valuei32:32
            valuei64:64
            valuestring:\'example@email.com\'
            valuebyte: 1
            valuebinary: 12
            valuedouble:3.14159
            valuebool: true
    }'
)
struct PersonB{
    1:required string name
}(
    // 错误的 field value
    option = 'entity.person_basic_info = {
                valuei16:hellostring
                valuei32:32
                valuei64:64
                valuestring:\'example@email.com\'
                valuebyte: 1
                valuebinary: 12
                valuedouble:3.14159
                valuebool: true
        }'
)

struct PersonC{
    1:required string name
}(
    // 错误的 field 名称
    option = 'entity.person_struct_info = {
            value_xxx:{
             name: "lee"
             innerStruct:{
                 email:"no email"
              }
            }
    }'
)

struct PersonD{
    1:required string name
}(
    // 错误的 option 语法
    option = 'entity.person_struct_info := {
            valueteststruct:{
             name: "lee"
             innerStruct:{
                 email:"no email"
              }
            }
    }'
)
struct PersonE{
    1:required string name
}(
    // 错误的 kv 语法
     option = 'entity.person_container_info = {
                valuemap:{{"hey1":"value1"}
                valuelist:["list1","list2"]
                valueset:["list3","list4"]
                valuelistset:[[a,b,c],[d,e,f]]
                valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]
                valuemapstruct:{k1:{email:e1} k2:{email:e2}}
        }'
)
struct PersonF{
    1:required string name (option = "entity.person_field_info='the name of this person'")
}(
    // 没有 include 对应 IDL
    option = 'validation.person_string_info = hello'
)


