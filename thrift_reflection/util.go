package thrift_reflection

import (
	"bytes"
	"compress/gzip"
	"github.com/cloudwego/thriftgo/generator/golang/extension/meta"
	"github.com/cloudwego/thriftgo/parser"
	"io/ioutil"
	"strings"
)

func getTypeDescriptor(path string, typeStruct *parser.Type) *TypeDescriptor {

	if typeStruct == nil {
		return nil
	}

	return &TypeDescriptor{
		Filepath:  path,
		TypeName:  typeStruct.Name,
		KeyType:   getTypeDescriptor(path, typeStruct.KeyType),
		ValueType: getTypeDescriptor(path, typeStruct.ValueType),
	}
}

func getStructDescriptor(path string, structLike *parser.StructLike) *StructDescriptor {

	fields := []*FieldDescriptor{}
	for _, f := range structLike.GetFields() {
		fields = append(fields, getFieldDescriptor(path, f))
	}

	return &StructDescriptor{
		Filepath: path,
		Name:     structLike.GetName(),
		Fields:   fields,

		Annotation: getAnnotationMap(structLike.GetAnnotations()),
		Comments:   structLike.GetReservedComments(),
		Extra:      nil,
	}
}

func getTypedefDescriptor(path string, td *parser.Typedef) *TypedefDescriptor {
	return &TypedefDescriptor{
		Filepath:    path,
		Type:        getTypeDescriptor(path, td.GetType()),
		Alias:       td.GetAlias(),
		Annotations: getAnnotationMap(td.GetAnnotations()),
		Comments:    td.GetReservedComments(),
		Extra:       nil,
	}
}

func getEnumDescriptor(path string, enum *parser.Enum) *EnumDescriptor {

	values := []*EnumValueDescriptor{}
	for _, ev := range enum.Values {
		values = append(values, &EnumValueDescriptor{
			Filepath:    path,
			Name:        ev.GetName(),
			Value:       ev.GetValue(),
			Annotations: getAnnotationMap(ev.GetAnnotations()),
			Comments:    ev.ReservedComments,
			Extra:       nil,
		})
	}

	return &EnumDescriptor{
		Filepath:    path,
		Name:        enum.GetName(),
		Values:      values,
		Annotations: getAnnotationMap(enum.GetAnnotations()),
		Comments:    enum.GetReservedComments(),
		Extra:       nil,
	}
}

func getFieldDescriptor(path string, field *parser.Field) *FieldDescriptor {
	return &FieldDescriptor{
		Filepath:     path,
		Name:         field.GetName(),
		Type:         getTypeDescriptor(path, field.GetType()),
		Requiredness: field.GetRequiredness().String(),
		ID:           field.GetID(),
		Annotations:  getAnnotationMap(field.Annotations),
		Comments:     field.GetReservedComments(),
		Extra:        nil,
	}
}

func getAnnotationMap(annotations parser.Annotations) map[string][]string {
	annotationsMap := map[string][]string{}
	for _, annotation := range annotations {
		annotationsMap[annotation.Key] = annotation.Values
	}
	return annotationsMap
}

func getMethodDescriptor(path string, method *parser.Function) *MethodDescriptor {

	args := []*FieldDescriptor{}
	for _, arg := range method.Arguments {
		args = append(args, getFieldDescriptor(path, arg))
	}

	throws := []*FieldDescriptor{}
	for _, t := range method.Throws {
		throws = append(throws, getFieldDescriptor(path, t))
	}

	return &MethodDescriptor{
		Filepath:        path,
		Name:            method.GetName(),
		Response:        getTypeDescriptor(path, method.FunctionType),
		Args:            args,
		Annotations:     getAnnotationMap(method.GetAnnotations()),
		Comments:        method.GetReservedComments(),
		Extra:           nil,
		ThrowExceptions: throws,
	}
}

func getServiceDescriptor(path string, service *parser.Service) *ServiceDescriptor {

	methods := []*MethodDescriptor{}
	for _, method := range service.Functions {
		methods = append(methods, getMethodDescriptor(path, method))
	}

	return &ServiceDescriptor{
		Filepath:    path,
		Name:        service.GetName(),
		Methods:     methods,
		Annotations: getAnnotationMap(service.GetAnnotations()),
		Comments:    service.GetReservedComments(),
		Extra:       nil,
	}
}

func MarshalAst(fd *FileDescriptor) ([]byte, error) {
	bytes, err := meta.Marshal(fd)
	if err != nil {
		return nil, err
	}
	return doGzip(bytes)
}

func UnmarshalAst(bytes []byte) (*FileDescriptor, error) {
	bytes, err := unGzip(bytes)
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
	compressedData, err := ioutil.ReadAll(&buffer)
	return compressedData, nil
}

func unGzip(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(data)
	reader, _ := gzip.NewReader(&buffer)
	defer reader.Close()
	undatas, _ := ioutil.ReadAll(reader)
	return undatas, nil
}

func (f *FileDescriptor) GetStructDescriptor(name string) *StructDescriptor {
	if f == nil {
		return nil
	}
	for _, s := range f.Structs {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (f *FileDescriptor) GetTypedefDescriptor(name string) *TypedefDescriptor {
	if f == nil {
		return nil
	}
	for _, s := range f.Typedefs {
		if s.Alias == name {
			return s
		}
	}
	return nil
}

func (f *FileDescriptor) GetEnumDescriptor(name string) *EnumDescriptor {
	if f == nil {
		return nil
	}
	for _, s := range f.Enums {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (f *FileDescriptor) GetExceptionDescriptor(name string) *StructDescriptor {
	if f == nil {
		return nil
	}
	for _, s := range f.Exceptions {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (f *FileDescriptor) GetUnionDescriptor(name string) *StructDescriptor {
	if f == nil {
		return nil
	}
	for _, s := range f.Unions {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func GetFileDescriptorFromAst(ast *parser.Thrift) *FileDescriptor {

	services := []*ServiceDescriptor{}

	for _, s := range ast.Services {
		services = append(services, getServiceDescriptor(ast.Filename, s))
	}

	structs := []*StructDescriptor{}
	for _, st := range ast.Structs {
		structs = append(structs, getStructDescriptor(ast.Filename, st))
	}

	exceptions := []*StructDescriptor{}
	for _, ex := range ast.Exceptions {
		exceptions = append(exceptions, getStructDescriptor(ast.Filename, ex))
	}

	unions := []*StructDescriptor{}
	for _, un := range ast.Unions {
		unions = append(unions, getStructDescriptor(ast.Filename, un))
	}

	enums := []*EnumDescriptor{}
	for _, enm := range ast.Enums {
		enums = append(enums, getEnumDescriptor(ast.Filename, enm))
	}

	typedefs := []*TypedefDescriptor{}
	for _, td := range ast.Typedefs {
		typedefs = append(typedefs, getTypedefDescriptor(ast.Filename, td))
	}

	includesMap := map[string]string{}
	for _, inc := range ast.Includes {
		path := inc.Path
		arr := strings.Split(path, ".")
		alias := arr[len(arr)-1]
		includesMap[alias] = path
	}

	namespaceMap := map[string]string{}
	for _, ns := range ast.Namespaces {
		namespaceMap[ns.GetLanguage()] = ns.GetName()
	}

	// todo const!!!!
	//ast.Constants
	return &FileDescriptor{
		Filepath:   ast.Filename,
		Includes:   includesMap,
		Namespaces: namespaceMap,
		Services:   services,
		Structs:    structs,
		Exceptions: exceptions,
		Typedefs:   typedefs,
		Enums:      enums,
		Unions:     unions,
	}
}
