// Code generated by thriftgo (0.4.1). DO NOT EDIT.

package plugin

import (
	"fmt"
	"github.com/cloudwego/thriftgo/generator/golang/extension/meta"
	"github.com/cloudwego/thriftgo/parser"
)

type Request struct {
	Version             string         `thrift:"Version,1,required" frugal:"1,required,string" json:"Version"`
	GeneratorParameters []string       `thrift:"GeneratorParameters,2,required" frugal:"2,required,list<string>" json:"GeneratorParameters"`
	PluginParameters    []string       `thrift:"PluginParameters,3,required" frugal:"3,required,list<string>" json:"PluginParameters"`
	Language            string         `thrift:"Language,4,required" frugal:"4,required,string" json:"Language"`
	OutputPath          string         `thrift:"OutputPath,5,required" frugal:"5,required,string" json:"OutputPath"`
	Recursive           bool           `thrift:"Recursive,6,required" frugal:"6,required,bool" json:"Recursive"`
	AST                 *parser.Thrift `thrift:"AST,7,required" frugal:"7,required,parser.Thrift" json:"AST"`
}

func init() {
	meta.RegisterStruct(NewRequest, []byte{
		0xb, 0x0, 0x1, 0x0, 0x0, 0x0, 0x7, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0xb, 0x0,
		0x2, 0x0, 0x0, 0x0, 0x6, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0xf, 0x0, 0x3, 0xc, 0x0,
		0x0, 0x0, 0x7, 0x6, 0x0, 0x1, 0x0, 0x1, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x7, 0x56,
		0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc, 0x0, 0x4,
		0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x2, 0xb, 0x0,
		0x2, 0x0, 0x0, 0x0, 0x13, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x50, 0x61,
		0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc,
		0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xf, 0xc, 0x0, 0x3, 0x8, 0x0, 0x1, 0x0,
		0x0, 0x0, 0xb, 0x0, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x3, 0xb, 0x0, 0x2, 0x0, 0x0,
		0x0, 0x10, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65,
		0x72, 0x73, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc, 0x0, 0x4, 0x8, 0x0, 0x1, 0x0,
		0x0, 0x0, 0xf, 0xc, 0x0, 0x3, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0,
		0x6, 0x0, 0x1, 0x0, 0x4, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x8, 0x4c, 0x61, 0x6e, 0x67,
		0x75, 0x61, 0x67, 0x65, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc, 0x0, 0x4, 0x8, 0x0,
		0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x5, 0xb, 0x0, 0x2, 0x0,
		0x0, 0x0, 0xa, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x50, 0x61, 0x74, 0x68, 0x8, 0x0, 0x3,
		0x0, 0x0, 0x0, 0x1, 0xc, 0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0,
		0x6, 0x0, 0x1, 0x0, 0x6, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x9, 0x52, 0x65, 0x63, 0x75,
		0x72, 0x73, 0x69, 0x76, 0x65, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc, 0x0, 0x4, 0x8,
		0x0, 0x1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x7, 0xb, 0x0, 0x2,
		0x0, 0x0, 0x0, 0x3, 0x41, 0x53, 0x54, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc, 0x0,
		0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0,
	})
}

func NewRequest() *Request {
	return &Request{}
}

func (p *Request) InitDefault() {
}

func (p *Request) GetVersion() (v string) {
	return p.Version
}

func (p *Request) GetGeneratorParameters() (v []string) {
	return p.GeneratorParameters
}

func (p *Request) GetPluginParameters() (v []string) {
	return p.PluginParameters
}

func (p *Request) GetLanguage() (v string) {
	return p.Language
}

func (p *Request) GetOutputPath() (v string) {
	return p.OutputPath
}

func (p *Request) GetRecursive() (v bool) {
	return p.Recursive
}

var Request_AST_DEFAULT *parser.Thrift

func (p *Request) GetAST() (v *parser.Thrift) {
	if !p.IsSetAST() {
		return Request_AST_DEFAULT
	}
	return p.AST
}

func (p *Request) IsSetAST() bool {
	return p.AST != nil
}

func (p *Request) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Request(%+v)", *p)
}

var fieldIDToName_Request = map[int16]string{
	1: "Version",
	2: "GeneratorParameters",
	3: "PluginParameters",
	4: "Language",
	5: "OutputPath",
	6: "Recursive",
	7: "AST",
}

type Generated struct {
	Content        string  `thrift:"Content,1,required" frugal:"1,required,string" json:"Content"`
	Name           *string `thrift:"Name,2,optional" frugal:"2,optional,string" json:"Name,omitempty"`
	InsertionPoint *string `thrift:"InsertionPoint,3,optional" frugal:"3,optional,string" json:"InsertionPoint,omitempty"`
}

