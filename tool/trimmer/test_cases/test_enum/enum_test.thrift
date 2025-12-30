// Copyright 2025 CloudWeGo Authors
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

// This enum is used by Person struct
enum Status {
    ACTIVE = 1,
    INACTIVE = 2,
    PENDING = 3
}

// This enum is NOT used anywhere and should be trimmed
enum UnusedColor {
    RED = 1,
    GREEN = 2,
    BLUE = 3
}

// This enum is used as a field type
enum Gender {
    MALE = 1,
    FEMALE = 2,
    OTHER = 3
}

// This enum is NOT used and should be trimmed
enum UnusedPriority {
    LOW = 1,
    MEDIUM = 2,
    HIGH = 3
}

// This enum is used in service method
enum ResponseCode {
    SUCCESS = 0,
    ERROR = 1,
    TIMEOUT = 2
}

struct Person {
    1: string name
    2: Status status
    3: Gender gender
}

service TestService {
    ResponseCode getStatus(1: string id)
    Person getPerson(1: string id)
}
