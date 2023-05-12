package thrift_reflection

import (
	"fmt"
)

var GlobalFd = map[string]*FileDescriptor{}

func RegisterIDL(bytes []byte) *FileDescriptor {
	fd, err := UnmarshalAst(bytes)
	if err != nil {
		fmt.Println("[Error]:" + err.Error())
	}
	GlobalFd[fd.Filepath] = fd
	return fd
}
