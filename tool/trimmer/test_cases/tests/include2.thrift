include "include3.thrift"

struct B{
    1: B b
    2: include3.C c
}