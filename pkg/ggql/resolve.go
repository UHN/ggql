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
	"time"
)

// MaxResolveDepth is the maximum depth to allow during resolving. This will
// cause recursively nested queries to abort when the maximum depth is
// reached.
var MaxResolveDepth = 100

// ResolveBytes parses an SDL executable []byte and then evaluates it.
func (root *Root) ResolveBytes(src []byte, op string, vars map[string]interface{}) map[string]interface{} {
	return root.ResolveReader(bytes.NewReader(src), op, vars)
}

// ResolveString parses an SDL executable string and then evaluates it.
func (root *Root) ResolveString(src string, op string, vars map[string]interface{}) map[string]interface{} {
	return root.ResolveReader(strings.NewReader(src), op, vars)
}

// ResolveReader parses an SDL reader and then evaluates it.
func (root *Root) ResolveReader(r io.Reader, op string, vars map[string]interface{}) map[string]interface{} {
	var result map[string]interface{}
	exe, err := root.ParseExecutableReader(r)
	if err == nil {
		if result, err = root.ResolveExecutable(exe, op, vars); result == nil {
			result = map[string]interface{}{"data": nil}
		}
	}
	if err != nil {
		errors := FormErrorsResult(err)
		if result == nil {
			result = map[string]interface{}{"errors": errors}
		}
		result["errors"] = errors
	}
	return result
}

// ResolveExecutable resolves an Executable.
func (root *Root) ResolveExecutable(
	exe *Executable,
	opName string,
	vars map[string]interface{}) (result map[string]interface{}, err error) {

	// Returned error can be either an array of errors as a Errors, an Error,
	// or just a plain fmt.Errorf() return.

	root.assureSchema()
	op := exe.Ops[opName]
	if op == nil {
		if len(exe.Ops) == 1 {
			for _, o := range exe.Ops {
				op = o
				break
			}
		}
		if op == nil {
			return nil, fmt.Errorf("%w, could not determine operation to evaluate", ErrResolve)
		}
	}
	field := Field{Alias: "data", Name: string(op.Type), SelBase: SelBase{Sels: op.Sels}}

	var opVars map[string]interface{}
	if 0 < len(op.Variables) {
		opVars = map[string]interface{}{}
		for _, vd := range op.Variables {
			opVars[vd.Name] = vd.Default
			if vars != nil {
				if v := vars[vd.Name]; v != nil {
					if ic, _ := vd.Type.(InCoercer); ic != nil { // validated in SDL validation
						v, err = ic.CoerceIn(v)
					}
					if err != nil {
						err = resError(vd.line, vd.col, "%s for %s", err, vd.Name)
						return
					}
					opVars[vd.Name] = v
				}
			}
		}
	}
	result = map[string]interface{}{}
	if op.Type == OpSubscription {
		var ea []error
		if ea = root.resolveField(root.obj, opVars, &field, root.schema, result, 1); len(ea) == 0 {
			found := false
			subMap, _ := result["data"].(map[string]interface{})
			for _, val := range subMap {
				if sub, _ := val.(*Subscription); sub != nil {
					root.subscribe(sub)
					found = true
				}
			}
			if !found {
				ea = append(ea, resError(op.line, op.col, "no subscriptions resolved"))
			}
		}
		if 0 < len(ea) {
			err = Errors(ea)
		}
		return nil, err
	}
	if ea := root.resolveField(root.obj, opVars, &field, root.schema, result, MaxResolveDepth); 0 < len(ea) {
		err = Errors(ea)
	}
	return
}

