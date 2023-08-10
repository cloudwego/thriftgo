include "include2.thrift"

struct A{
    1: include2.B b
}

service S{
    A serve()
}