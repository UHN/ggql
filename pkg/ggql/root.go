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
	"strings"
	"sync"
)

// Relaxed if true relaxes coercion rules so that JSON types can be converted
// to GraphQL types. For example a string can be coerced into an enum or a
// JSON object can be converted into a GraphQL input type. Note that turning
// this on goes against the spec since strings should not be coerced into
// enums.
var Relaxed = false

// Root the root of a GraphQL schema.
type Root struct {
	types         *typeList
	dirs          *typeList
	obj           interface{}
	queryObj      interface{}
	schema        *Schema
	uuSchemaType  *uuSchema
	AnyResolver   AnyResolver
	subscriptions []*Subscription
	subLock       sync.Mutex
	excludeTime   bool
	excludeInt64  bool
}

// NewRoot creates a new GraphQL schema root with a root resolver object. The
// optional exclude arguments are the built in type to exclude. The built in
// type that can be excluded are "Time" and "Int64". Excluding a type allows
// it to be implemented separately without conflict.
func NewRoot(obj interface{}, exclude ...string) *Root {
	root := Root{obj: obj}
	for _, x := range exclude {
		switch x {
		case "Time":
			root.excludeTime = true
		case "Int64":
			root.excludeInt64 = true
		}
	}
	root.init()

	return &root
}

// AddTypes adds types to the schema.
func (root *Root) AddTypes(types ...Type) (err error) {
	root.init()

	// Save the original map and create a duplicate. If there is an error,
	// revert to the original version.
	origTypes := root.types
	origDirs := root.dirs
	root.types = origTypes.dup()
	root.dirs = origDirs.dup()

	err = root.addTypes(types...)
	if err == nil {
		err = root.validate()
	}
	if err != nil {
		root.types = origTypes
		root.dirs = origDirs
	}
	return
}

// Types returns a list of all the types in the root.
func (root *Root) Types() []Type {
	root.init()

	return root.types.list
}

// RegisterType associates a Go type with a GraphQL type. This is only needed
// for Object or Schema types used in GraphQL unions or more accurately the
// condition of a fragment when used with a union. It should also be used when
// a field or field arguments can not be mapped automatically.
func (root *Root) RegisterType(sample interface{}, gqlType string) error {
	obj, input, err := root.getObjType(gqlType)
	if err != nil {
		return err
	}
	if obj != nil {
		return root.regType(sample, obj)
	}
	return root.regInput(sample, input)
}

func (root *Root) getObjType(gqlType string) (obj *Object, input *Input, err error) {
	root.init()
	if 0 < len(gqlType) && gqlType != schemaStr {
		t := root.types.get(gqlType)
		switch tt := t.(type) {
		case *Object:
			obj = tt
		case *Input:
			input = tt
		default:
			err = fmt.Errorf("%s %w or not an Object", gqlType, ErrNotFound)
		}
	} else {
		root.assureSchema()
		obj = &root.schema.Object
	}
	return
}

func (root *Root) regType(sample interface{}, obj *Object) error {
	meta := reflect.TypeOf(sample)
	if obj.meta != nil && obj.meta != meta {
		return fmt.Errorf("%w: %s is already registered as a %s", ErrDuplicate, obj.N, obj.meta.String())
	}
	obj.meta = meta

	return nil
}

func (root *Root) regInput(sample interface{}, input *Input) error {
	meta := reflect.TypeOf(sample)
	if input.meta != nil && input.meta != meta {
		return fmt.Errorf("%w: %s is already registered as a %s", ErrDuplicate, input.N, input.meta.String())
	}
	input.meta = meta

	return nil
}

func (root *Root) getReflectType(meta reflect.Type) (obj Type) {
	for _, t := range root.types.list {
		if o, _ := t.(*Object); o != nil && o.meta == meta {
			obj = o
			break
		}
	}
	return
}

// RegisterField registers a field for a type. This function should be used to
// register fields or functions that don't follow the simple conversion of
// capitalizing the GraphQL field name. It can also be used to change the
// order of the arguments to the field.
func (root *Root) RegisterField(gqlType, gqlField, goField string, args ...string) (err error) {
	obj, _, err := root.getObjType(gqlType)
	if err != nil {
		return err
	}
	if obj == nil || obj.meta == nil {
		return fmt.Errorf("%w: %s has not been registered as a type", ErrMeta, obj.N)
	}
	fd := obj.fields.get(gqlField)
	if fd == nil {
		return fmt.Errorf("%w: %s is not a field of %s", ErrMeta, gqlField, obj.N)
	}
	return root.regField(obj, fd, goField, args...)
}

