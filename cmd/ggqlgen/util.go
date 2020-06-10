// Copyright 2019-2020 University Health Network
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"strings"

	"github.com/uhn/ggql/pkg/ggql"
)

func typeStr(t ggql.Type) string {
	// Handle the core scalars first.
	switch t.Name() {
	case "String", "ID":
		return "string"
	case "Int":
		return "int"
	case "Float":
		return "float32"
	case timeStr:
		return "time.Time"
	case "Boolean":
		return "bool"
	}
	switch tt := t.(type) {
	case *ggql.Enum:
		return "string"
	case *ggql.Input, *ggql.Object:
		return "*" + t.Name()
	case *ggql.List:
		return "[]" + typeStr(tt.Base)
	case *ggql.NonNull:
		return typeStr(tt.Base)
	}
	return "interface{}"
}

func publicName(s string) string {
	public := strings.Title(s)
	public = strings.ReplaceAll(public, "Id", "ID")
	public = strings.ReplaceAll(public, "id", "ID")
	return public
}
