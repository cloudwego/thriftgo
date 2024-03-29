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


struct PersonA{
    1:required string name
}(
    // 错误的 option 名称
    entity.person_xxx_info = '{
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
    entity.person_basic_info = '{
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
    entity.person_struct_info = '{
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
}//(
    // 错误的 option 语法
//    option = 'entity.person_struct_info := {
//            valueteststruct:{
//             name: "lee"
//             innerStruct:{
//                 email:"no email"
//              }
//            }
//    }'
//)
struct PersonE{
    1:required string name
}(
    // 错误的 kv 语法
     entity.person_container_info = '{
                valuemap:{{"hey1":"value1"}
                valuelist:["list1","list2"]
                valueset:["list3","list4"]
                valuelistset:[[a,b,c],[d,e,f]]
                valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]
                valuemapstruct:{k1:{email:e1} k2:{email:e2}}
        }'
)
struct PersonF{
    1:required string name (entity.person_field_info="'the name of this person'")
}(
    // 没有 include 对应 IDL
    validation.person_string_info = 'hello'
)


