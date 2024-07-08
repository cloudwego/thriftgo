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

namespace go option_gen
include "annotations/entity/entity.thrift"
include "annotations/validation/validation.thrift"

struct IDCard{
    1:required string number
    2:required i8 age
}


struct Person{
    1:required string name (entity.person_field_info='the name of this person')
    2:required IDCard id
}(
    aaa.bbb = "hello"
    entity.person_basic_info = '{
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
    entity.person_struct_info = '{
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
    entity.person_container_info = '{
            valuemap:{"hey1":"value1"}
            valuelist:["list1","list2"]
            valueset:["list3","list4"]
            valuelistset:[[a,b,c],[d,e,f]]
            valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]
            valuemapstruct:{k1:{email:e1} k2:{email:e2}}
    }'
    validation.person_string_info = 'hello'
    validation.person_map_info = '{"hey1":"value1"}'
    validation.person_enum_info = 'XXL'
    validation.person_basic_typedef_info = '"hello there"'
    validation.person_struct_typedef_info = '{name:"empty name"}'
    validation.person_struct_default_value_info = '{v1:"v1 string"}'
)

struct PersonB{

}(
    entity.person_basic_info = '{
            valuei8:8,valuei16:16,
            valuei32:32,
            valuei64:64 ,
            valuestring:\'example@email.com\',
            valuebyte: 1,
            valuebinary: 12 ,
            valuedouble:3.14159,
            valuebool: true,
    }'
    entity.person_struct_info = '{
            valuestruct:{email:"empty email"}
            valueteststruct:{
             name: "lee",
             innerStruct:{
                 email:"no email",
              }
            }
            valueenum: B valuestructtypedef:{email:"empty email"} , valuebasictypedef: "hello there"
    }'
    entity.person_container_info = '{
            valuemap:{"hey1":"value1"},valuelist:["list1","list2"] valueset:["list3","list4"] ,valuelistset:[[a,b,c],[d,e,f]]
            valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]],
            valuemapstruct:{k1:{email:e1} k2:{email:e2}}
    }'
)


struct PersonC{

}(
    entity.person_basic_info.valuei8 = '8'
    entity.person_basic_info.valuei16 = '16'
    entity.person_struct_info.valuestruct = '{email:"empty email"}'
    entity.person_struct_info.valueteststruct.name = "lee"
    entity.person_struct_info.valueteststruct.innerStruct.email = "no email"
)

enum MyEnum{
    A
    (
        validation.enum_value_info = '{
            name: EnumValueInfoName
            number: 222
        }'
    )
    B
}(
    validation.enum_info = '{
        name: EnumInfoName
        number: 333
    }'
)

service MyService{
    string M1()
    (
        validation.method_info = '{
            name: MethodInfoName
            number: 555
        }'
    )
    string M2()
    (
        validation.method_info = '{
            name: MethodInfoName
            number: 444
        }'
    )
}(
    validation.svc_info = '{
        name: ServiceInfoName
        number: 666
    }'

)

struct TinyStruct{
    1:required bool b1
    2:required bool b2
}

