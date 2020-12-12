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
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Input is a GraphQL InputObject type.
type Input struct {
	Base

	// Fields in the input object.
	fields inputFieldList

	meta reflect.Type
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
				return nil, fmt.Errorf("%s is not a field in %s", k, t.Name())
			}
		}
		var rv reflect.Value
		rt := t.meta
		if rt != nil {
			if rt.Kind() == reflect.Ptr {
				rt = rt.Elem()
			}
			if rt.Kind() == reflect.Struct {
				rv = reflect.New(rt)
			}
		}
		for k, f := range t.fields.dict {
			ov := tv[k]
			if ov == nil {
				if f.Default != nil { // if not set then add the default value if not nil
					if rt != nil {
						if err := t.reflectSetKey(rv, k, f.Default); err != nil {
							return nil, inErr(err, k)
						}
					} else {
						tv[k] = f.Default
					}
				} else if _, ok := f.Type.(*NonNull); ok {
					return nil, fmt.Errorf("%s is required but missing", k)
				}
			} else if co, _ := f.Type.(InCoercer); co != nil {
				if cv, err := co.CoerceIn(ov); err == nil {
					if rt != nil {
						if err = t.reflectSetKey(rv, k, cv); err != nil {
							return nil, inErr(err, k)
						}
					} else {
						tv[k] = cv
					}
				} else {
					return nil, inErr(err, k)
				}
			} else {
				return nil, newCoerceErr(ov, f.Type.Name())
			}
		}
		if rt != nil {
			v = rv.Interface()
		}
	default:
		if t.meta == nil || t.meta != reflect.TypeOf(v) {
			return nil, newCoerceErr(v, t.Name())
		}
	}
	return v, nil
}

func inErr(err error, k string) error {
	var gerr *Error
	if errors.As(err, &gerr) {
		gerr.in(k)
	} else {
		gerr = &Error{Base: err, Path: []interface{}{k}}
		err = gerr
	}
	return err
}

func (t *Input) reflectSetKey(rv reflect.Value, key string, v interface{}) (err error) {
	rv = rv.Elem()
	rv = rv.FieldByNameFunc(func(k string) bool { return strings.EqualFold(k, key) })

	return t.reflectSet(rv, v)
}

func (t *Input) reflectSet(rv reflect.Value, v interface{}) (err error) {
	if rv.CanSet() {
		vv := reflect.ValueOf(v)
		vt := vv.Type()
		if vt.AssignableTo(rv.Type()) {
			rv.Set(vv)
			return
		}
		// Try type conversions of int and float types.
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch vt.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				rv.SetInt(vv.Int())
				return
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			switch vt.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i := vv.Int(); 0 <= i {
					rv.SetUint(uint64(i))
					return
				}
			}
		case reflect.Float32, reflect.Float64:
			switch vt.Kind() {
			case reflect.Float32, reflect.Float64:
				rv.SetFloat(vv.Float())
				return
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				rv.SetFloat(float64(vv.Int()))
				return
			}
		case reflect.Slice:
			if vt.Kind() == reflect.Slice {
				i := vv.Len()
				av := reflect.MakeSlice(rv.Type(), i, i)
				for i--; 0 <= i; i-- {
					if err = t.reflectSet(av.Index(i), vv.Index(i).Interface()); err != nil {
						return
					}
				}
				rv.Set(av)
				return
			}
		}
	}
	return fmt.Errorf("can not coerce a %T into a %s", v, rv.Kind())
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