func (root *Root) regField(obj *Object, fd *FieldDef, goField string, args ...string) (err error) {
	meta := obj.meta
	if meta.Kind() == reflect.Ptr {
		meta = meta.Elem()
	}
	if meta.Kind() == reflect.Struct {
		if field, ok := meta.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, goField)
		}); ok {
			fd.goField = field.Name
			if 0 < len(args) {
				err = fmt.Errorf("%w: field %s on %s does not have argument", ErrMeta, goField, meta)
			}
			return
		}
	}
	for i := obj.meta.NumMethod() - 1; 0 <= i; i-- {
		m := obj.meta.Method(i)
		if strings.EqualFold(m.Name, goField) {
			fd.method = &m.Func
			break
		}
	}
	if fd.method != nil {
		if 0 < len(args) {
			if fd.args.Len() != len(args) {
				return fmt.Errorf("%w: not enough arguments for field %s of %s", ErrMeta, goField, obj.meta)
			}
			newArgs := argList{}
			for _, arg := range args {
				if a := fd.args.get(arg); a != nil {
					_ = newArgs.add(a)
				} else {
					err = fmt.Errorf("%w: %s is not an argument on field %s of %s", ErrMeta, arg, goField, obj.meta)
					break
				}
			}
			fd.args = newArgs
		}
		return
	}
	return fmt.Errorf("%w: %s is not a field of %s", ErrMeta, goField, obj.meta)
}

func (root *Root) addTypes(types ...Type) error {
	for _, t := range types {
		name := t.Name()
		if dir, _ := t.(*Directive); dir != nil {
			if root.dirs.get(name) != nil {
				return fmt.Errorf("%w: %s is already in the schema", ErrDuplicate, name)
			}
			root.dirs.add(t)
		} else {
			if root.types.get(name) != nil {
				// If a scalar, do not replace and do not complain.
				if t.Rank() == rankScalar {
					continue
				}
				return fmt.Errorf("%w: %s is already in the schema", ErrDuplicate, name)
			}
			switch t.(type) {
			case *List, *NonNull, *Ref:
				return fmt.Errorf("%w: %s, a %T can not be added", ErrTypeMismatch, name, t)
			default:
				root.types.add(t)
			}
		}
	}
	return root.ReplaceRefs()
}

func (root *Root) addExtends(extends ...*Extend) (err error) {
	for _, x := range extends {
		if err = root.replaceTypeRefs(x.Adds); err != nil {
			return
		}
		var cur Type
		if 0 < len(x.Adds.Name()) {
			cur = root.types.get(x.Adds.Name())
			if cur == nil {
				cur = root.dirs.get(x.Adds.Name())
			}
		} else if schema, _ := x.Adds.(*Schema); schema != nil {
			cur = root.schema
		}
		if cur == nil {
			return fmt.Errorf("%s can not be extended because it was %w", x.Adds.Name(), ErrNotFound)
		}
		if reflect.TypeOf(x.Adds) != reflect.TypeOf(cur) {
			return fmt.Errorf("%w: %s, a %T can not extend a %T", ErrTypeMismatch, x.Adds.Name(), x.Adds, cur)
		}
		if err = cur.Extend(x.Adds); err != nil {
			return
		}
	}
	return nil
}

// GetType returns the type that matches the provided name or nil if none
// match.
func (root *Root) GetType(name string) Type {
	root.init()
	t := root.types.get(name)
	if t == nil {
		t = root.dirs.get(name)
	}
	return t
}

// ParseString parses an SDL string into a Doc.
func (root *Root) ParseString(s string) error {
	return root.ParseReader(strings.NewReader(s))
}

// Parse parses an SDL byte array into a Doc.
func (root *Root) Parse(b []byte) error {
	return root.ParseReader(bytes.NewReader(b))
}

// ParseReader parses an SDL reader into a Doc.
func (root *Root) ParseReader(r io.Reader) error {
	root.init()

	// Save the original map and create a duplicate. If there is an error,
	// revert to the original version.
	origTypes := root.types
	origDirs := root.dirs
	root.types = origTypes.dup()
	root.dirs = origDirs.dup()

	types, extends, err := parseSDL(root, r)
	if err == nil {
		err = root.addTypes(types...)
	}
	if err == nil {
		err = root.addExtends(extends...)
	}
	if err == nil {
		err = root.validate()
	}
	if err != nil {
		root.types = origTypes
		root.dirs = origDirs
	}
	return err
}

