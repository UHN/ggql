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
	"fmt"
	"io"
)

// NonNull is the NonNull GraphQL type.
type NonNull struct {

	// Base type for the non-null type.
	Base Type

	line int
	col  int
}

// Rank of the type.
func (t *NonNull) Rank() int {
	return rankHidden
}

// Core returns true if the type is one of the built in types.
func (t *NonNull) Core() bool {
	return false
}

// Name returns the name of the type.
func (t *NonNull) Name() string {
	return t.Base.Name() + "!"
}

// Description returns the description of the type.
func (t *NonNull) Description() string {
	return ""
}

// Directives returns the directive associated with the type.
func (t *NonNull) Directives() []*DirectiveUse {
	return nil
}

// Line the type was defined on in the schema.
func (t *NonNull) Line() int {
	return t.line
}

// Column the type was defined on in the schema.
func (t *NonNull) Column() int {
	return t.col
}

// String representation of the type.
func (t *NonNull) String() string {
	return t.Name()
}

// SDL returns an SDL representation of the type.
func (t *NonNull) SDL(desc ...bool) string {
	return ""
}

// Write the type as SDL.
func (t *NonNull) Write(w io.Writer, desc bool) error {
	return nil
}

// Extend a type.
func (t *NonNull) Extend(x Type) error {
	return nil
}

// Validate a type.
func (t *NonNull) Validate(root *Root) (errs []error) {
	return
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (t *NonNull) CoerceIn(v interface{}) (interface{}, error) {
	if co, _ := t.Base.(InCoercer); co != nil && v != nil {
		return co.CoerceIn(v)
	}
	return nil, newCoerceErr(v, t.Name())
}

// CoerceOut coerces a result value into a value suitable for output.
func (t *NonNull) CoerceOut(v interface{}) (interface{}, error) {
	if co, _ := t.Base.(OutCoercer); co != nil && v != nil {
		return co.CoerceOut(v)
	}
	return nil, newCoerceErr(v, t.Name())
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
func (t *NonNull) Resolve(field *Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case kindStr, descriptionStr:
		return "NON_NULL", nil
	case nameStr:
		return t.Name(), nil
	case ofTypeStr:
		return t.Base, nil
	case interfacesStr, fieldsStr, possibleTypesStr, enumValuesStr, inputFieldsStr:
		return nil, nil
	}
	return nil, fmt.Errorf("type __Type does not have field %s", field)
}
