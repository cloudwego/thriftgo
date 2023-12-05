module example.com/test

go 1.17

require (
	github.com/apache/thrift v0.13.0
	github.com/cloudwego/thriftgo v0.0.0-00010101000000-000000000000
)

replace github.com/cloudwego/thriftgo => ../../..