// ParseExecutableString parses an SDL string into a Doc.
func (root *Root) ParseExecutableString(s string) (*Executable, error) {
	return root.ParseExecutableReader(strings.NewReader(s))
}

// ParseExecutable parses an SDL byte array into a Doc.
func (root *Root) ParseExecutable(b []byte) (*Executable, error) {
	return root.ParseExecutableReader(bytes.NewReader(b))
}

// ParseExecutableReader parses an SDL reader into a Doc.
func (root *Root) ParseExecutableReader(r io.Reader) (*Executable, error) {
	root.init() // Schema should have been loaded already but just to avoid issue check again.
	exe, err := parseExe(root, r)
	if err == nil {
		if errs := exe.Validate(root); 0 < len(errs) {
			err = Errors(errs)
		}
	}
	return exe, err
}

// SDL returns a SDL representation of the instance.
func (root *Root) SDL(full bool, desc ...bool) string {
	var b strings.Builder

	for _, t := range root.types.list {
		if full || !t.Core() {
			b.Write([]byte{'\n'})
			b.WriteString(t.SDL(desc...))
		}
	}
	for _, t := range root.dirs.list {
		if full || !t.Core() {
			b.Write([]byte{'\n'})
			b.WriteString(t.SDL(desc...))
		}
	}
	return b.String()
}

func (root *Root) validate() error {
	var errs []error

	for _, t := range root.types.list {
		errs = append(errs, root.validateTypeName("type", t)...)
		errs = append(errs, root.validateDirUses(t)...)
		errs = append(errs, t.Validate(root)...)
	}
	for _, t := range root.dirs.list {
		errs = append(errs, root.validateTypeName("directive", t)...)
		errs = append(errs, root.validateDirUses(t)...)
		errs = append(errs, t.Validate(root)...)
	}
	if 0 < len(errs) {
		return Errors(errs)
	}
	return nil
}

// Checks that name satisfies the requirements specified in
// https://spec.graphql.org/June2018/#Name.
// If allowReserved is false an error is returned when name has a '__' prefix.
func validateName(allowReserved bool, typeName, name string, line, col int) (errs []error) {
	if len(name) == 0 {
		errs = append(errs, fmt.Errorf("%w, a %s name can not be blank at %d:%d",
			ErrValidation, typeName, line, col))
	} else {
		for _, b := range name {
			if charMap[b] != tokenChar {
				errs = append(errs, fmt.Errorf("%w, %s is not a valid %s name at %d:%d",
					ErrValidation, name, typeName, line, col))
				break
			}
		}
		if !allowReserved && strings.HasPrefix(name, "__") {
			errs = append(errs, fmt.Errorf("%w, %s is not a valid %s name, it begins with '__' at %d:%d",
				ErrValidation, name, typeName, line, col))
		}
		if '0' <= name[0] && name[0] <= '9' {
			errs = append(errs, fmt.Errorf("%w, %s is not a valid %s name, it begins with a number at %d:%d",
				ErrValidation, name, typeName, line, col))
		}
	}
	return
}

func (root *Root) validateTypeName(typeName string, t Type) (errs []error) {
	name := t.Name()
	if st, _ := t.(*Schema); st != nil {
		if 0 < len(name) {
			errs = append(errs, fmt.Errorf("%w, a schema can not have a name at %d:%d",
				ErrValidation, st.line, st.col))
		}
	} else {
		errs = append(errs, validateName(t.Core(), typeName, name, t.Line(), t.Column())...)
	}
	return
}

func (root *Root) validateDirUses(t Type) (errs []error) {
	for _, du := range t.Directives() {
		errs = append(errs, root.validateDirUse(t.Name(), Locate(t), du)...)
	}
	return
}

