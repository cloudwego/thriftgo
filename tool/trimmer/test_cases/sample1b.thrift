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

namespace go sample1b

typedef i32 neccesary_typedef

struct Department{
    1: string name
    2: i32 id
}

struct Trash{
    1: bool trashh
}

struct Person{
    1: i32 id
}

service GetPerson{
    Person get(1: i32 id = 1) throws (1: UserException e)
}

service UselessSvc{
    Trash get()
}

exception UserException {
  1: i32 errorCode = DEFAULT_CODE,
  2: string message,
  3: string userinfo
}

exception AnotherException{
  1: i32 abc
}

exception NotDirectInclude{

}

const i32 DEFAULT_CODE = 3000;
const string trash_string = "trash!"