func (root *Root) resolve(
	obj interface{},
	vars map[string]interface{},
	field *Field,
	t Type,
	depth int) (result interface{}, ea []error) {

	if depth <= 0 || IsNil(obj) {
		// If not intended then generate an error later when trying to
		// generate output.
		return obj, nil
	}
	switch tt := t.(type) {
	case *List:
		result, ea = root.resolveList(obj, vars, field, tt, depth-1)
	case *Object, *Schema, *Interface, *uuSchema:
		result, ea = root.resolveFieldSels(obj, vars, field, t, depth-1)
	case *NonNull:
		result, ea = root.resolve(obj, vars, field, tt.Base, depth)
	case *Union:
		resMap := map[string]interface{}{}
		result = resMap
		// Use reflection to get the type meta for the object then walk
		// through all the members of the union looking for a match. The
		// object may have its meta field already set but the first time it
		// will be nil so check for a @go directive then a type argument that
		// matches the object type. If there is a match then set the meta.
		objType := reflect.TypeOf(obj)
		for _, m := range tt.Members {
			if ot, _ := m.(*Object); ot != nil { // already checked in validation
				if ot.meta == nil {
					if err := ot.metaCheck(objType); err != nil {
						return nil, []error{err}
					}
				}
				if objType == ot.meta {
					result, ea = root.resolveFieldSels(obj, vars, field, m, depth-1)
					break
				}
			}
		}
	default:
		// Validation makes sure all output types are valid so no need to
		// check again here. The worse case is that null is returned if
		// someone finds a way to corrupt the return type.
		if co, _ := tt.(OutCoercer); co != nil {
			var err error
			if result, err = co.CoerceOut(obj); err != nil {
				ea = append(ea, resWarn(field.line, field.col, "%s", err))
			}
		}
	}
	return
}

func (root *Root) resolveFieldSels(
	obj interface{},
	vars map[string]interface{},
	field *Field,
	t Type,
	depth int) (result interface{}, ea []error) {

	mr := map[string]interface{}{}
	ea = root.resolveSels(obj, vars, field.Sels, t, mr, depth)
	result = mr

	return
}

func (root *Root) resolveSels(
	obj interface{},
	vars map[string]interface{},
	sels []Selection,
	t Type,
	result map[string]interface{},
	depth int) (ea []error) {

	if len(sels) == 0 {
		return []error{resWarnp(nil, "%s is not a valid output leaf type", t.Name())}
	}
	for _, sel := range sels {
		skip, ea2 := root.skipSel(sel, vars)
		ea = append(ea, ea2...)
		if skip {
			continue
		}
		switch ts := sel.(type) {
		case *Inline:
			ea2 = root.resolveInline(obj, vars, ts, t, result, depth)
		case *FragRef:
			ea2 = root.resolveFragRef(obj, vars, ts, t, result, depth)
		case *Field:
			ea2 = root.resolveField(obj, vars, ts, t, result, depth)
		}
		ea = append(ea, ea2...)
	}
	return
}

func (root *Root) skipSel(sel Selection, vars map[string]interface{}) (skip bool, ea []error) {
	for _, du := range sel.Directives() {
		switch du.Directive.Name() {
		case "skip":
			// Check on argument exist and type are checked during parse and
			// validation so no need to recheck here.
			if av := du.Args["if"]; av != nil {
				switch v := av.Value.(type) {
				case bool:
					skip = v
				case Var:
					if b, ok := vars[string(v)].(bool); ok {
						skip = b
					} else {
						skip = true // default to skipping
						ea = append(ea, resWarnp(sel, "%v is not a valid 'if' value for @skip", v))
					}
				}
			}
		case "include":
			// Check on argument exist and type are checked during parse and
			// validation so no need to recheck here.
			if av := du.Args["if"]; av != nil {
				switch v := av.Value.(type) {
				case bool:
					skip = !v
				case Var:
					if b, ok := vars[string(v)].(bool); ok {
						skip = !b
					} else {
						skip = true // default to skipping
						ea = append(ea, resWarnp(sel, "%v is not a valid 'if' value for @include", v))
					}
				}
			}
		}
	}
	return
}