func (root *Root) validateDirUse(where string, loc Location, du *DirectiveUse) (errs []error) {
	d, _ := du.Directive.(*Directive)
	if d == nil {
		errs = append(errs, fmt.Errorf("%w, invalid directive at %s a %s at %d:%d",
			ErrValidation, where, loc, du.line, du.col))
		return
	}
	found := false
	for _, on := range d.On {
		if loc == on {
			found = true
			break
		}
	}
	if !found {
		return append(errs, fmt.Errorf("%w, directive @%s can not be applied to %s, a %s at %d:%d",
			ErrValidation, d.Name(), where, loc, du.line, du.col))
	}
	var a *Arg
	for _, av := range du.Args {
		if a = d.findArg(av.Arg); a == nil {
			errs = append(errs, fmt.Errorf("%w, directive argument %s for directive %s on %s not found at %d:%d",
				ErrValidation, av.Arg, du.Directive.Name(), where, av.line, av.col))
			continue
		}
		// The Directive validation should check all argument type to make
		// sure they are Coercers so no need to double report the issue
		// here. A Var is also allowed.
		if _, ok := av.Value.(Var); !ok {
			if co, _ := a.Type.(InCoercer); co != nil {
				if v, err := co.CoerceIn(av.Value); err != nil {
					errs = append(errs, fmt.Errorf("%w at %d:%d", err, av.line, av.col))
				} else if v != av.Value {
					// Might as well replace the coerced value since it is really
					// what is needed.
					av.Value = v
				}
			}
		}
	}
	return
}

func (root *Root) init() {
	if root.types != nil {
		return
	}
	root.types = newTypeList()
	root.dirs = newTypeList()
	addType := func(t Type) Type {
		root.types.add(t)
		return t
	}
	strType := addType(newStringScalar())
	addType(newIntScalar())
	addType(newFloatScalar())
	addType(newFloat64Scalar())
	addType(newBooleanScalar())
	addType(newIDScalar())
	if !root.excludeTime {
		addType(newTimeScalar())
	}
	if !root.excludeInt64 {
		addType(newInt64Scalar())
	}
	typeKind := addType(root.newTypeKind())
	addType(root.newDirectiveLocation())

	uuType := addType(root.newUuType(typeKind, strType))
	inputValue := addType(root.newUuInputValue(uuType, strType))
	uuDir := addType(root.newUuDirective(inputValue, strType))
	addType(root.newUuSchema(uuType, uuDir))
	addType(root.newUuField(uuType, inputValue, strType))
	addType(root.newUuEnumValue(strType))

	root.dirs.add(root.newSkipDirective())
	root.dirs.add(root.newIncludeDirective())
	root.dirs.add(root.newDeprecatedDirective())
	root.dirs.add(root.newGoDirective())

	// Okay to not check the error here as unit tests cover the case where an
	// error could occur.
	_ = root.ReplaceRefs()
}

// ReplaceRefs replaces reference in the schema with the referenced types.
func (root *Root) ReplaceRefs() (err error) {
	for _, t := range root.types.list {
		if err = root.replaceTypeRefs(t); err != nil {
			break
		}
	}
	if err == nil {
		for _, t := range root.dirs.list {
			if err = root.replaceTypeRefs(t); err != nil {
				break
			}
		}
	}
	return
}

func (root *Root) replaceTypeRefs(t Type) (err error) {
	if err = root.replaceDirRefs(t.Directives()); err != nil {
		return
	}

	switch tt := t.(type) {
	case *Schema:
		if err = root.replaceFieldRefs(&tt.fields); err != nil {
			return
		}
	case *Object:
		if err = root.replaceFieldRefs(&tt.fields); err != nil {
			return
		}
		if err = root.replaceInterfaceRefs(tt.Interfaces); err != nil {
			return
		}
	case *Interface:
		if err = root.replaceFieldRefs(&tt.fields); err != nil {
			return
		}
	case *Input:
		if err = root.replaceInputFieldRefs(&tt.fields); err != nil {
			return
		}
	case *Union:
		for i, m := range tt.Members {
			if r, _ := m.(*Ref); r != nil {
				if nt := root.types.get(m.Name()); nt != nil {
					tt.Members[i] = nt
				} else {
					return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, m.Name(), tt.line, tt.col)
				}
			}
		}
	case *Enum:
		for _, ev := range tt.values.list {
			if err = root.replaceDirRefs(ev.Directives); err != nil {
				return
			}
		}
	case *Directive:
		if err = root.replaceArgRefs(&tt.args); err != nil {
			return
		}
	}
	return
}

func (root *Root) replaceInterfaceRefs(infs []Type) error {
	for i, t := range infs {
		if tt, _ := t.(*Ref); tt != nil {
			if infs[i] = root.types.get(tt.Name()); infs[i] == nil {
				return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, tt.Name(), t.Line(), t.Column())
			}
		}
	}
	return nil
}

func (root *Root) replaceListRefs(list *List) error {
	switch bt := list.Base.(type) {
	case *Ref:
		if list.Base = root.types.get(bt.Name()); list.Base == nil {
			return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, bt.Name(), list.line, list.col)
		}
	case *List:
		return root.replaceListRefs(bt)
	case *NonNull:
		return root.replaceNonNullRefs(bt)
	}
	return nil
}

