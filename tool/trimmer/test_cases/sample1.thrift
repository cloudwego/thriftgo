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

include "sample1b.thrift"
include "sample1c.thrift"

namespace go sample1a (a = "b\"c\"")
//test
cpp_include "sample1.thrift"
const sample1b.neccesary_typedef const_abc = 1 (b = "c")
const list<string> theList = ["a","b"]
const map<string, i32> theMap = {"a":1, "b":2, "c": Gender.MALE}
const map<string, string> anotherMap = {}
// test2

//test3
struct Person {
    // test4
    1: required string name
// test 5
    2: optional i32 age

    //test 6

    3: optional Gender gender
    4: optional MyAnotherGender xx
}

/*
test7
 */

typedef Gender(key="v") MyGender (key = "1", key = "2", key2 = "v2")

typedef MyGender MyAnotherGender
typedef i32 a

// out enum"ZZ"
enum Gender {
    // in enum
    MALE = 3,
    FEMALE (key = "1", key = "2", key2 = "v2")
} (a = "b")

struct Address {
    1: required string(key = "v") street
    2: required string city
    3: optional string state
    4: required a country
}

// @pResErve
struct Company {
    1: required string name
    2: optional Address address
}

struct Employee {
    1: required string id
    2: required Person person
    3: optional set<Company> company
    4: optional list<string> skills
    5: optional map<Experience, a> experience
    6: optional sample1b.Department department
}

struct Experience {
    1: required string company
    2: required string job_title
    3: optional Address address
    4: optional i32 start_year
    5: optional i32 end_year
    6: optional list<string> responsibilities
}

exception Project {
    1: required string name
    2: optional Company company
    3: optional list<Employee> employees
}

struct Simple { // should not appear
    1: string str
    2: i32 int
}

struct MaybeUseless{
}

service EmployeeService extends sample1b.GetPerson {
    Employee getEmployee(1: string id)
    void addEmployee(1: Employee employee)
    void updateEmployee(1: set<string> id, 2: Employee employee)
}

service ProjectService {
    Project getProject(1: string id)
    oneway void addProject(1: Project project)
    void updateProject(1: string id,
    2: Project project)
}

service CompanyService {
    Company getCompany(1: string id)
    void addCompany(1: Company company) throws(1: sample1b.AnotherException exc)
    void updateCompany(1: string id, 2: Company company)
    list<sample1b.Department> getDepartments(1: string company_id)
    void anotherUselessMethod(1: MaybeUseless useless)
}


