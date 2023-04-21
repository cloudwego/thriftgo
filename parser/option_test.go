package parser

import (
	"testing"
)

func TestOption(t *testing.T) {

	ast, err := ParseFile("option_idl/test.thrift", []string{"option_idl"}, true)
	assert(t, err == nil)

	p, err := getStruct("Person", ast)
	assert(t, err == nil, err)
	options, err := ParseStructOption(p.structLike, p.fromAst)
	assert(t, err == nil, err)

	opt := options["containerStruct"]
	assert(t, opt != nil)
	v, err := opt.GetFieldValue("valuei64")
	assert(t, err != nil)
	v, err = opt.GetFieldValue("valuemap")
	assert(t, err == nil)
	valuemap, ok := v.(map[interface{}]interface{})
	assert(t, ok)
	assert(t, len(valuemap) == 3)
	assert(t, valuemap["k1"] == "v1")
	assert(t, valuemap["k2"] == "v2")
	assert(t, valuemap["k3"] == "v3")

	v, err = opt.GetFieldValue("valuelist")
	assert(t, err == nil)
	valuelist, ok := v.([]interface{})
	assert(t, ok)
	assert(t, len(valuelist) == 4)
	assert(t, valuelist[0] == "a")
	assert(t, valuelist[1] == "b")
	assert(t, valuelist[2] == "c")
	assert(t, valuelist[3] == "d")

}

func TestParseOptionStr(t *testing.T) {
	st1 := "IsOdd=true"
	name, content, ok := parseOptionStr(st1)
	assert(t, ok && name == "IsOdd" && content == "true")

	st2 := "m2.IsOdd = true"
	name, content, ok = parseOptionStr(st2)
	assert(t, ok && name == "m2.IsOdd" && content == "true")

	st3 := " IsOdd= true"
	name, content, ok = parseOptionStr(st3)
	assert(t, ok && name == "IsOdd" && content == "true")

	st4 := "MyStruct={a:b c:d e=f}"
	name, content, ok = parseOptionStr(st4)
	assert(t, ok && name == "MyStruct" && content == "{a:b c:d e=f}")

	st5 := `MyStruct={
			a:b c:d
			e:f
			g:h
		}
	`
	name, content, ok = parseOptionStr(st5)
	assert(t, ok && name == "MyStruct" && content == "{    a:b c:d    e:f    g:h   }")
}

func TestParseKV(t *testing.T) {
	// basic test
	input := "{k1:v1 k2:[{kk1:vv1 kkk1:vvv1},{kk2:vv2}] k3:v3 k4:v4 k5:{kkkk1:kvvvv1}}"
	kv, err := parseKV(input)
	assert(t, err == nil && len(kv) == 5)
	assert(t, kv["k1"] == "v1")
	assert(t, kv["k2"] == "[{kk1:vv1 kkk1:vvv1},{kk2:vv2}]")
	assert(t, kv["k3"] == "v3")
	assert(t, kv["k4"] == "v4")
	assert(t, kv["k5"] == "{kkkk1:kvvvv1}")

	// space test
	input = "{k1:v1 \n  k2:[{kk1:vv1 kkk1:vvv1},{kk2:v  v2}] k3 : v3  \t k4: v4 k5:{  kkkk1 :kvvvv1}}"
	//input := "{k2:[{kk1:vv1 kkk1:vvv1},{kk2:v  v2}] k3 : v3}"
	kv, err = parseKV(input)
	assert(t, err == nil && len(kv) == 5)
	assert(t, kv["k1"] == "v1")
	assert(t, kv["k2"] == "[{kk1:vv1 kkk1:vvv1},{kk2:v v2}]")
	assert(t, kv["k3"] == "v3")
	assert(t, kv["k4"] == "v4")
	assert(t, kv["k5"] == "{kkkk1:kvvvv1}")

	// illegal test
	input = "{k1:}"
	kv, err = parseKV(input)
	assert(t, err != nil)
	input = "{k1:v1 k2:}"
	kv, err = parseKV(input)
	assert(t, err != nil)
	input = "{k1:v1 k2}"
	kv, err = parseKV(input)
	assert(t, err != nil)
	input = "{k1}"
	kv, err = parseKV(input)
	assert(t, err != nil)

	// simple test
	input = "{k1:\"v2\"}"
	kv, err = parseKV(input)
	assert(t, err == nil)

	input = "{k1:\"v2}"
	kv, err = parseKV(input)
	assert(t, err != nil)

	input = "{k1:'v2\"}"
	kv, err = parseKV(input)
	assert(t, err != nil)

	// bracket test
	input = "{k1:v2{}"
	kv, err = parseKV(input)
	assert(t, err != nil)

	input = "{\n        valuemap:{k1:v1 k2:v2 k3:}\n        valuelist:[a,b,c,d]\n        valueset:[{email:e1},{email:e2}]\n        valuelistset:[[a,b,c],[d,e,f]]\n        valuelistsetstruct:[[{email:e1},{email:e2}],[{email:e3},{email:e4}]]\n        valuemapStruct:[k1:{email:e1} k2:{email:e2}]\n    }"
	kv, err = parseKV(input)
	assert(t, err == nil)

}

func TestParseArr(t *testing.T) {

	input := "[a,b,c]"
	arr, err := parseArr(input)
	assert(t, err == nil && len(arr) == 3)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "b")
	assert(t, arr[2] == "c")

	input = "[a,\"b\",c]"
	arr, err = parseArr(input)
	assert(t, err == nil && len(arr) == 3)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "\"b\"")
	assert(t, arr[2] == "c")

	input = "[a,'b,c]"
	arr, err = parseArr(input)
	assert(t, err != nil)

	input = "[a,[b,c],c]"
	arr, err = parseArr(input)
	assert(t, err == nil && len(arr) == 3)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "[b,c]")
	assert(t, arr[2] == "c")

	input = "[a,[b,{c,d}],c,{e,f}]"
	arr, err = parseArr(input)
	assert(t, err == nil && len(arr) == 4)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "[b,{c,d}]")
	assert(t, arr[2] == "c")
	assert(t, arr[3] == "{e,f}")

	input = "[a ,[  b , {c, d} ],\nc,\t{e,f} ]"
	arr, err = parseArr(input)
	assert(t, err == nil && len(arr) == 4)
	assert(t, arr[0] == "a")
	assert(t, arr[1] == "[b,{c,d}]")
	assert(t, arr[2] == "c")
	assert(t, arr[3] == "{e,f}")

}

func assert(t *testing.T, cond bool, val ...interface{}) {
	t.Helper()
	if !cond {
		if len(val) > 0 {
			val = append([]interface{}{"assertion failed:"}, val...)
			t.Fatal(val...)
		} else {
			t.Fatal("assertion failed")
		}
	}
}