func (root *Root) replaceNonNullRefs(nn *NonNull) error {
	switch bt := nn.Base.(type) {
	case *Ref:
		if nn.Base = root.types.get(bt.Name()); nn.Base == nil {
			return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, bt.Name(), nn.line, nn.col)
		}
	case *List:
		return root.replaceListRefs(bt)
	case *NonNull:
		return fmt.Errorf("%w error, '%s' non-null in non-null not allowed at %d:%d",
			ErrValidation, bt.Name(), nn.line, nn.col)
	}
	return nil
}

func (root *Root) replaceFieldRefs(fields *fieldList) (err error) {
	for _, f := range fields.list {
		t := f.Type
		switch tt := f.Type.(type) {
		case *Ref:
			if f.Type = root.types.get(t.Name()); f.Type == nil {
				return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, t.Name(), f.line, f.col)
			}
		case *List:
			if err = root.replaceListRefs(tt); err != nil {
				return
			}
		case *NonNull:
			if err = root.replaceNonNullRefs(tt); err != nil {
				return
			}
		}
		if err = root.replaceArgRefs(&f.args); err != nil {
			return
		}
		if err = root.replaceDirRefs(f.Dirs); err != nil {
			return
		}
	}
	return
}

func (root *Root) replaceInputFieldRefs(args *inputFieldList) (err error) {
	for _, a := range args.list {
		t := a.Type
		switch tt := a.Type.(type) {
		case *Ref:
			if a.Type = root.types.get(t.Name()); a.Type == nil {
				return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, t.Name(), a.line, a.col)
			}
		case *List:
			if err = root.replaceListRefs(tt); err != nil {
				return
			}
		case *NonNull:
			if err = root.replaceNonNullRefs(tt); err != nil {
				return
			}
		}
	}
	return
}

func (root *Root) replaceArgRefs(args *argList) (err error) {
	for _, a := range args.list {
		t := a.Type
		switch tt := a.Type.(type) {
		case *Ref:
			if a.Type = root.types.get(t.Name()); a.Type == nil {
				return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, t.Name(), a.line, a.col)
			}
		case *List:
			if err = root.replaceListRefs(tt); err != nil {
				return
			}
		case *NonNull:
			if err = root.replaceNonNullRefs(tt); err != nil {
				return
			}
		}
		if err = root.replaceDirRefs(a.Dirs); err != nil {
			return
		}
	}
	return
}

func (root *Root) replaceDirRefs(dirs []*DirectiveUse) (err error) {
	for _, du := range dirs {
		t := du.Directive
		// Just check for *Ref. Any others will be rejected later in validation.
		if tt, _ := t.(*Ref); tt != nil {
			if du.Directive = root.dirs.get(t.Name()); du.Directive == nil {
				return fmt.Errorf("%w error, '%s' not defined at %d:%d", ErrValidation, t.Name(), du.line, du.col)
			}
		}
	}
	return
}

// type __Type {
//   kind: __TypeKind!
//   name: String
//   description: String
//   fields(includeDeprecated: Boolean = false): [__Field!]
//   interfaces: [__Type!]
//   possibleTypes: [__Type!]
//   enumValues(includeDeprecated: Boolean = false): [__EnumValue!]
//   inputFields: [__InputValue!]
//   ofType: __Type
// }.
func (root *Root) newUuType(typeKind, strType Type) Type {
	t := Object{
		Base: Base{
			N:    "__Type",
			core: true,
		},
		fields: fieldList{dict: map[string]*FieldDef{}},
	}
	_ = t.fields.add(&FieldDef{Base: Base{N: kindStr}, Type: &NonNull{Base: typeKind}})
	_ = t.fields.add(&FieldDef{Base: Base{N: nameStr}, Type: strType})
	_ = t.fields.add(&FieldDef{Base: Base{N: descriptionStr}, Type: strType})

	fd := &FieldDef{Base: Base{N: fieldsStr}, Type: &List{Base: &NonNull{Base: &Ref{Base: Base{N: "__Field"}}}}}
	_ = fd.args.add(&Arg{Base: Base{N: includeDeprecatedStr}, Type: root.types.get("Boolean"), Default: false})

	_ = t.fields.add(fd)

	fd = &FieldDef{Base: Base{N: enumValuesStr}, Type: &List{Base: &NonNull{Base: &Ref{Base: Base{N: "__EnumValue"}}}}}
	_ = fd.args.add(&Arg{Base: Base{N: includeDeprecatedStr}, Type: root.types.get("Boolean"), Default: false})

	_ = t.fields.add(fd)
	_ = t.fields.add(&FieldDef{Base: Base{N: inputFieldsStr},
		Type: &List{Base: &NonNull{Base: &Ref{Base: Base{N: "__InputValue"}}}},
	})

	typeList := &NonNull{Base: &List{Base: &NonNull{Base: &t}}}
	_ = t.fields.add(&FieldDef{Base: Base{N: interfacesStr}, Type: typeList})
	_ = t.fields.add(&FieldDef{Base: Base{N: possibleTypesStr}, Type: typeList})
	_ = t.fields.add(&FieldDef{Base: Base{N: ofTypeStr}, Type: &t})

	return &t
}

