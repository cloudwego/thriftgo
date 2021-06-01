all: thrift.peg.go AST.go

thrift.peg.go: thrift.peg
	which peg >/dev/null || go get -u github.com/pointlander/peg
	peg ./thrift.peg

AST.go: AST.thrift
	thriftgo --gen go AST.thrift && mv gen-go/parser/AST.go AST.go && rm -r gen-go
	sed -i '/func (p \*\(Type\|ConstValue\)) String() string/,+6d' AST.go
