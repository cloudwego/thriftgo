
# String Literals in the IDL

String literals in the IDL has two forms: single quoted or double quoted.
They are used in include declarations, string constant initializations and annotations.
As the literal itself will be copied to generated codes and interpreted by the target language, it is necessary to define how escape sequences will be handled by thriftgo.

Take a look at the grammar in **thrift.peg**, we can see that the thriftgo parser only cares about the delimiters that are used to quote the string literal. So it is possible for a user to write any escape sequences in a literal as long as they do not conflict with the delimiter.

```
EscapeLiteralChar <- '\\' ["']

Literal <- Skip '"' <(EscapeLiteralChar / !'"' .)*> '"' Indent*
        / Skip "'" <(EscapeLiteralChar / !"'" .)*> "'" Indent*
```

Since most programming languages will interprete escape sequences in string literals, we should not translate escape sequences after parsing. But for delimiters in literals, they must be unescaped. Otherwise there will be some strings that can not be expressed in the IDL.

So the rule to handle string literals in the IDL used by thriftgo is:

* **Literals will be copied to the generated code with only delimiters they use unescaped.**

Take golang as an example, an IDL segment like

```thriftgo
const string str = "'double'\t\\"quoted\"" // double quoted

const string str2 = "\u65b0\u9f99\u6cc9\u5bfa"

struct S {
    1: string f1 = 'single\'"quoted' (go.tag = "json:\"hello\tworld\" vd:\"regexp('^[\\w\u4e00-\u9fa5 _]+$')\"")
    2: string f2 = "\u65b0\u9f99\u6cc9\u5bfa"
}
```

will be compiled into go code like

```go
const (
	Str = "'double'\t\\\"quoted\""

	Str2 = "\u65b0\u9f99\u6cc9\u5bfa"
)

type S struct {
	F1 string `thrift:"f1,1" json:"hello\tworld" vd:"regexp('^[\\w\u4e00-\u9fa5 _]+$')"`
	F2 string `thrift:"f2,2" json:"f2"`
}

func NewS() *S {
	return &S{

		F1: "single'\"quoted",
		F2: "\u65b0\u9f99\u6cc9\u5bfa",
	}
}
```

Note that it is unnecessary to escape a single quote in a double quoted literal or vice versa.
But for compatibility with existing IDLs, escaped double quotes in a quoted literal is allowed in go generator by default. It can be disabled by specifying `unescape_double_quote=false`.