func (root *Root) resolveList(
	obj interface{},
	vars map[string]interface{},
	field *Field,
	t *List,
	depth int) (result interface{}, ea []error) {

	var ea2 []error

	lt := t.Base
	switch list := obj.(type) {
	case ListResolver:
		var rlist []interface{}
		cnt := list.Len()
		var v interface{}
		for i := 0; i < cnt; i++ {
			v, ea2 = root.resolve(list.Nth(i), vars, field, lt, depth)
			Errors(ea2).in(i)
			ea = append(ea, ea2...)
			rlist = append(rlist, v)
		}
		result = rlist
	case []interface{}:
		rlist := make([]interface{}, 0, len(list))
		var v interface{}
		for i, x := range list {
			v, ea2 = root.resolve(x, vars, field, lt, depth)
			Errors(ea2).in(i)
			ea = append(ea, ea2...)
			rlist = append(rlist, v)
		}
		result = rlist
	case []string:
		rlist := make([]interface{}, 0, len(list))
		for _, s := range list {
			rlist = append(rlist, s)
		}
		result = rlist
	case []int:
		rlist := make([]interface{}, 0, len(list))
		for _, i := range list {
			rlist = append(rlist, i)
		}
		result = rlist
	case []int64:
		rlist := make([]interface{}, 0, len(list))
		for _, i := range list {
			rlist = append(rlist, i)
		}
		result = rlist
	case []bool:
		rlist := make([]interface{}, 0, len(list))
		for _, b := range list {
			rlist = append(rlist, b)
		}
		result = rlist
	case []float32:
		rlist := make([]interface{}, 0, len(list))
		for _, f := range list {
			rlist = append(rlist, f)
		}
		result = rlist
	case []float64:
		rlist := make([]interface{}, 0, len(list))
		for _, f := range list {
			rlist = append(rlist, f)
		}
		result = rlist
	case []time.Time:
		rlist := make([]interface{}, 0, len(list))
		for _, f := range list {
			rlist = append(rlist, f)
		}
		result = rlist
	default:
		if root.AnyResolver != nil {
			var rlist []interface{}
			cnt := root.AnyResolver.Len(obj)
			for i := 0; i < cnt; i++ {
				v, err := root.AnyResolver.Nth(obj, i)
				if err == nil {
					v, ea2 = root.resolve(v, vars, field, lt, depth)
					Errors(ea2).in(i)
					ea = append(ea, ea2...)
				} else {
					e := resWarnp(nil, "%s", err)
					e.(*Error).in(i)
					ea = append(ea, e)
				}
				rlist = append(rlist, v)
			}
			result = rlist
		} else {
			rv := reflect.ValueOf(obj)
			switch rv.Kind() {
			case reflect.Slice, reflect.Array:
				var rlist []interface{}
				cnt := rv.Len()
				for i := 0; i < cnt; i++ {
					v := rv.Index(i).Interface()
					v, ea2 = root.resolve(v, vars, field, lt, depth)
					rlist = append(rlist, v)
					Errors(ea2).in(i)
					ea = append(ea, ea2...)
				}
				result = rlist
			default:
				ea = append(ea, resWarn(field.line, field.col, "%T is not a list type", obj))
			}
		}
	}
	return
}

func (root *Root) formArgs(
	vars map[string]interface{},
	field *Field,
	fd *FieldDef) (args map[string]interface{}, ea []error) {

	required := map[string]bool{}
	if fd != nil {
		if fd.args.dict != nil {
			for k, a := range fd.args.dict {
				if _, ok := a.Type.(*NonNull); ok {
					required[k] = false
				}
			}
		}
	}
	// Build the args by combining provided args and variable values as
	// appropriate.
	if 0 < len(field.Args) {
		args = map[string]interface{}{}
		for _, av := range field.Args {
			if av != nil {
				var at Type
				if fd != nil {
					if a := fd.getArg(av.Arg); a != nil {
						at = a.Type
					}
				}
				if av.Value != nil {
					required[av.Arg] = true
				}
				var ea2 []error
				args[av.Arg], ea2 = root.replaceArgVars(vars, av.Value, at)
				Errors(ea2).in(av.Arg)
				ea = append(ea, ea2...)
			}
		}
	}
	for k, v := range required {
		if !v {
			ea = append(ea, resWarn(field.line, field.col, "%s is required but missing", k))
		}
	}
	return
}

func (root *Root) replaceArgVars(vars map[string]interface{}, v interface{}, at Type) (val interface{}, ea []error) {
	var err error
	var ea2 []error
	val = v
	switch tv := val.(type) {
	case Var:
		val = vars[string(tv)]
		if at != nil {
			if ic, _ := at.(InCoercer); ic != nil { // validated in SDL validation
				if val, err = ic.CoerceIn(val); err != nil {
					ea = append(ea, resWarnp(nil, "%s", err))
				}
			}
		}
	case map[string]interface{}:
		var it *Input
		if at != nil {
			it, _ = BaseType(at).(*Input)
		}
		required := map[string]bool{}
		if it != nil {
			for _, f := range it.fields.list {
				if _, ok := f.Type.(*NonNull); ok {
					required[f.N] = false
				}
			}
		}
		for k, v := range tv {
			var vt Type
			if it != nil {
				if f := it.fields.get(k); f != nil {
					vt = f.Type
					required[k] = true
				} else {
					ea = append(ea, resWarnp(nil, "%s not a field in %s", k, it.Name()))
				}
			}
			tv[k], ea2 = root.replaceArgVars(vars, v, vt)
			ea = append(ea, ea2...)
		}
		for k, v := range required {
			if !v {
				ea = append(ea, resWarnp(nil, "%s is required but missing", k))
			}
		}
	case []interface{}:
		var mt Type
		if at != nil {
			if lt, _ := at.(*List); lt != nil {
				mt = lt.Base
			}
		}
		for i, v := range tv {
			tv[i], ea2 = root.replaceArgVars(vars, v, mt)
			ea = append(ea, ea2...)
		}
	case Symbol:
		bt := BaseType(at)
		if et, _ := bt.(*Enum); et != nil {
			if _, has := et.values.dict[string(tv)]; !has {
				ea = append(ea, resWarnp(nil, "%s is not a valid enum value in %s", tv, et.N))
			}
		}
	default:
		if at != nil {
			if ic, _ := at.(InCoercer); ic != nil { // validated in SDL validation
				if val, err = ic.CoerceIn(val); err != nil {
					ea = append(ea, resWarnp(nil, "%s", err))
				}
			}
		}
	}
	return
}

