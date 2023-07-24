// Copyright 2023 CloudWeGo Authors
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

namespace go thrift_reflection_test

struct IDCard{
    1:required string number
    2:required i8 age
}

union MyUnion{
    1:optional string number
    2:optional i8 age
}

exception MyException{
    1:required string msg
    2:required i8 code
}

// Person Comment
struct Person{
    1:required string name (tag = "this is name tag")
    2:required IDCard id
    3:required Gender gender
    4:required MyException exp
    5:required MyUnion uni
    6:required SpecialString typedefValue
    7:required string defaultValue = "123321"
    8:required string defaultConst = MY_CONST
}(
    k1 = 'hello'
    k2 = 'hey'
)

enum Gender{
    MALE
    FEMALE
}

enum Size{
    S
    M
    L
    XL
    XXL
}

const string MY_CONST = "hello"
const i64 MY_INT_CONST = 123
const double MY_FLOAT_CONST = 123.333
const bool MY_BOOL_CONST = true
const byte MY_BYTE_CONST = 1
const binary MY_BINARY_CONST = "1"
const list<string> MY_LIST_CONST = ["a","b","c"]
const map<string,string> MY_MAP_CONST = {"k1":"v1","k2":"v2","k3":"v3"}

typedef string SpecialString

typedef Person SpecialPerson

service MyService{
    string M1(1:required Person p),
    string M2(1:required Person p2),
    A1 M3(1:required A0 a0,2:required A3 a3)
}

struct A0{
    // 间接依赖 B B1 C D D1 D2 E F
    1:required string f1
    2:required B f2
    3:required map<string,C> f3
    4:required map<D,map<E,list<F>>> f4
}


struct A1{
    // 间接依赖 A2
    1:required string f1
    2:required A2 f2
}

struct A2{
    1:required string f1
}

struct A3{
    1:required string f1
}

struct B{
    // 间接依赖 B1 E C
    1:required B1 f1
    2:required E f2
}

struct B1{
    1:required string name
}

struct C{
    // 间接依赖 B B1 E
    1:required B f1
}

struct D{
    // 间接依赖 D1 D2
    1:required map<D1,D2> f1
}

struct D1{
    1:required string name
}

struct D2{
    1:required string name
}

struct E{
    // 间接依赖 C B B1
    1:required C f1
}

struct F{
    1:required string name
}


