include "sample1b.thrift"
include "sample1c.thrift"

namespace go sample1a

struct Person {
    1: required string name
    2: optional i32 age
    3: optional Gender gender
    4: optional MyAnotherGender xx
}

typedef Gender MyGender

typedef MyGender MyAnotherGender

enum Gender {
    MALE = 0,
    FEMALE = 1
}

struct Address {
    1: required string street
    2: required string city
    3: optional string state
    4: required string country
}

struct Company {
    1: required string name
    2: optional Address address
}

struct Employee {
    1: required string id
    2: required Person person
    3: optional Company company
    4: optional list<string> skills
    5: optional list<Experience> experience
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

struct Project {
    1: required string name
    2: optional Company company
    3: optional list<Employee> employees
}

struct Simple { // should not appear
    1: string str
    2: i32 int
}

service EmployeeService extends sample1b.GetPerson {
    Employee getEmployee(1: string id)
    void addEmployee(1: Employee employee)
    void updateEmployee(1: string id, 2: Employee employee)
}

service ProjectService {
    Project getProject(1: string id)
    void addProject(1: Project project)
    void updateProject(1: string id, 2: Project project)
}

service CompanyService {
    Company getCompany(1: string id)
    void addCompany(1: Company company)
    void updateCompany(1: string id, 2: Company company)
    list<sample1b.Department> getDepartments(1: string company_id)
}


