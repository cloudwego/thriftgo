// Copyright 2021 CloudWeGo
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

package styles

// Naming determine naming style of the identifier converted from IDL.
type Naming interface {
	Identify(name string) (string, error)

	UseInitialisms(enable bool)
}

// NamingStyles returns all supported naming styles.
func NamingStyles() []string {
	return []string{
		"golint", "apache", "thriftgo",
	}
}

// NewNamingStyle creates a Naming with the given name.
// If the given name is supported, this function returns nil.
func NewNamingStyle(name string) Naming {
	switch name {
	case "golint":
		return new(GoLint)
	case "apache":
		return new(Apache)
	case "thriftgo":
		return new(ThriftGo)
	default:
		return nil
	}
}
