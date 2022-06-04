// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

namespace * meta

enum TMessageType {
	INVALID_MESSAGE_TYPE
	CALL
	REPLY
	EXCEPTION
	ONEWAY
}

enum TTypeID {
	STOP   = 0
	VOID   = 1
	BOOL   = 2
	BYTE   = 3
	DOUBLE = 4
	I16    = 6
	I32    = 8
	I64    = 10
	STRING = 11
	STRUCT = 12
	MAP    = 13
	SET    = 14
	LIST   = 15
	UTF8   = 16
	UTF16  = 17
}

enum TRequiredness {
	DEFAULT
	REQUIRED
	OPTIONAL
}

struct TypeMeta {
    1: required TTypeID type_id
    2: optional TypeMeta key_type
    3: optional TypeMeta value_type
}

struct FieldMeta {
    1: required i16 field_id
    2: required string name
    3: required TRequiredness requiredness
    4: required TypeMeta field_type
}

struct StructMeta {
   1: required string name
   2: required string category // "struct", "union" or "exception"
   3: required list<FieldMeta> fields
}
