namespace go testtypedef

typedef i64 UserId
typedef i64 ID
const i64 default_num = 2
const ID def = default_num

struct U{
    1:  UserId id
}

service S{
    U get()
}