// type __Schema {
//   types: [__Type!]!
//   queryType: __Type!
//   mutationType: __Type
//   subscriptionType: __Type
//   directives: [__Directive!]!
// }.
func (root *Root) newUuSchema(uuType, uuDir Type) Type {
	root.uuSchemaType = &uuSchema{
		root: root,
		Object: Object{
			Base: Base{
				N:    "__Schema",
				core: true,
			},
			fields: fieldList{dict: map[string]*FieldDef{}},
		},
	}
	_ = root.uuSchemaType.fields.add(&FieldDef{
		Base: Base{N: "types"},
		Type: &NonNull{Base: &List{Base: &NonNull{Base: uuType}}},
	})
	_ = root.uuSchemaType.fields.add(&FieldDef{Base: Base{N: "queryType"}, Type: &NonNull{Base: uuType}})
	_ = root.uuSchemaType.fields.add(&FieldDef{Base: Base{N: "mutationType"}, Type: uuType})
	_ = root.uuSchemaType.fields.add(&FieldDef{Base: Base{N: "subscriptionType"}, Type: uuType})
	_ = root.uuSchemaType.fields.add(&FieldDef{Base: Base{N: "directives"},
		Type: &NonNull{Base: &List{Base: &NonNull{Base: uuDir}}},
	})
	return root.uuSchemaType
}

// type __InputValue {
//   name: String!
//   description: String
//   type: __Type!
//   defaultValue: String
// }.
func (root *Root) newUuInputValue(uuType, strType Type) Type {
	t := Object{
		Base: Base{
			N:    "__InputValue",
			core: true,
		},
		fields: fieldList{dict: map[string]*FieldDef{}},
	}
	_ = t.fields.add(&FieldDef{Base: Base{N: nameStr}, Type: &NonNull{Base: strType}})
	_ = t.fields.add(&FieldDef{Base: Base{N: descriptionStr}, Type: strType})
	_ = t.fields.add(&FieldDef{Base: Base{N: typeStr}, Type: &NonNull{Base: uuType}})
	_ = t.fields.add(&FieldDef{Base: Base{N: defaultValueStr}, Type: strType})

	return &t
}

// type __Field {
//   name: String!
//   description: String
//   args: [__InputValue!]!
//   type: __Type!
//   isDeprecated: Boolean!
//   deprecationReason: String
// }.
func (root *Root) newUuField(uuType, inputValue, strType Type) Type {
	t := Object{
		Base: Base{
			N:    "__Field",
			core: true,
		},
		fields: fieldList{dict: map[string]*FieldDef{}},
	}
	_ = t.fields.add(&FieldDef{Base: Base{N: nameStr}, Type: &NonNull{Base: strType}})
	_ = t.fields.add(&FieldDef{Base: Base{N: descriptionStr}, Type: strType})
	_ = t.fields.add(&FieldDef{Base: Base{N: "args"},
		Type: &NonNull{Base: &List{Base: &NonNull{Base: inputValue}}},
	})
	_ = t.fields.add(&FieldDef{Base: Base{N: typeStr}, Type: &NonNull{Base: uuType}})
	_ = t.fields.add(&FieldDef{Base: Base{N: isDeprecatedStr}, Type: &NonNull{Base: root.types.get(booleanStr)}})
	_ = t.fields.add(&FieldDef{Base: Base{N: deprecationReasonStr}, Type: strType})

	return &t
}

