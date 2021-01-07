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

// Directive is a GraphQL Directive.
type Directive struct {
	Base

	// Args for the directive.
	args argList

	// On the locations the directive can be used.
	On []Location
}

// Rank of the type.
func (t *Directive) Rank() int {
	return rankDirective
}

// String representation of the type.
func (t *Directive) String() string {
	return strings.TrimSpace(t.SDL())
}

// SDL returns an SDL representation of the type.
func (t *Directive) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Directive) Write(w io.Writer, desc bool) (err error) {
	if err = writeDesc(w, t.Desc, 0, desc); err == nil {
		if _, err = w.Write([]byte("directive @")); err == nil {
			if _, err = w.Write([]byte(t.N)); err == nil {
				if err = writeArgs(w, &t.args, desc); err == nil {
					if _, err = w.Write([]byte(" on ")); err == nil {
						for i, on := range t.On {
							if 0 < i {
								on = " | " + on
							}
							if _, err = w.Write([]byte(string(on))); err != nil {
								break
							}
						}
					}
				}
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte{'\n'})
	}
	return
}

// Extend a type.
func (t *Directive) Extend(x Type) error {
	return fmt.Errorf("%w, can not extend a directive (%s)", ErrValidation, t.Name())
}

// AddArg adds and argument to the directive.
func (t *Directive) AddArg(a *Arg) error {
	return t.args.add(a)
}

func (t *Directive) findArg(name string) *Arg {
	return t.args.get(name)
}

// Validate a type.
func (t *Directive) Validate(root *Root) (errs []error) {
	for _, loc := range t.On {
		if !IsLocation(loc) {
			errs = append(errs, fmt.Errorf("%w, %s is not a valid location at %d:%d",
				ErrValidation, loc, t.line, t.col))
		}
	}
	for _, a := range t.args.list {
		errs = append(errs, validateName(a.core, "argument", a.N, a.line, a.col)...)
		if co, _ := a.Type.(InCoercer); co != nil {
			if a.Default != nil {
				if v, err := co.CoerceIn(a.Default); err != nil {
					errs = append(errs, fmt.Errorf("%w at %d:%d", err, a.line, a.col))
				} else if v != a.Default {
					// Might as well replace the coerced value since it is really
					// what is needed.
					a.Default = v
				}
			}
		} else {
			errs = append(errs, fmt.Errorf("%w, directive %s argument %s, a %T is not an input type at %d:%d",
				ErrValidation, t.Name(), a.Name(), a.Type, a.line, a.col))
		}
		for _, du := range a.Directives() {
			errs = append(errs, root.validateDirUse(t.Name()+"."+a.Name(), Locate(a), du)...)
		}
	}
	if path := t.hasDirLoop(map[string]bool{t.Name(): true}); 0 < len(path) {
		errs = append(errs, fmt.Errorf("%w, directive %s has a directive loop - %s at %d:%d",
			ErrValidation, t.Name(), strings.Join(path, "->"), t.line, t.col))
	}
	return
}

func (t *Directive) hasDirLoop(hits map[string]bool) []string {
	for _, a := range t.args.list {
		for _, du := range a.Directives() {
			name := du.Directive.Name()
			if hits[name] {
				return []string{t.Name() + "." + a.Name(), name}
			}
			hits[name] = true
			if d2, _ := du.Directive.(*Directive); d2 != nil {
				if path := d2.hasDirLoop(hits); 0 < len(path) {
					return append([]string{t.Name() + "." + a.Name()}, path...)
				}
			}
		}
	}
	return nil
}

// Resolve returns one of the following:
//   name: String!
//   description: String
//   locations: [__DirectiveLocation!]!
//   args: [__InputValue!]!
func (t *Directive) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case nameStr:
		result = t.N
	case descriptionStr:
		result = t.Desc
	case locationsStr:
		list := make([]string, 0, len(t.On))
		for _, loc := range t.On {
			list = append(list, string(loc))
		}
		result = list
	case argsStr:
		result = &t.args
	}
	return
}
