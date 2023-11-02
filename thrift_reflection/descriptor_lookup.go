// Copyright 2023 CloudWeGo Authors
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

package thrift_reflection

import "errors"

func (gd *GlobalDescriptor) LookupFD(filepath string) *FileDescriptor {
	if gd == nil {
		return nil
	}
	return gd.globalFD[filepath]
}

func (gd *GlobalDescriptor) LookupEnum(name, filepath string) *EnumDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetEnumDescriptor(name)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetEnumDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupConst(name, filepath string) *ConstDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetConstDescriptor(name)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetConstDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupTypedef(alias, filepath string) *TypedefDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetTypedefDescriptor(alias)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetTypedefDescriptor(alias)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupStruct(name, filepath string) *StructDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetStructDescriptor(name)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetStructDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupUnion(name, filepath string) *StructDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetUnionDescriptor(name)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetUnionDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupException(name, filepath string) *StructDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetExceptionDescriptor(name)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetExceptionDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupService(name, filepath string) *ServiceDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetServiceDescriptor(name)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetServiceDescriptor(name)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupMethod(method, service, filepath string) *MethodDescriptor {
	if gd == nil {
		return nil
	}
	if filepath != "" {
		return gd.LookupFD(filepath).GetMethodDescriptor(service, method)
	}
	for _, fd := range gd.globalFD {
		val := fd.GetMethodDescriptor(service, method)
		if val != nil {
			return val
		}
	}
	return nil
}

func (gd *GlobalDescriptor) LookupIncludedStructsFromMethod(method *MethodDescriptor) ([]*StructDescriptor, error) {
	if gd == nil {
		return nil, errors.New("global descriptor is nil")
	}
	structMap := map[*StructDescriptor]bool{}
	typeArr := []*TypeDescriptor{}
	typeArr = append(typeArr, method.GetResponse())
	for _, arg := range method.GetArgs() {
		typeArr = append(typeArr, arg.GetType())
	}
	for _, typeDesc := range typeArr {
		err := gd.lookupIncludedStructsFromType(typeDesc, structMap)
		if err != nil {
			return nil, err
		}
	}
	structArr := make([]*StructDescriptor, 0, len(structMap))
	for st := range structMap {
		structArr = append(structArr, st)
	}
	return structArr, nil
}

// LookupIncludedStructsFromStruct finds all struct descriptor included by this structDescriptor (and current struct descriptor is also included in the return result)
func (gd *GlobalDescriptor) LookupIncludedStructsFromStruct(sd *StructDescriptor) ([]*StructDescriptor, error) {
	if gd == nil {
		return nil, errors.New("global descriptor is nil")
	}
	structMap := map[*StructDescriptor]bool{}
	err := gd.lookupIncludedStructsFromStruct(sd, structMap)
	if err != nil {
		return nil, err
	}
	structArr := make([]*StructDescriptor, 0, len(structMap))
	for st := range structMap {
		structArr = append(structArr, st)
	}
	return structArr, nil
}

func (gd *GlobalDescriptor) LookupIncludedStructsFromType(td *TypeDescriptor) ([]*StructDescriptor, error) {
	if gd == nil {
		return nil, errors.New("global descriptor is nil")
	}
	structMap := map[*StructDescriptor]bool{}
	err := gd.lookupIncludedStructsFromType(td, structMap)
	if err != nil {
		return nil, err
	}
	structArr := make([]*StructDescriptor, 0, len(structMap))
	for st := range structMap {
		structArr = append(structArr, st)
	}
	return structArr, nil
}

func (gd *GlobalDescriptor) lookupIncludedStructsFromType(typeDesc *TypeDescriptor, structMap map[*StructDescriptor]bool) error {
	if gd == nil {
		return errors.New("global descriptor is nil")
	}
	if typeDesc.IsStruct() {
		stDesc, err := typeDesc.GetStructDescriptor()
		if err != nil {
			return err
		}
		return gd.lookupIncludedStructsFromStruct(stDesc, structMap)
	}
	if typeDesc.IsContainer() {
		err := gd.lookupIncludedStructsFromType(typeDesc.GetValueType(), structMap)
		if err != nil {
			return err
		}
		if typeDesc.IsMap() {
			er := gd.lookupIncludedStructsFromType(typeDesc.GetKeyType(), structMap)
			if er != nil {
				return er
			}
		}
	}
	return nil
}

func (gd *GlobalDescriptor) lookupIncludedStructsFromStruct(sd *StructDescriptor, structMap map[*StructDescriptor]bool) error {
	if gd == nil {
		return errors.New("global descriptor is nil")
	}
	if structMap[sd] {
		return nil
	}
	structMap[sd] = true
	for _, f := range sd.GetFields() {
		typeDesc := f.GetType()
		err := gd.lookupIncludedStructsFromType(typeDesc, structMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func LookupFD(filepath string) *FileDescriptor {
	return defaultGlobalDescriptor.LookupFD(filepath)
}

func LookupEnum(name, filepath string) *EnumDescriptor {
	return defaultGlobalDescriptor.LookupEnum(name, filepath)
}

func LookupConst(name, filepath string) *ConstDescriptor {
	return defaultGlobalDescriptor.LookupConst(name, filepath)
}

func LookupTypedef(alias, filepath string) *TypedefDescriptor {
	return defaultGlobalDescriptor.LookupTypedef(alias, filepath)
}

func LookupStruct(name, filepath string) *StructDescriptor {
	return defaultGlobalDescriptor.LookupStruct(name, filepath)
}

func LookupUnion(name, filepath string) *StructDescriptor {
	return defaultGlobalDescriptor.LookupUnion(name, filepath)
}

func LookupException(name, filepath string) *StructDescriptor {
	return defaultGlobalDescriptor.LookupException(name, filepath)
}

func LookupService(name, filepath string) *ServiceDescriptor {
	return defaultGlobalDescriptor.LookupService(name, filepath)
}

func LookupMethod(method, service, filepath string) *MethodDescriptor {
	return defaultGlobalDescriptor.LookupMethod(method, service, filepath)
}

func LookupIncludedStructsFromMethod(method *MethodDescriptor) ([]*StructDescriptor, error) {
	return defaultGlobalDescriptor.LookupIncludedStructsFromMethod(method)
}

// LookupIncludedStructsFromStruct finds all struct descriptor included by this structDescriptor (and current struct descriptor is also included in the return result)
func LookupIncludedStructsFromStruct(sd *StructDescriptor) ([]*StructDescriptor, error) {
	return defaultGlobalDescriptor.LookupIncludedStructsFromStruct(sd)
}

func LookupIncludedStructsFromType(td *TypeDescriptor) ([]*StructDescriptor, error) {
	return defaultGlobalDescriptor.LookupIncludedStructsFromType(td)
}
