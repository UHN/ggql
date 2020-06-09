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

package ggql

import (
	"bytes"
	"fmt"
	"io"
)

// Interface is a GraphQL Interface.
type Interface struct {
	Base

	Root *Root // needed to get possibleTypes

	// Fields in the interface.
	fields fieldList
}

// Rank of the type.
func (t *Interface) Rank() int {
	return rankInterface
}

// String representation of the type.
func (t *Interface) String() string {
	return t.SDL()
}

// SDL returns an SDL representation of the type.
func (t *Interface) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Interface) Write(w io.Writer, desc bool) (err error) {
	if err = t.writeHeader(w, "interface ", desc); err == nil {
		for _, fd := range t.fields.list {
			if err = fd.Write(w, desc); err != nil {
				break
			}
		}
		if err == nil {
			_, err = w.Write([]byte{'}', '\n'})
		}
	}
	return
}

// Extend a type.
func (t *Interface) Extend(x Type) error {
	if ix, ok := x.(*Interface); ok { // Already checked so no need to report an error again.
		for _, f := range ix.fields.list {
			if err := t.fields.add(f); err != nil {
				return fmt.Errorf("%w: on %s", err, t.N)
			}
		}
	}
	return t.Base.Extend(x)
}

// GetField returns the field matching the name or nil if not found.
func (t *Interface) GetField(name string) *FieldDef {
	return t.fields.get(name)
}

// Validate a type.
func (t *Interface) Validate(root *Root) (errs []error) {
	return append(errs, t.validateFieldDefs(t.Name(), &t.fields)...)
}

// AddField is used to add fields to an interface.
func (t *Interface) AddField(fd *FieldDef) error {
	return t.fields.add(fd)
}

// Resolve returns one of the following:
//   kind: __TypeKind!
//   name: String
//   description: String
//   fields(includeDeprecated: Boolean = false): [__Field!]
//   interfaces: [__Type!]
//   possibleTypes: [__Type!]
//   enumValues(includeDeprecated: Boolean = false): [__EnumValue!]
//   inputfields: [__InputValue!]
//   ofType: __Type
func (t *Interface) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case kindStr:
		result = string(Locate(t))
	case nameStr:
		result = t.N
	case descriptionStr:
		result = t.Desc
	case fieldsStr:
		result = &t.fields
	case possibleTypesStr:
		result = t.possibleTypes()
	case interfacesStr, enumValuesStr, inputFieldsStr, ofTypeStr:
		// nil result
	}
	return
}

func (t *Interface) possibleTypes() *typeList {
	list := newTypeList()

	for _, pt := range t.Root.types.list {
		if obj, _ := pt.(*Object); obj != nil {
			for _, i := range obj.Interfaces {
				if t == i {
					list.add(obj)
				}
			}
		}
	}
	return list
}
