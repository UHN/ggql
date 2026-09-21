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
	"errors"
	"unsafe"
)

// IsNil checks for a nil value of an interface. Go values have two components
// not exposed, a type component and a value component. Further reading:
// https://research.swtch.com/interfaces. To ascertain whether the value is
// nil we ignore the type component and just check if the value component is
// set to 0.
func IsNil(v interface{}) bool {
	return (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0
}

// BaseType returns the base type. List and NonNull return their Base. The
// function recurses until a non List or NonNull is returned.
func BaseType(t Type) Type {
	if t != nil {
		switch tt := t.(type) {
		case *NonNull:
			t = BaseType(tt.Base)
		case *List:
			t = BaseType(tt.Base)
		}
	}
	return t
}

// FormErrorsResult forms an errors array suitable for returning from GraphQL
// request. The result will include path and location when possible.
func FormErrorsResult(err error) []interface{} {
	eList := []interface{}{}
	var ea Errors

	if errors.As(err, &ea) {
		for _, e := range ea {
			eList = append(eList, formOneErrorResult(e))
		}
	} else {
		eList = append(eList, formOneErrorResult(err))
	}
	return eList
}

func formOneErrorResult(err error) map[string]interface{} {
	em := map[string]interface{}{}
	var e *Error
	if errors.As(err, &e) {
		em["message"] = e.Base.Error()
		if 0 < e.Line || 0 < e.Column {
			em["locations"] = []interface{}{map[string]interface{}{"line": e.Line, "column": e.Column}}
		}
		if 0 < len(e.Path) {
			em["path"] = e.Path
		}
		if e.Extensions != nil {
			em["extensions"] = e.Extensions
		}
	} else {
		em["message"] = err.Error()
	}
	return em
}

// unwrapNonNull removes any NonNull wrappers from a type, leaving lists intact.
func unwrapNonNull(t Type) Type {
	if nn, _ := t.(*NonNull); nn != nil {
		return unwrapNonNull(nn.Base)
	}
	return t
}

// symbolizeEnums converts the strings in a variable value into enum Symbols
// wherever the declared type expects an enum, descending through lists and
// input objects.
//
// JSON has no enum type, so a client transports an enum value as a string; the
// specification allows that for variable values while requiring that a string
// literal written into a document be refused. Converting here, against the
// variable's declared type, keeps those two cases apart: a variable's string
// becomes a Symbol and is then validated against the enum's members by the
// ordinary coercion, exactly as a literal token is, while a string appearing in
// the document is untouched and still refused.
func symbolizeEnums(t Type, v interface{}) interface{} {
	if t == nil {
		return v
	}
	switch tv := v.(type) {
	case string:
		if _, ok := BaseType(t).(*Enum); ok {
			return Symbol(tv)
		}
	case []interface{}:
		if lt, _ := unwrapNonNull(t).(*List); lt != nil {
			for i, m := range tv {
				tv[i] = symbolizeEnums(lt.Base, m)
			}
		}
	case map[string]interface{}:
		if it, _ := BaseType(t).(*Input); it != nil {
			for k, m := range tv {
				if f := it.fields.get(k); f != nil {
					tv[k] = symbolizeEnums(f.Type, m)
				}
			}
		}
	}
	return v
}
