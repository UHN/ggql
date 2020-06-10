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
)

// VarDef is a GraphQL variable definition.
type VarDef struct {

	// Name of the variable.
	Name string

	// Type for the varDef value.
	Type Type

	// Default value for the var.
	Default interface{}

	// Dirs is an array of the Directive uses.
	Dirs []*DirectiveUse

	line int
	col  int
}

// Validate a variable definition.
func (v *VarDef) Validate(root *Root) (errs []error) {
	if !IsInputType(v.Type) {
		errs = append(errs, fmt.Errorf("%w: %s is not a valid input type for $%s at %d:%d",
			ErrValidation, v.Type.Name(), v.Name, v.line, v.col))
	}
	for _, du := range v.Dirs {
		errs = append(errs, root.validateDirUse(v.Name, Locate(v), du)...)
	}
	return
}

func (v *VarDef) write(buf *bytes.Buffer) {
	_, _ = buf.WriteString("$")
	_, _ = buf.WriteString(v.Name)
	_, _ = buf.WriteString(": ")
	_, _ = buf.WriteString(v.Type.Name())
	if v.Default != nil {
		_, _ = buf.WriteString(" = ")
		_, _ = buf.WriteString(valueString(v.Default))
	}
	_ = writeDirectiveUses(buf, v.Dirs)
}

func writeVarDefs(buf *bytes.Buffer, vars []*VarDef) {
	if 0 < len(vars) {
		_, _ = buf.WriteString("(")
		for i, v := range vars {
			if 0 < i {
				_, _ = buf.WriteString(", ")
			}
			v.write(buf)
		}
		_, _ = buf.WriteString(")")
	}
}
