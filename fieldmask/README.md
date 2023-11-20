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

# ThriftPath RFC

## What is thrift path?
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

## Syntax
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
[`index`,index...] | get any number of elements in an List/Set corepsonding to indices. Indices must be integer.For example: `$.FieldList[1,3,4]` .Notice: a index beyond actual list size can written but is useless.
{"key","key"...} | get any number of values corepsonding to key in a string-typed-key map. For example: `$.StrMap{"abcd","1234"}` 
{id,id...} | get the child field with specific id in a integer-typed-key map. For example, `$.IntMap{1,2}` 
\* | get **ALL** fields/elements, that is: `$.StrMap{*}.FieldX` menas gets all the elements' FieldX in a map Root.StrMap; `$.List[*].FieldX` means get all the elements' FieldX in a list Root.List
</byte-sheet-html-origin><!--EndFragment-->

