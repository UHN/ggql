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
	"strings"
)

// Field of a selection.
type Field struct {
	SelBase

	// Alias of the field.
	Alias string

	// Name of the field.
	Name string

	// Arguments to the field.
	Args []*ArgValue

	// Type of the field container.
	ConType Type

	// Context is for user provided data and is only used by the Resolvers,
	// not this package.
	Context interface{}
}

// String representation of the instance.
func (f *Field) String() string {
	var b bytes.Buffer

	f.write(&b, 0)

	return b.String()
}

// Validate a type.
func (f *Field) Validate(root *Root) (errs []error) {
	if len(f.Name) == 0 {
		errs = append(errs, valError(f.line, f.col, "empty field name with alias %s", f.Alias))
	}
	errs = append(errs, f.SelBase.Validate(root)...)
	dups := map[string]bool{}
	for _, av := range f.Args {
		if _, has := dups[av.Arg]; has {
			errs = append(errs, valError(av.line, av.col, "duplicate argument %s to %s", av.Arg, f.Name))
		}
		dups[av.Arg] = true
	}
	for _, du := range f.Directives() {
		errs = append(errs, root.validateDirUse(f.Name, Locate(f), du)...)
	}
	// Additional argument checks are performed during the resolve phase so no
	// need to attempt to validate argument type matching and coerce success.
	return
}

func (f *Field) key() string {
	if 0 < len(f.Alias) {
		return f.Alias
	}
	return f.Name
}

func (f *Field) write(buf *bytes.Buffer, depth int) {
	_, _ = buf.WriteString(strings.Repeat("  ", depth))
	if 0 < len(f.Alias) {
		_, _ = buf.WriteString(f.Alias)
		_, _ = buf.WriteString(": ")
	}
	_, _ = buf.WriteString(f.Name)
	if 0 < len(f.Args) {
		_, _ = buf.WriteString("(")
		first := true
		for _, av := range f.Args {
			if av != nil {
				if first {
					first = false
				} else {
					_, _ = buf.WriteString(", ")
				}
				_ = av.Write(buf)
			}
		}
		_, _ = buf.WriteString(")")
	}
	_ = writeDirectiveUses(buf, f.Dirs)
	f.writeSels(buf, depth)
}

func (f *Field) getArg(name string) (av *ArgValue) {
	for _, a := range f.Args {
		if a.Arg == name {
			av = a
			break
		}
	}
	return
}

func (f *Field) sortArgs() (errors []error) {
	if 0 < len(f.Args) {
		if ot, _ := f.ConType.(*Object); ot != nil {
			if fd := ot.fields.get(f.Name); fd != nil {
				args := make([]*ArgValue, 0, len(f.Args))
				for _, a := range fd.args.list {
					args = append(args, f.getArg(a.N))
				}
				if len(args) != len(f.Args) {
					for _, av := range f.Args {
						if fd.getArg(av.Arg) == nil {
							errors = append(errors, valError(av.line, av.col, "%s is not an argument to %s", av.Arg, f.Name))
						}
					}
				}
				f.Args = args
			}
		}
	}
	return
}