func (root *Root) resolveField(
	obj interface{},
	vars map[string]interface{},
	field *Field,
	t Type,
	result map[string]interface{},
	depth int) (ea []error) {

	if field.ConType == nil {
		field.ConType = t
		ea = append(ea, field.sortArgs()...)
		if 0 < len(ea) {
			Errors(ea).in(field.key())
			return
		}
	}
	var ea2 []error
	switch field.Name {
	case "__typename":
		result[field.key()] = t.Name()
		return nil
	case "__type":
		if obj == root.queryObj {
			var fv interface{} // field value
			var av *ArgValue

			for _, av = range field.Args {
				if av.Arg == nameStr {
					break
				}
			}
			if av == nil {
				ea = append(ea, resWarnp(field, "__type meta-field is missing a name argument"))
			} else {
				var nv interface{}
				if vr, ok := av.Value.(Var); ok && vars != nil {
					nv = vars[string(vr)]
				} else {
					nv = av.Value
				}
				name, _ := nv.(string)
				t = root.GetType(name)
				if t != nil {
					fv, ea2 = root.resolve(t, vars, field, root.GetType("__Type"), depth)
					ea = append(ea, ea2...)
					Errors(ea).in(field.key())
				}
			}
			result[field.key()] = fv
			return
		}
		ea = append(ea, resWarnp(field, "__type meta-field is only on the query object"))
		return
	case "__schema":
		if obj == root.queryObj {
			var fv interface{} // field value

			fv, ea2 = root.resolve(root, vars, field, root.uuSchemaType, depth)
			ea = append(ea, ea2...)
			Errors(ea).in(field.key())
			result[field.key()] = fv
			return
		}
		ea = append(ea, resWarnp(field, "__schema meta-field is only on the query object"))
		return
	}
	// Find the attribute of the object first.
	var attr interface{}
	var err error

	res, _ := obj.(Resolver)
	fd := root.getFieldDef(t, field.Name)
	if fd == nil {
		ea = append(ea, resWarnp(field, "%s is not a field in %s", field.Name, t.Name()))
		return
	}
	switch {
	case res != nil:
		var args map[string]interface{}
		if args, ea2 = root.formArgs(vars, field, fd); len(ea2) == 0 {
			attr, err = res.Resolve(field, args)
		}
		ea = append(ea, ea2...)
	case root.AnyResolver != nil:
		var args map[string]interface{}
		if args, ea2 = root.formArgs(vars, field, fd); len(ea2) == 0 {
			attr, err = root.AnyResolver.Resolve(obj, field, args)
		}
		ea = append(ea, ea2...)
	default:
		attr, ea2 = root.resolveReflect(obj, vars, field, t)
		ea = append(ea, ea2...)
	}
	if err != nil {
		ea = root.addError(field, ea, err)
	}
	if MaxResolveDepth == depth && root.queryObj == nil && field.Name == string(OpQuery) {
		root.queryObj = attr
	}
	if IsNil(attr) {
		result[field.key()] = nil
	} else {
		var ft Type
		if fd != nil {
			ft = fd.Type
		}
		var fv interface{} // field value
		fv, ea2 = root.resolve(attr, vars, field, ft, depth)
		ea = append(ea, ea2...)
		result[field.key()] = fv
	}
	if depth < MaxResolveDepth {
		Errors(ea).in(field.key())
	}
	return
}