func init() {
	meta.RegisterStruct(NewGenerated, []byte{
		0xb, 0x0, 0x1, 0x0, 0x0, 0x0, 0x9, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64,
		0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x6, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0xf, 0x0, 0x3,
		0xc, 0x0, 0x0, 0x0, 0x3, 0x6, 0x0, 0x1, 0x0, 0x1, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0,
		0x7, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x1, 0xc,
		0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x2,
		0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x4, 0x4e, 0x61, 0x6d, 0x65, 0x8, 0x0, 0x3, 0x0, 0x0,
		0x0, 0x2, 0xc, 0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x6, 0x0,
		0x1, 0x0, 0x3, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0xe, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74,
		0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x2, 0xc,
		0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0,
	})
}

func NewGenerated() *Generated {
	return &Generated{}
}

func (p *Generated) InitDefault() {
}

func (p *Generated) GetContent() (v string) {
	return p.Content
}

var Generated_Name_DEFAULT string

func (p *Generated) GetName() (v string) {
	if !p.IsSetName() {
		return Generated_Name_DEFAULT
	}
	return *p.Name
}

var Generated_InsertionPoint_DEFAULT string

func (p *Generated) GetInsertionPoint() (v string) {
	if !p.IsSetInsertionPoint() {
		return Generated_InsertionPoint_DEFAULT
	}
	return *p.InsertionPoint
}

func (p *Generated) IsSetName() bool {
	return p.Name != nil
}

func (p *Generated) IsSetInsertionPoint() bool {
	return p.InsertionPoint != nil
}

func (p *Generated) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Generated(%+v)", *p)
}

var fieldIDToName_Generated = map[int16]string{
	1: "Content",
	2: "Name",
	3: "InsertionPoint",
}

type Response struct {
	Error    *string      `thrift:"Error,1,optional" frugal:"1,optional,string" json:"Error,omitempty"`
	Contents []*Generated `thrift:"Contents,2,optional" frugal:"2,optional,list<Generated>" json:"Contents,omitempty"`
	Warnings []string     `thrift:"Warnings,3,optional" frugal:"3,optional,list<string>" json:"Warnings,omitempty"`
}

func init() {
	meta.RegisterStruct(NewResponse, []byte{
		0xb, 0x0, 0x1, 0x0, 0x0, 0x0, 0x8, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0xb,
		0x0, 0x2, 0x0, 0x0, 0x0, 0x6, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0xf, 0x0, 0x3, 0xc,
		0x0, 0x0, 0x0, 0x3, 0x6, 0x0, 0x1, 0x0, 0x1, 0xb, 0x0, 0x2, 0x0, 0x0, 0x0, 0x5,
		0x45, 0x72, 0x72, 0x6f, 0x72, 0x8, 0x0, 0x3, 0x0, 0x0, 0x0, 0x2, 0xc, 0x0, 0x4, 0x8,
		0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x2, 0xb, 0x0, 0x2,
		0x0, 0x0, 0x0, 0x8, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x8, 0x0, 0x3, 0x0,
		0x0, 0x0, 0x2, 0xc, 0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xf, 0xc, 0x0, 0x3,
		0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x6, 0x0, 0x1, 0x0, 0x3, 0xb,
		0x0, 0x2, 0x0, 0x0, 0x0, 0x8, 0x57, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x73, 0x8, 0x0,
		0x3, 0x0, 0x0, 0x0, 0x2, 0xc, 0x0, 0x4, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xf, 0xc,
		0x0, 0x3, 0x8, 0x0, 0x1, 0x0, 0x0, 0x0, 0xb, 0x0, 0x0, 0x0, 0x0,
	})
}

func NewResponse() *Response {
	return &Response{}
}

func (p *Response) InitDefault() {
}

var Response_Error_DEFAULT string

func (p *Response) GetError() (v string) {
	if !p.IsSetError() {
		return Response_Error_DEFAULT
	}
	return *p.Error
}

var Response_Contents_DEFAULT []*Generated

func (p *Response) GetContents() (v []*Generated) {
	if !p.IsSetContents() {
		return Response_Contents_DEFAULT
	}
	return p.Contents
}

var Response_Warnings_DEFAULT []string

func (p *Response) GetWarnings() (v []string) {
	if !p.IsSetWarnings() {
		return Response_Warnings_DEFAULT
	}
	return p.Warnings
}

func (p *Response) IsSetError() bool {
	return p.Error != nil
}

func (p *Response) IsSetContents() bool {
	return p.Contents != nil
}

func (p *Response) IsSetWarnings() bool {
	return p.Warnings != nil
}

func (p *Response) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("Response(%+v)", *p)
}

var fieldIDToName_Response = map[int16]string{
	1: "Error",
	2: "Contents",
	3: "Warnings",
}
