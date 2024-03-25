/**
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fieldmask

// ValueMapper is used to mapping values
// type ValueMapper struct {
// 	// AllowNested indicates if a recursive type (LIST/SET/MAP) is acceptable to this.
// 	// If it is true, every elem (both key for MAP) will trigger `OnXX` mapping function
// 	AllowRecurse bool

// 	//mapping functions
// 	OnInt    func(isNil bool, val int) (int, bool)
// 	OnFloat  func(isNil bool, val float64) (int, bool)
// 	OnBool   func(isNil bool, val bool) (bool, bool)
// 	OnString func(isNil bool, val string) (string, bool)
// }

// PathNapper is the definition of a ValueMapper for specific path
// type PathMapper struct {
// 	Path   string
// 	Mapper ValueMapper
// }

// func NewFieldMapper(desc thrift_reflection.TypeDescriptor, maps ...PathMapper) *FieldMask
