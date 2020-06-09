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

// Union is a GraphQL Union.
type Union struct {
	Base
	Members []Type
}

// Rank of the type.
func (t *Union) Rank() int {
	return rankUnion
}

// String representation of the type.
func (t *Union) String() string {
	return t.SDL()
}

// SDL returns an SDL representation of the type.
func (t *Union) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Union) Write(w io.Writer, desc bool) (err error) {
	if err = writeDesc(w, t.Desc, 0, desc); err == nil {
		if _, err = w.Write([]byte("union ")); err == nil {
			if _, err = w.Write([]byte(t.N)); err == nil {
				err = writeDirectiveUses(w, t.Dirs)
				if err == nil {
					_, err = w.Write([]byte{' ', '=', ' '})
					for i, m := range t.Members {
						name := m.Name()
						if 0 < i {
							name = " | " + name
						}
						if _, err = w.Write([]byte(name)); err != nil {
							break
						}
					}
				}
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte{'\n'})
	}
	return
}

// Extend a type.
func (t *Union) Extend(x Type) error {
	if ux, ok := x.(*Union); ok { // Already checked so no need to report an error again.
		for _, m := range ux.Members {
			for _, exist := range t.Members {
				if m.Name() == exist.Name() {
					return fmt.Errorf("%w: union member %s already exists on %s", ErrDuplicate, m.Name(), t.N)
				}
			}
			t.Members = append(t.Members, m)
		}
	}
	return t.Base.Extend(x)
}

// Validate a type.
func (t *Union) Validate(root *Root) (errs []error) {
	// All members must be Objects and there must be at least one member.
	if 0 < len(t.Members) {
		for _, m := range t.Members {
			if _, ok := m.(*Object); !ok {
				errs = append(errs, fmt.Errorf("%w, %s can not be a union member since it is a %T, not an *ggql.Object at %d:%d",
					ErrValidation, m.Name(), m, t.line, t.col))
			}
		}
	} else {
		errs = append(errs, fmt.Errorf("%w, union %s must have at least one member at %d:%d",
			ErrValidation, t.Name(), t.line, t.col))
	}
	return
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
func (t *Union) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case kindStr:
		result = string(Locate(t))
	case nameStr:
		result = t.N
	case descriptionStr:
		result = t.Desc
	case possibleTypesStr:
		list := newTypeList()
		list.add(t.Members...)
		result = list
	case fieldsStr, interfacesStr, enumValuesStr, inputFieldsStr, ofTypeStr:
		// nil result
	}
	return
}
