<!--
 Copyright 2023 CloudWeGo Authors
 
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 
     http://www.apache.org/licenses/LICENSE-2.0
 
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->
# Thrift FieldMask RFC

## What is thrift fieldmask?
FieldMask is inspired by [Protobuf](https://protobuf.dev/reference/protobuf/google.protobuf/#field-mask) and used to indicates the data that users care about, and filter out useless data, during a RPC call, in order to reducing network package size and accelerating serializing/deserializing process. This tech has been widely used among Protobuf [services](https://netflixtechblog.com/practical-api-design-at-netflix-part-1-using-protobuf-fieldmask-35cfdc606518).

## How to construct a fieldmask?
To construct a fieldmask, you need two things: 
 - [Thrift Path](#thrift-path) for describing the data you want
 - [Type Descriptor](#type-descriptor) for validating the thrift path you pass is compatible with thrift message definition (IDL)

### Thrift Path

#### What is thrift path?
A path string represents a arbitrary endpoint of thrift object. It is used for locating data from thrift root message, and defined from-top-to-bottom.
For exapmle, a thrift message defined as below:
```thrift
struct Example {
    1: string Foo,
    2: i64 Bar
    3: Example Self
}
```
A thrift path `$.Foo` represents the string value of Example.Foo, and `$.Self.Bar` represents the secondary layer i64 value of Example.Self.Bar
Since thrift has four nesting types (LIST/SET/MAP/STRUCT), thrift path should also support locating elements in all these types' object, not only STRUCT.

#### Syntax
Here are basic hypothesis:
- `fieldname` is the field name of a field in a struct, it **MUST ONLY** contain '[a-zA-Z]' alphabet letters, integer numbers and char '_'.
- `index` is the index of a element in a list or set, it **MUST ONLY** contain integer numbers.
- `key` is the string-typed key of a element in a map, it can contain any letters, but it **MUST** be a quoted string.
- `id` is the integer-typed key of a element in a map, it **MUST ONLY** contain integer numbers.

Here is detailed syntax:
<!--StartFragment--><byte-sheet-html-origin data-id="1700208276535" data-version="4" data-is-embed="true" data-grid-line-hidden="false" data-copy-type="col">
ThriftPath | Description
-- | --
$ | the root object,every path must start with it.
.`fieldname` | get the child field of a struct corepsonding to fieldname. For example, `$.FieldA.ChildrenB`
[`index`,`index`...] | get any number of elements in an List/Set corepsonding to indices. Indices must be integer.For example: `$.FieldList[1,3,4]` .Notice: a index beyond actual list size can written but is useless.
{"`key`","`key`"...} | get any number of values corepsonding to key in a string-typed-key map. For example: `$.StrMap{"abcd","1234"}` 
{`id`,`id`...} | get the child field with specific id in a integer-typed-key map. For example, `$.IntMap{1,2}` 
\* | get **ALL** fields/elements, that is: `$.StrMap{*}.FieldX` menas gets all the elements' FieldX in a map Root.StrMap; `$.List[*].FieldX` means get all the elements' FieldX in a list Root.List. 
</byte-sheet-html-origin><!--EndFragment-->

#### Agreement Of Implementation
- A field in mask means "PASS" (**will be** serialized/deserialized),  and the other field not in mask means "Filtered" ((**won't be** serialized/deserialized))
- A empty mask means "PASS ALL" (all field is "PASS")
- For map of neither-string-nor-integer typed key, only syntax token of all '*' (see above) is allowed in

### Type Descriptor
Type descriptor is the runtime representation of a message definition, in aligned with [Protobuf Descriptor](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto). To get a type descriptor, you must enable thrift reflection feature first, which was introduced in thriftgo [v0.3.0](https://github.com/cloudwego/thriftgo/pull/83). you can generate related codes for this feature using option `with_reflection`.

## How to use fieldmask?
1. First, you must generates codes for this feature using two options `with_fieldmask` and `with_reflection`
```
$ thriftgo -g with_field_mask,with_reflection ${your_idl}
```
2. Create a fieldmask in the initializing phase of your application (recommanded), or just in the bizhandler before you return a response
```go
import (
	"sync"
	"github.com/cloudwego/thriftgo/fieldmask"
	nbase "github.com/cloudwego/thriftgo/test/golang/fieldmask/gen-new/base"
)

var fieldmaskCache sync.Map

func init() {
	// new a obj to get its TypeDescriptor
	obj := nbase.NewBase()
    desc := obj.GetTypeDescriptor()

	// construct a fieldmask with TypeDescriptor and thrift pathes
	fm, err := fieldmask.NewFieldMask(desc,
		"$.Addr", "$.LogID", "$.TrafficEnv.Code", "$.Meta.IntMap{1}", "$.Meta.StrMap{\"1234\"}", "$.Meta.List[1]", "$.Meta.Set[1]")
	if err != nil {
		panic(err)
	}

	// cache it for future usage of nbase.Base
	fieldmaskCache.Store("Mask1ForBase", fm)
}
```
3. Now you can use fieldmask in both client-side and server-side
  - For server-side, you can set fieldmask with generated API `Set_FieldMask()` on your response object. Then the object itself will notice the fieldmask and using it during its serialization
  ```go
  func bizHandler(req any) (*nbase.Base) {
    // handle request ...

	// biz logic: handle and get final response object
	obj := bizBase()

	// Load fieldmask from cache
	fm, _ := fieldmaskCache.Load("Mask1ForBase")
	if fm != nil {
		// load ok, set fieldmask onto the object using codegen API
		obj.Set_FieldMask(fm.(*fieldmask.FieldMask))
	}

	return obj
  }
  ```
  - For client-side: related to the deserialization process of framework. For kitex, it's WIP.

## How to pass fieldmask between programs?
Generally, you can add one binary field on your request definition to carry fieldmask, and explicitly serialize/deserialize the fieldmask you are using into/from this field. We provide two encapsulated API for serialization/deserialization:
- [FieldMask.MarshalJSON()/UnmarshalJSON()](serdes.go): Object methods, serialize/deserialize fieldmask into/from json bytes
- [thriftgo/fieldmask.Marshal()/Unmarshal()](serdes.go): Package functions, serialize/deserialize fieldmask into/from binary bytes. We recommand you to use this API rather than the last one, because it is **much faster** due to using cache -- Unless your application is lack of memory.


## Benchmark
See [(main_test.go)](../test/golang/fieldmask/main_test.go)
```
goos: darwin
goarch: amd64
pkg: github.com/cloudwego/thriftgo/test/golang/fieldmask
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkWriteWithFieldMask/old-16               1447 ns/op               0 B/op          0 allocs/op
BenchmarkWriteWithFieldMask/new-16               1513 ns/op               0 B/op          0 allocs/op
BenchmarkWriteWithFieldMask/new-mask-half-16      710.4 ns/op             0 B/op          0 allocs/op
BenchmarkReadWithFieldMask/old-16                3832 ns/op            1628 B/op         36 allocs/op
BenchmarkReadWithFieldMask/new-16                3697 ns/op            1708 B/op         36 allocs/op
BenchmarkReadWithFieldMask/new-mask-half-16      3436 ns/op            1212 B/op         27 allocs/op
```
Explain case names:
- Write: serialization test
- Read: deserializtion test
- old: not generate with_fieldmask API
- new: generate with_fieldmask API, but not use fieldmask
- new-mask-half: generate with_fieldmask API and use fieldmask to mask half of the data