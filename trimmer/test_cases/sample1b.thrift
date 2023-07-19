namespace go sample1a

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

const i32 DEFAULT_CODE = 3000;
const string trash_string = "trash!"
