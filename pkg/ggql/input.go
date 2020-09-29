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
	"strings"
)

// Input is a GraphQL InputObject type.
type Input struct {
	Base

	// Fields in the input object.
	fields inputFieldList
}

// Rank of the type.
func (t *Input) Rank() int {
	return rankInput
}

// String representation of the type.
func (t *Input) String() string {
	return t.SDL()
}

// AddField is used to add fields to an interface.
func (t *Input) AddField(f *InputField) error {
	return t.fields.add(f)
}

// Fields returns a list of all fields in the input object.
func (t *Input) Fields() []*InputField {
	return t.fields.list
}

// SDL returns an SDL representation of the type.
func (t *Input) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Input) Write(w io.Writer, desc bool) (err error) {
	if err = t.writeHeader(w, "input ", desc); err == nil {
		for _, f := range t.fields.list {
			if err = f.Write(w, desc); err != nil {
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
func (t *Input) Extend(x Type) error {
	if ix, ok := x.(*Input); ok { // Already checked so no need to report an error again.
		for k, f := range ix.fields.dict {
			if err := t.fields.add(f); err != nil {
				return fmt.Errorf("%w: field %s on %s", err, k, t.N)
			}
		}
	}
	return t.Base.Extend(x)
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (t *Input) CoerceIn(v interface{}) (interface{}, error) {
	switch tv := v.(type) {
	case nil:
		// nil is okay at this point
	case map[string]interface{}:
		for k := range tv {
			if t.fields.get(k) == nil {
				return nil, fmt.Errorf("%w %s is not in %s", ErrCoerce, k, t.Name())
			}
		}
		for k, f := range t.fields.dict {
			ov := tv[k]
			if ov == nil && f.Default != nil { // if not set then add the default value if not nil
				tv[k] = f.Default
			} else if co, _ := f.Type.(InCoercer); co != nil {
				if cv, err := co.CoerceIn(ov); err == nil {
					tv[k] = cv
				} else {
					return nil, err
				}
			} else {
				return nil, newCoerceErr(ov, f.Type.Name())
			}
		}
	default:
		return nil, newCoerceErr(v, t.Name())
	}
	return v, nil
}

// Validate a type.
func (t *Input) Validate(root *Root) (errs []error) {
	if 0 < t.fields.Len() { // must have at least one field
		for _, f := range t.fields.list {
			if strings.HasPrefix(f.Name(), "__") {
				errs = append(errs, fmt.Errorf("%w, %s is not a valid field name, it begins with '__' at %d:%d",
					ErrValidation, f.Name(), f.line, f.col))
			}
			if !IsInputType(f.Type) {
				errs = append(errs, fmt.Errorf("%w, %s does not return an input type at %d:%d",
					ErrValidation, f.Name(), f.line, f.col))
			}
		}
	} else {
		errs = append(errs, fmt.Errorf("%w, input object %s must have at least one field at %d:%d",
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
func (t *Input) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case kindStr:
		result = string(Locate(t))
	case nameStr:
		result = t.N
	case descriptionStr:
		result = t.Desc
	case inputFieldsStr:
		result = &t.fields
	case possibleTypesStr, interfacesStr, enumValuesStr, fieldsStr, ofTypeStr:
		// return nil
	}
	return
}
