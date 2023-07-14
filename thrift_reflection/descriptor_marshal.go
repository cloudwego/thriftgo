package thrift_reflection

import (
	"bytes"
	"compress/gzip"
	"github.com/cloudwego/thriftgo/generator/golang/extension/meta"
	"io/ioutil"
)

func (fd *FileDescriptor) Marshal() ([]byte, error) {
	bytes, err := meta.Marshal(fd)
	if err != nil {
		return nil, err
	}
	return doGzip(bytes)
}

func Unmarshal(bytes []byte) (*FileDescriptor, error) {
	bytes, err := doUnzip(bytes)
	if err != nil {

		return nil, err
	}
	fd := NewFileDescriptor()
	if err := meta.Unmarshal(bytes, fd); err != nil {
		return nil, err
	}
	return fd, nil
}

func doGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return ioutil.ReadAll(&buffer)
}

func doUnzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(data)
	reader, err := gzip.NewReader(&buffer)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}