// type __EnumValue {
//   name: String!
//   description: String
//   isDeprecated: Boolean!
//   deprecationReason: String
// }.
func (root *Root) newUuEnumValue(strType Type) Type {
	t := Object{
		Base: Base{
			N:    "__EnumValue",
			core: true,
		},
		fields: fieldList{dict: map[string]*FieldDef{}},
	}
	_ = t.fields.add(&FieldDef{Base: Base{N: nameStr}, Type: &NonNull{Base: strType}})
	_ = t.fields.add(&FieldDef{Base: Base{N: descriptionStr}, Type: strType})
	_ = t.fields.add(&FieldDef{Base: Base{N: isDeprecatedStr}, Type: &NonNull{Base: root.types.get(booleanStr)}})
	_ = t.fields.add(&FieldDef{Base: Base{N: deprecationReasonStr}, Type: strType})

	return &t
}

// type __Directive {
//   name: String!
//   description: String
//   locations: [__DirectiveLocation!]!
//   args: [__InputValue!]!
// }.
func (root *Root) newUuDirective(inputValue, strType Type) Type {
	t := Object{
		Base: Base{
			N:    "__Directive",
			core: true,
		},
		fields: fieldList{dict: map[string]*FieldDef{}},
	}
	_ = t.fields.add(&FieldDef{Base: Base{N: nameStr}, Type: &NonNull{Base: strType}})
	_ = t.fields.add(&FieldDef{Base: Base{N: descriptionStr}, Type: strType})
	_ = t.fields.add(&FieldDef{Base: Base{N: locationsStr},
		Type: &NonNull{Base: &List{Base: &NonNull{Base: root.types.get("__DirectiveLocation")}}},
	})
	_ = t.fields.add(&FieldDef{Base: Base{N: argsStr},
		Type: &NonNull{Base: &List{Base: &NonNull{Base: inputValue}}},
	})
	return &t
}

// directive @skip(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT.
func (root *Root) newSkipDirective() Type {
	t := Directive{
		Base: Base{
			N:    "skip",
			core: true,
		},
		On: []Location{LocField, LocFragmentSpread, LocInlineFragment},
	}
	_ = t.args.add(&Arg{Base: Base{N: "if"}, Type: &NonNull{Base: root.types.get(booleanStr)}})

	return &t
}

// directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT.
func (root *Root) newIncludeDirective() Type {
	t := Directive{
		Base: Base{
			N:    "include",
			core: true,
		},
		On: []Location{LocField, LocFragmentSpread, LocInlineFragment},
	}
	_ = t.args.add(&Arg{Base: Base{N: "if"}, Type: &NonNull{Base: root.types.get(booleanStr)}})

	return &t
}

// directive @deprecated(reason: String = "No longer supported") on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT.
func (root *Root) newDeprecatedDirective() Type {
	t := Directive{
		Base: Base{
			N:    deprecatedStr,
			core: true,
		},
		On: []Location{LocFieldDefinition, LocEnumValue},
	}
	// The Default value double quotes the value. Seems broken but GraphiQL
	// expects it. Odd that it deviates from any other string value though.
	_ = t.args.add(&Arg{Base: Base{N: reasonStr}, Type: root.types.get(stringStr), Default: "\"No longer supported\""})

	return &t
}

// directive @go(type: String!) on SCHEMA | QUERY | MUTATION | SUBSCRIPTION | OBJECT | FIELD | INPUT_OBJECT.
func (root *Root) newGoDirective() Type {
	t := Directive{
		Base: Base{
			N:    "go",
			core: true,
		},
		On: []Location{LocSchema, LocQuery, LocMutation, LocSubscription, LocObject, LocFieldDefinition},
	}
	_ = t.args.add(&Arg{Base: Base{N: typeStr}, Type: &NonNull{Base: root.types.get(stringStr)}})

	return &t
}

func (root *Root) newTypeKind() Type {
	t := Enum{
		Base: Base{
			N:    "__TypeKind",
			core: true,
		},
	}
	_ = t.values.add(&EnumValue{Value: "SCALAR"})
	_ = t.values.add(&EnumValue{Value: "OBJECT"})
	_ = t.values.add(&EnumValue{Value: "INTERFACE"})
	_ = t.values.add(&EnumValue{Value: "UNION"})
	_ = t.values.add(&EnumValue{Value: "ENUM"})
	_ = t.values.add(&EnumValue{Value: "INPUT_OBJECT"})
	_ = t.values.add(&EnumValue{Value: "LIST"})
	_ = t.values.add(&EnumValue{Value: "NON_NULL"})

	return &t
}

