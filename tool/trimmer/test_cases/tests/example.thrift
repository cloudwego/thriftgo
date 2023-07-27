namespace go hello.cloudwego.team

include "base.thrift"

struct MyReq{
    1:required string name,
    2:required string id,
    255: optional base.Base base
}

struct MyResp{
    1:required string text
    2:required base.BaseResp baseResp
    3:required list<MyResp> myselfs
}

service greet {
	MyResp Hello(1:required MyReq aareqaaa)
}