func (root *Root) addError(f *Field, ea []error, err error) []error {
	switch te := err.(type) {
	case Errors:
		for _, e := range te {
			ea = root.addError(f, ea, e)
		}
	default:
		if _, ok := err.(*Error); !ok {
			err = resWarn(f.line, f.col, "%s", err)
		}
		ea = append(ea, err)
	}
	return ea
}

func (root *Root) resolveReflect(
	obj interface{},
	vars map[string]interface{},
	field *Field,
	t Type) (value interface{}, ea []error) {

	ov := reflect.ValueOf(obj)
	var fd *FieldDef
	var err error
TOP:
	switch tt := t.(type) {
	case *Object:
		if tt.meta == nil {
			_ = root.regType(obj, tt)
		}
		if fd = tt.GetField(field.Name); fd != nil {
			if len(fd.goField) == 0 && fd.method == nil {
				err = root.regField(tt, fd, field.Name)
			}
		}
	case *Schema:
		if tt.meta == nil {
			_ = root.regType(obj, &tt.Object)
		}
		if fd = tt.GetField(field.Name); fd != nil {
			if len(fd.goField) == 0 && fd.method == nil {
				err = root.regField(&tt.Object, fd, field.Name)
			}
		}
	case *Interface:
		// Determine actual type based on the obj and try again.
		t = root.getReflectType(ov.Type())
		goto TOP
	}
	if err != nil {
		ea = append(ea, resWarn(field.line, field.col, "%s", err))
		return
	}
	if fd != nil {
		switch {
		case 0 < len(fd.goField):
			if ov.Kind() == reflect.Ptr {
				ov = ov.Elem()
			}
			if ov.Kind() == reflect.Struct {
				if fv := ov.FieldByName(fd.goField); fv.IsValid() {
					value = fv.Interface()
				}
			}
		case fd.method != nil:
			args := root.formReflectArgs(ov, vars, field)
			mva := fd.method.Call(args)
			switch len(mva) {
			case 1:
				value = mva[0].Interface()
			case 2: // assume (interface{}, error) return
				value = mva[0].Interface()
				if err, _ = mva[1].Interface().(error); err != nil {
					ea = append(ea, resWarn(field.line, field.col, "%s", err))
				}
			default:
				ea = append(ea, resWarn(field.line, field.col, "%T.%s returned more than 2 values", obj, field.Name))
			}
		}
	}
	return
}

func (root *Root) formReflectArgs(ov reflect.Value, vars map[string]interface{}, field *Field) (args []reflect.Value) {
	args = make([]reflect.Value, 0, len(field.Args)+1)
	args = append(args, ov)
	// Build the args by combining provided args and variable values as
	// appropriate.
	for _, av := range field.Args {
		if vr, ok := av.Value.(Var); ok && vars != nil {
			args = append(args, reflect.ValueOf(vars[string(vr)]))
		} else {
			args = append(args, reflect.ValueOf(av.Value))
		}
	}
	return
}

func (root *Root) resolveInline(
	obj interface{},
	vars map[string]interface{},
	sel *Inline,
	t Type,
	result map[string]interface{},
	depth int) (ea []error) {

	if sel.Condition == nil || sel.Condition == t {
		ea = root.resolveSels(obj, vars, sel.Sels, t, result, depth)
	}
	return
}

func (root *Root) resolveFragRef(
	obj interface{},
	vars map[string]interface{},
	sel *FragRef,
	t Type,
	result map[string]interface{},
	depth int) (ea []error) {

	if sel.Fragment.Condition == nil || sel.Fragment.Condition == t {
		ea = root.resolveSels(obj, vars, sel.Fragment.Sels, t, result, depth)
		if 0 < len(ea) {
			Errors(ea).in(fmt.Sprintf("fragment at %d:%d", sel.Line(), sel.Column()))
		}
	}
	return
}

func (root *Root) getFieldDef(t Type, name string) (fd *FieldDef) {
	switch tt := t.(type) {
	case *Object:
		fd = tt.GetField(name)
	case *uuSchema:
		fd = tt.GetField(name)
	case *Schema:
		fd = tt.GetField(name)
	case *Interface:
		fd = tt.GetField(name)
	}
	return
}

func (root *Root) getFieldType(t Type, name string) (ft Type) {
	if fd := root.getFieldDef(t, name); fd != nil {
		ft = fd.Type
	}
	return
}
