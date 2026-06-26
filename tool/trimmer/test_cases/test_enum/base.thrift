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

// This enum will be referenced and used
enum AccountType {
    FREE = 1,
    PREMIUM = 2,
    ENTERPRISE = 3
}

// This enum will not be used and should be trimmed
enum UnusedLevel {
    LEVEL1 = 1,
    LEVEL2 = 2,
    LEVEL3 = 3
}

// This enum will be used via typedef
enum Country {
    US = 1,
    UK = 2,
    CN = 3
}