func (root *Root) newDirectiveLocation() Type {
	t := Enum{
		Base: Base{
			N:    "__DirectiveLocation",
			core: true,
		},
	}
	_ = t.values.add(&EnumValue{Value: "QUERY"})
	_ = t.values.add(&EnumValue{Value: "MUTATION"})
	_ = t.values.add(&EnumValue{Value: "SUBSCRIPTION"})
	_ = t.values.add(&EnumValue{Value: "FIELD"})
	_ = t.values.add(&EnumValue{Value: "FRAGMENT_DEFINITION"})
	_ = t.values.add(&EnumValue{Value: "FRAGMENT_SPREAD"})
	_ = t.values.add(&EnumValue{Value: "INLINE_FRAGMENT"})
	_ = t.values.add(&EnumValue{Value: "SCHEMA"})
	_ = t.values.add(&EnumValue{Value: "SCALAR"})
	_ = t.values.add(&EnumValue{Value: "OBJECT"})
	_ = t.values.add(&EnumValue{Value: "FIELD_DEFINITION"})
	_ = t.values.add(&EnumValue{Value: "ARGUMENT_DEFINITION"})
	_ = t.values.add(&EnumValue{Value: "INTERFACE"})
	_ = t.values.add(&EnumValue{Value: "UNION"})
	_ = t.values.add(&EnumValue{Value: "ENUM"})
	_ = t.values.add(&EnumValue{Value: "ENUM_VALUE"})
	_ = t.values.add(&EnumValue{Value: "INPUT_OBJECT"})
	_ = t.values.add(&EnumValue{Value: "INPUT_FIELD_DEFINITION"})

	return &t
}

// Resolve root (__Schema) level fields.
func (root *Root) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case "types":
		result = root.types
	case "queryType":
		var t Type
		if root.schema != nil {
			if fd := root.schema.fields.get(string(OpQuery)); fd != nil {
				t = fd.Type
			}
		}
		result = t
	case "mutationType":
		var t Type
		if root.schema != nil {
			if fd := root.schema.fields.get(string(OpMutation)); fd != nil {
				t = fd.Type
			}
		}
		result = t
	case "subscriptionType":
		var t Type
		if root.schema != nil {
			if fd := root.schema.fields.get(string(OpSubscription)); fd != nil {
				t = fd.Type
			}
		}
		result = t
	case "directives":
		result = root.dirs
	}
	return
}

func (root *Root) subscribe(sub *Subscription) {
	sub.prep(root)
	root.subLock.Lock()
	root.subscriptions = append(root.subscriptions, sub)
	root.subLock.Unlock()
}

// Unsubscribe from an event stream.
func (root *Root) Unsubscribe(id string) (cnt int) {
	root.subLock.Lock()
	for i := len(root.subscriptions) - 1; 0 <= i; i-- {
		s := root.subscriptions[i]
		if s.sub.Match(id) {
			root.subscriptions = append(root.subscriptions[:i], root.subscriptions[i+1:]...)
			s.sub.Unsubscribe()
			cnt++
		}
	}
	root.subLock.Unlock()

	return
}

// AddEvent causes the event to be sent on any matching ID. The selection set
// for the subscription is used to form a result based on the type of event
// being published.
func (root *Root) AddEvent(id string, event interface{}) (cnt int, err error) {
	vars := map[string]interface{}{}
	var ea []error
	var failed []*Subscription
	root.subLock.Lock()
	for _, s := range root.subscriptions {
		if s.sub.Match(id) {
			result, ea2 := root.resolve(event, vars, s.field, s.field.ConType, MaxResolveDepth)
			ea = append(ea, ea2...)
			cnt++
			if err = s.sub.Send(result); err != nil {
				ea = append(ea, err)
				failed = append(failed, s)
			}
		}
	}
	root.subLock.Unlock()
	if 0 < len(ea) {
		err = Errors(ea)
	}
	root.subLock.Lock()
	for _, f := range failed {
		for i := len(root.subscriptions) - 1; 0 <= i; i-- {
			if f == root.subscriptions[i] {
				root.subscriptions = append(root.subscriptions[:i], root.subscriptions[i+1:]...)
				f.sub.Unsubscribe()
			}
		}
	}
	root.subLock.Unlock()

	return
}

func (root *Root) assureSchema() {
	if root.schema == nil {
		root.schema = &Schema{Object: Object{fields: fieldList{dict: map[string]*FieldDef{}}}}
		for _, cap := range []string{"Query", "Mutation", "Subscription"} {
			if t := root.types.get(cap); t != nil {
				name := strings.ToLower(cap)
				_ = root.schema.fields.add(&FieldDef{Base: Base{N: name}, Type: t})
			}
		}
	}
}
