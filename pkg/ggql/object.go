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
	"reflect"
)

// Object is a GraphQL Object.
type Object struct {
	Base

	// Interfaces the object supports.
	Interfaces []Type

	// Fields in the object.
	fields fieldList

	meta reflect.Type
}

// Rank of the type.
func (t *Object) Rank() int {
	switch t.N {
	case "Query":
		return rankQuery
	case "Mutation":
		return rankMutation
	case "Subscription":
		return rankSubscription
	default:
		return rankObject
	}
}

// Fields returns a list of the FieldDefs.
func (t *Object) Fields() []*FieldDef {
	return t.fields.list
}

// String representation of the type.
func (t *Object) String() string {
	return t.SDL()
}

// SDL returns an SDL representation of the type.
func (t *Object) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Object) Write(w io.Writer, desc bool) (err error) {
	return t.write(w, "type ", desc)
}

func (t *Object) write(w io.Writer, key string, desc bool) (err error) {
	if err = t.writeHeader(w, key, desc, t.Interfaces...); err == nil {
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
func (t *Object) Extend(x Type) error {
	if ox, ok := x.(*Object); ok { // Already checked so no need to report an error again.
		for _, fd := range ox.fields.list {
			if err := t.fields.add(fd); err != nil {
				return fmt.Errorf("%w: on %s", err, t.N)
			}
		}
		for _, i := range ox.Interfaces {
			for _, exist := range t.Interfaces {
				if i.Name() == exist.Name() {
					return fmt.Errorf("%w: interface %s already exists on %s", ErrDuplicate, i.Name(), t.N)
				}
			}
			t.Interfaces = append(t.Interfaces, i)
		}
	}
	return t.Base.Extend(x)
}

// GetField returns the field matching the name or nil if not found.
func (t *Object) GetField(name string) *FieldDef {
	return t.fields.get(name)
}

// Validate a type.
func (t *Object) Validate(root *Root) (errs []error) {
	for _, it := range t.Interfaces {
		if i, ok := it.(*Interface); ok {
			errs = append(errs, t.validateInterface(i)...)
		} else {
			errs = append(errs, fmt.Errorf("%w, %s is not an interface for %s at %d:%d",
				ErrValidation, it.Name(), t.Name(), t.line, t.col))
		}
	}
	return append(errs, t.validateFieldDefs(t.Name(), &t.fields)...)
}

func (t *Object) validateInterface(i *Interface) (errs []error) {
	for name, fi := range i.fields.dict {
		fo := t.fields.get(name)
		if fo == nil {
			errs = append(errs, fmt.Errorf("%w, %s is missing field %s from interface %s at %d:%d",
				ErrValidation, t.Name(), name, i.Name(), t.line, t.col))
			continue
		}
		errs = append(errs, t.validateField(fo, fi, i.Name())...)
	}
	return
}

// validateField checks the object FieldDef against the interface FieldDef.
func (t *Object) validateField(fo, fi *FieldDef, iName string) (errs []error) {
	if !t.isSubType(fi.Type, fo.Type) {
		errs = append(errs, fmt.Errorf("%w, interface %s not satisfied, field %s return type %s is not a sub-type of %s at %d:%d",
			ErrValidation, iName, fi.Name(), fo.Type.Name(), fi.Type.Name(), fo.line, fo.col))
	}
	for _, ai := range fi.args.list {
		if ao := fo.getArg(ai.Name()); ao == nil {
			errs = append(errs, fmt.Errorf("%w, interface %s not satisfied, argument %s to %s missing at %d:%d",
				ErrValidation, iName, fi.Name(), ai.Name(), fo.line, fo.col))
		}
	}
	for _, ao := range fo.args.list {
		ai := fi.getArg(ao.Name())
		if ai == nil {
			if _, ok := ao.Type.(*NonNull); ok {
				errs = append(errs, fmt.Errorf("%w, interface %s not satisfied, additional argument %s to field %s must be optional at %d:%d",
					ErrValidation, iName, ao.Name(), fo.Name(), ao.line, ao.col))
			}
		} else if !typeEqual(ai.Type, ao.Type) {
			errs = append(errs, fmt.Errorf("%w, interface %s not satisfied, argument return type for %s does not match at %d:%d",
				ErrValidation, iName, ai.Name(), ao.line, ao.col))
		}
	}
	return
}

func (t *Object) isSubType(target, sub Type) bool {
	if typeEqual(target, sub) {
		return true
	}
	switch tt := target.(type) {
	case *Union:
		for _, m := range tt.Members {
			if typeEqual(m, sub) {
				return true
			}
		}
	case *Interface:
		if ot, _ := sub.(*Object); ot != nil {
			for _, i := range ot.Interfaces {
				if typeEqual(i, target) {
					return true
				}
			}
		}
	case *List:
		if list, _ := sub.(*List); list != nil {
			return t.isSubType(tt.Base, list.Base)
		}
	case *NonNull:
		if nn, _ := sub.(*NonNull); nn != nil {
			return t.isSubType(tt.Base, nn.Base)
		}
	}
	return false
}

// AddField is used to add fields to an object.
func (t *Object) AddField(fd *FieldDef) error {
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
func (t *Object) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case kindStr:
		result = string(Locate(t))
	case nameStr:
		if 0 < len(t.N) {
			result = t.N
		} else {
			result = schemaStr
		}
	case descriptionStr:
		result = t.Desc
	case fieldsStr:
		if t.getBoolArg(args, includeDeprecatedStr) {
			result = &t.fields
		} else {
			list := fieldList{dict: map[string]*FieldDef{}}
			for _, f := range t.fields.list {
				if !f.isDeprecated() {
					_ = list.add(f)
				}
			}
			result = &list
		}
	case interfacesStr:
		result = t.Interfaces
	case possibleTypesStr, enumValuesStr, inputFieldsStr, ofTypeStr:
		// nil result
	}
	return
}

func (t *Object) metaCheck(rt reflect.Type) error {
	if t.meta == nil {
		bt := rt
		for bt.Kind() == reflect.Ptr {
			bt = bt.Elem()
		}
		du := t.GetDirective("go")
		if du != nil && du.Args != nil {
			if a := du.Args["type"]; a != nil {
				// Check the most specific name first which is full path and
				// type name. Next short package name and finally just then
				// name.
				s, _ := a.Value.(string)
				if s == bt.PkgPath()+"."+bt.Name() ||
					s == bt.String() ||
					s == bt.Name() {
					t.meta = rt
				}
			}
		} else if t.N == bt.Name() { // If no @go directive then try using the GraphQL type name.
			t.meta = rt
		}
	}
	if t.meta == nil {
		return resError(t.line, t.col, "failed to determine union member %s implementation type. Use @go directive", t.N)
	}
	return nil
}
