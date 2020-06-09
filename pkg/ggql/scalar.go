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
	"io"
)

// Scalar is GraphQL scalar.
type Scalar struct {
	Base
}

// Rank of the type.
func (t *Scalar) Rank() int {
	return rankScalar
}

// String representation of the type.
func (t *Scalar) String() string {
	return t.N
}

// SDL returns an SDL representation of the type.
func (t *Scalar) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Scalar) Write(w io.Writer, desc bool) (err error) {
	if err = writeDesc(w, t.Desc, 0, desc); err == nil {
		if _, err = w.Write([]byte("scalar ")); err == nil {
			if _, err = w.Write([]byte(t.Name())); err == nil {
				err = writeDirectiveUses(w, t.Dirs)
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte{'\n'})
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
func (t *Scalar) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case kindStr:
		result = string(Locate(t))
	case nameStr:
		result = t.N
	case descriptionStr:
		result = t.Desc
	case possibleTypesStr, fieldsStr, interfacesStr, enumValuesStr, inputFieldsStr, ofTypeStr:
		// nil result
	}
	return
}
