namespace go testtypedef

typedef i64 UserId

struct U{
    1:  UserId id
}

Service S{
    U get()
}