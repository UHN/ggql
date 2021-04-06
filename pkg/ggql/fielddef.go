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
	"io"
	"reflect"
	"sync"
)

// FieldDef in a representation of a field in an object or interface.
type FieldDef struct {
	Base

	// Type of the field return value.
	Type Type

	// Args for the field.
	args argList

	method  *reflect.Value
	goField string
	mu      sync.Mutex
}

// Write the type as SDL.
func (f *FieldDef) Write(w io.Writer, desc bool) (err error) {
	err = writeDesc(w, f.Desc, 1, desc)
	if err == nil {
		if !desc || len(f.Desc) == 0 {
			_, err = w.Write([]byte{' ', ' '})
		}
	}
	if err == nil {
		_, err = w.Write([]byte(f.N))
	}
	if err == nil {
		err = writeArgs(w, &f.args, desc)
	}
	if err == nil {
		_, err = w.Write([]byte{':', ' '})
	}
	if err == nil {
		_, err = w.Write([]byte(f.Type.Name()))
	}
	if err == nil {
		err = writeDirectiveUses(w, f.Dirs)
	}
	if err == nil {
		_, err = w.Write([]byte{'\n'})
	}
	return
}

// AddArg adds and argument.
func (f *FieldDef) AddArg(a *Arg) error {
	return f.args.add(a)
}

// Args returns a list of Args.
func (f *FieldDef) Args() []*Arg {
	return f.args.list
}

func (f *FieldDef) getArg(name string) *Arg {
	return f.args.get(name)
}

func (f *FieldDef) isDeprecated() bool {
	return f.GetDirective(deprecatedStr) != nil
}

// Resolve returns one of the following:
//   name: String!
//   description: String
//   args: [__InputValue!]!
//   type: __Type!
//   isDeprecated: Boolean!
//   deprecationReason: String
func (f *FieldDef) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case nameStr:
		result = f.N
	case descriptionStr:
		result = f.Desc
	case argsStr:
		result = &f.args
	case typeStr:
		result = f.Type
	case isDeprecatedStr:
		result = false
		for _, du := range f.Dirs {
			if du.Directive.Name() == deprecatedStr {
				result = true
			}
		}
	case deprecationReasonStr:
		for _, du := range f.Dirs {
			if du.Directive.Name() == deprecatedStr {
				if av := du.Args[reasonStr]; av != nil {
					result = av.Value
				}
			}
		}
	}
	return
}
