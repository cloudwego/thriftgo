// Copyright 2021 CloudWeGo Authors
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

package golang

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cloudwego/thriftgo/generator/golang/common"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/semantic"
)

// GetServiceIdentifier returns the identifier without a package prefix.
func (cu *CodeUtils) GetServiceIdentifier(id string) (string, error) {
	if id == "" {
		return "", nil
	}
	parts := semantic.SplitType(id)
	switch len(parts) {
	case 1:
		return cu.identify0(parts[0])
	case 2:
		return cu.identify0(parts[1])
	default:
		return "", fmt.Errorf("invalid service name: '%s'", id)
	}
}

// GetNewFunc returns the construction function of the given type.
func (cu *CodeUtils) GetNewFunc(typeName string) string {
	if idx := strings.Index(typeName, "."); idx >= 0 {
		idx++
		return typeName[:idx] + "New" + typeName[idx:]
	}
	return "New" + typeName
}

// ID returns the ID of a field. If the ID is a minus number, the slash will be replaced by an underscore.
func (cu *CodeUtils) ID(f *parser.Field) string {
	id := fmt.Sprint(f.ID)
	return strings.ReplaceAll(id, "-", "_")
}

// Deref removes the "&" and "*" prefix of the given code.
func (cu *CodeUtils) Deref(code string) string {
	return strings.TrimLeft(code, "&*")
}

// IsExported determines whether a name is exported.
func (cu *CodeUtils) IsExported(name string) bool {
	for _, r := range name {
		return unicode.IsUpper(r)
	}
	return false
}

// Unexport returns an unexported form of the given identifier.
func (cu *CodeUtils) Unexport(name string) (string, error) {
	return common.LowerFirstRune(name), nil
}

// GetFilePath returns a path to the generated file for the given IDL.
// Note that the result is a path relative to the root output path.
func (cu *CodeUtils) GetFilePath(t *parser.Thrift) string {
	ref, _, pth := cu.ParseNamespace(t)
	full := filepath.Join(pth, ref+".go")
	if strings.HasSuffix(full, "_test.go") {
		full = strings.ReplaceAll(full, "_test.go", "_test_.go")
	}
	return full
}

// GetPackageName returns a go package name for the given thrift AST.
func (cu *CodeUtils) GetPackageName(ast *parser.Thrift) string {
	namespace := ast.GetNamespaceOrReferenceName("go")
	return cu.NamespaceToPackage(namespace)
}

// NamespaceToPackage converts a namespace to a package.
func (cu *CodeUtils) NamespaceToPackage(ns string) string {
	parts := strings.Split(ns, ".")
	return strings.ToLower(parts[len(parts)-1])
}

// NamespaceToImportPath returns an import path for the given namespace.
// Note that the result will not have the package prefix set with SetPackagePrefix.
func (cu *CodeUtils) NamespaceToImportPath(ns string) string {
	pkg := strings.ReplaceAll(ns, ".", "/")
	return pkg
}

// getImport returns the package name and an import path for the given AST.
// If the package prefix is set, the import path will contain it.
func (cu *CodeUtils) getImport(t *parser.Thrift) (pkg, pth string) {
	ns := t.GetNamespaceOrReferenceName("go")
	pkg = cu.NamespaceToPackage(ns)
	pth = cu.NamespaceToImportPath(ns)
	if cu.packagePrefix != "" {
		pth = filepath.Join(cu.packagePrefix, pth)
	}
	return
}

// GenTags generates go tags for the given parser.Field.
func (cu *CodeUtils) GenTags(f *parser.Field, insertPoint string) (string, error) {
	var tags []string
	if f.Requiredness == parser.FieldType_Required {
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d,required"`, f.Name, f.ID))
	} else {
		tags = append(tags, fmt.Sprintf(`thrift:"%s,%d"`, f.Name, f.ID))
	}

	if gotags := f.Annotations.Get("go.tag"); len(gotags) > 0 {
		tags = append(tags, gotags[0])
	} else {
		if cu.Features().GenDatabaseTag {
			tags = append(tags, fmt.Sprintf(`db:"%s"`, f.Name))
		}

		if f.Requiredness.IsOptional() && cu.Features().GenOmitEmptyTag {
			tags = append(tags, fmt.Sprintf(`json:"%s,omitempty"`, f.Name))
		} else {
			tags = append(tags, fmt.Sprintf(`json:"%s"`, f.Name))
		}
	}
	str := fmt.Sprintf("`%s%s`", strings.Join(tags, " "), insertPoint)
	return str, nil
}

// GetTypeIDConstant returns the thrift type ID literal for the given type which
// is suitable to concate with "thrift." to produce a valid type ID constant.
func (cu *CodeUtils) GetTypeIDConstant(t *parser.Type) string {
	tid := cu.GetTypeID(t)
	tid = strings.ToUpper(tid)
	if tid == "BINARY" {
		tid = "STRING"
	}
	return tid
}

// GetTypeID returns the thrift type ID literal for the given type which is suitable
// to concate with "Read" or "Write" to produce a valid method name in the TProtocol
// interface. Note that enum types results in I32.
func (cu *CodeUtils) GetTypeID(t *parser.Type) string {
	// Bool|Byte|I16|I32|I64|Double|String|Binary|Set|List|Map|Struct
	return category2TypeID[t.Category]
}
