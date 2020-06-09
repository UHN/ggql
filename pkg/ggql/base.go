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
	"strings"
)

// Base of all Types. This is for the common members of all types. Note
// the short names. This is public so that it is easier to define a type
// without calling a multiple functions. Since this is an lower level
// technique the member names are not as descriptive as the interface
// functions that they back.
type Base struct {
	// N is the name of the type. Short for tighter declarations. Access other
	// than for creation is by the Name() method.
	N string

	// Desc is the description of the type.
	Desc string

	// Dirs is an array of the Directive uses.
	Dirs []*DirectiveUse

	core bool

	line int
	col  int
}

// Core returns true if the type is one of the built in types.
func (b *Base) Core() bool {
	return b.core
}

// Name returns the name of the type.
func (b *Base) Name() string {
	return b.N
}

// Description returns the description of the type.
func (b *Base) Description() string {
	return b.Desc
}

// Directives returns the directive associated with the type.
func (b *Base) Directives() []*DirectiveUse {
	return b.Dirs
}

// Line the type was defined on in the schema.
func (b *Base) Line() int {
	return b.line
}

// Column the type was defined on in the schema.
func (b *Base) Column() int {
	return b.col
}

// GetDirective returns the directive with the name provided or nil if not
// found.
func (b *Base) GetDirective(name string) (du *DirectiveUse) {
	for _, d := range b.Dirs {
		if dt, _ := d.Directive.(*Directive); dt != nil && dt.N == name {
			du = d
			break
		}
	}
	return
}

// Extend a type.
func (b *Base) Extend(x Type) error {
	for _, du := range x.Directives() {
		// Verify that the directive does not already exist.
		for _, exist := range b.Dirs {
			if du.Directive.Name() == exist.Directive.Name() {
				return fmt.Errorf("%w: directive %s already exists on %s", ErrDuplicate, du.Directive.Name(), b.N)
			}
		}
		b.Dirs = append(b.Dirs, du)
	}
	return nil
}

// Validate a type.
func (b *Base) Validate(root *Root) (errs []error) {
	return
}

func (b *Base) validateFieldDefs(typeName string, fields *fieldList) (errs []error) {
	if 0 < fields.Len() { // must have at least one field
		for _, f := range fields.list {
			if strings.HasPrefix(f.Name(), "__") {
				errs = append(errs, fmt.Errorf("%w, %s is not a valid field name, it begins with '__' at %d:%d",
					ErrValidation, f.Name(), f.line, f.col))
			}
			for _, a := range f.args.list {
				if strings.HasPrefix(a.Name(), "__") {
					errs = append(errs, fmt.Errorf("%w, %s is not a valid argument name, it begins with '__' at %d:%d",
						ErrValidation, a.Name(), a.line, a.col))
				}
				if !IsInputType(a.Type) {
					errs = append(errs, fmt.Errorf("%w, argument %s of %s must be an input type at %d:%d",
						ErrValidation, a.Name(), f.Name(), a.line, a.col))
				}
			}
			if !IsOutputType(f.Type) {
				errs = append(errs, fmt.Errorf("%w, %s does not return an output type at %d:%d",
					ErrValidation, f.Name(), f.line, f.col))
			}
		}
	} else {
		errs = append(errs, fmt.Errorf("%w, %s must have at least one field at %d:%d",
			ErrValidation, typeName, b.line, b.col))
	}
	return
}

func (b *Base) writeHeader(w io.Writer, kind string, withDesc bool, interfaces ...Type) (err error) {
	err = writeDesc(w, b.Desc, 0, withDesc)
	if err == nil {
		_, err = w.Write([]byte(kind))
	}
	if err == nil {
		_, err = w.Write([]byte(b.N))
	}
	if err == nil && 0 < len(interfaces) {
		_, err = w.Write([]byte(" implements "))
		for i, imp := range interfaces {
			if 0 < i {
				if _, err = w.Write([]byte{' ', '&', ' '}); err != nil {
					break
				}
			}
			if _, err = w.Write([]byte(imp.Name())); err != nil {
				break
			}
		}
	}
	if err == nil {
		err = writeDirectiveUses(w, b.Dirs)
	}
	if err == nil {
		_, err = w.Write([]byte{' ', '{', '\n'})
	}
	return
}

func writeDesc(w io.Writer, desc string, indent int, withDesc bool) (err error) {
	if len(desc) == 0 || !withDesc {
		return nil
	}
	if 0 < indent {
		if _, err = w.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	shift := strings.Repeat("  ", indent)
	if strings.ContainsAny(desc, "\n\"") {
		if _, err = w.Write([]byte(shift)); err == nil {
			shift = "\n" + shift
			if _, err = w.Write([]byte(`"""`)); err == nil {
				if _, err = w.Write([]byte(shift)); err == nil {
					if _, err = w.Write([]byte(strings.ReplaceAll(desc, "\n", shift))); err == nil {
						if _, err = w.Write([]byte(shift)); err == nil {
							_, err = w.Write([]byte(`"""`))
						}
					}
				}
			}
		}
	} else if _, err = w.Write([]byte(shift)); err == nil {
		if _, err = w.Write([]byte{'"'}); err == nil {
			if _, err = w.Write([]byte(desc)); err == nil {
				_, err = w.Write([]byte{'"', '\n'})
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte(shift))
	}
	return
}

// Ideally a default value should be specified but since the only current use
// cases have a default of false lint complains if a default argument is
// included.
func (b *Base) getBoolArg(args map[string]interface{}, key string) bool {
	if b, ok := args[key].(bool); ok {
		return b
	}
	return false
}
