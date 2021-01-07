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

// Enum is a GraphQL Enum.
type Enum struct {
	Base

	// Values of the enum.
	values enumValueList
}

// Rank of the type.
func (t *Enum) Rank() int {
	return rankEnum
}

// AddValue is used to add values to the enum.
func (t *Enum) AddValue(ev *EnumValue) error {
	return t.values.add(ev)
}

// Values returns a list of EnumValues.
func (t *Enum) Values() []*EnumValue {
	return t.values.list
}

func (t *Enum) String() string {
	return t.SDL()
}

// SDL returns an SDL representation of the type.
func (t *Enum) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Enum) Write(w io.Writer, desc bool) (err error) {
	if err = t.writeHeader(w, "enum ", desc); err == nil {
		for _, ev := range t.values.list {
			if err = ev.Write(w, desc); err != nil {
				break
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte{'}', '\n'})
	}
	return
}

// Extend a type.
func (t *Enum) Extend(x Type) error {
	if ex, ok := x.(*Enum); ok { // Already checked so no need to report an error again.
		for _, ev := range ex.values.list {
			if err := t.values.add(ev); err != nil {
				return fmt.Errorf("%w: enum value %s on %s", err, ev.Value, t.N)
			}
		}
	}
	return t.Base.Extend(x)
}

// Validate a type.
func (t *Enum) Validate(root *Root) (errs []error) {
	errs = append(errs, root.validateTypeName("enum", t)...)
	// All members must be Objects and there must be at least one member.
	if 0 < t.values.Len() {
		for _, ev := range t.values.list {
			switch ev.Value {
			case Symbol("true"), Symbol("false"), Symbol("null"):
				errs = append(errs, fmt.Errorf("%w, %s is not a valid enum value for enum %s at %d:%d",
					ErrValidation, ev.Value, t.Name(), ev.line, ev.col))
			default:
				errs = append(errs, validateName(t.core, "enum value", string(ev.Value), ev.line, ev.col)...)
			}
			for _, du := range ev.Directives {
				errs = append(errs, root.validateDirUse(t.Name()+"."+string(ev.Value), Locate(ev), du)...)
			}
		}
	} else {
		errs = append(errs, fmt.Errorf("%w, enum %s must have at least one value at %d:%d",
			ErrValidation, t.Name(), t.line, t.col))
	}
	return
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (t *Enum) CoerceIn(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	if s, ok := v.(Symbol); ok {
		if t.values.has(s) {
			return s, nil
		}
		return nil, fmt.Errorf("%s is not a valid enum value in %s", s, t.Name())
	}
	if Relaxed {
		if s, ok := v.(string); ok {
			return Symbol(s), nil
		}
	}
	return nil, newCoerceErr(v, t.N)
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *Enum) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case Symbol:
		v = string(tv)
	case string:
		// ok as is
	default:
		err = newCoerceErr(tv, t.Name())
		v = nil
	}
	return v, err
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
func (t *Enum) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case kindStr:
		result = string(Locate(t))
	case nameStr:
		result = t.N
	case descriptionStr:
		result = t.Desc
	case enumValuesStr:
		if t.getBoolArg(args, "includeDeprecated") {
			result = &t.values
		} else {
			list := enumValueList{dict: map[string]*EnumValue{}}
			for _, ev := range t.values.list {
				if !ev.isDeprecated() {
					_ = list.add(ev)
				}
			}
			result = &list
		}
	case possibleTypesStr, fieldsStr, interfacesStr, inputFieldsStr, ofTypeStr:
		// nil result
	}
	return
}
