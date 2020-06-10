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

// SelBase is the base for all selections.
type SelBase struct {
	// Dirs is an array of the Directive uses.
	Dirs []*DirectiveUse

	// Sels is an array of the sub-selections in the current selection.
	Sels []Selection

	line int
	col  int
}

// Directives returns the directive associated with the type.
func (sb *SelBase) Directives() []*DirectiveUse {
	return sb.Dirs
}

// SelectionSet returns the selections of the instance.
func (sb *SelBase) SelectionSet() []Selection {
	return sb.Sels
}

// Validate fields in the instance.
func (sb *SelBase) Validate(root *Root) (errs []error) {
	for _, sel := range sb.Sels {
		errs = append(errs, sel.Validate(root)...)
	}
	return
}

// Line the selection was defined on in the request.
func (sb *SelBase) Line() int {
	return sb.line
}

// Column of the start of the selection in the request.
func (sb *SelBase) Column() int {
	return sb.col
}

// SetContextRecursive sets the context for each field in the request tree
// below this instance. If the ctx argument implements the Nester interface
// then a new context is created by calling the New function on the Nester
// with the current field.
func (sb *SelBase) SetContextRecursive(ctx interface{}) {
	for _, sel := range sb.Sels {
		switch ts := sel.(type) {
		case *Field:
			if n, _ := ctx.(Nester); n != nil {
				ctx = n.Nest(ts)
			}
			ts.Context = ctx
			ts.SetContextRecursive(ts.Context)
		case *Inline:
			ts.SetContextRecursive(ctx)
		case *FragRef:
			ts.Fragment.SetContextRecursive(ctx)
		}
	}
}

func (sb *SelBase) writeSels(buf *bytes.Buffer, depth int) {
	if 0 < len(sb.Sels) {
		d2 := depth + 1
		_, _ = buf.WriteString(" {\n")
		for _, s := range sb.Sels {
			s.write(buf, d2)
			_, _ = buf.WriteString("\n")
		}
		if 0 < depth {
			_, _ = buf.WriteString(strings.Repeat("  ", depth))
		}
		_, _ = buf.WriteString("}")
	}
}
