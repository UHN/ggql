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

// FragRef is a reference to s Fragment.
type FragRef struct {

	// Fragment referenced.
	Fragment *Fragment

	// Dirs is an array of the Directive uses.
	Dirs []*DirectiveUse

	line int
	col  int
}

// String representation of the instance.
func (fr *FragRef) String() string {
	var b bytes.Buffer

	fr.write(&b, 0)

	return b.String()
}

// Directives returns the directive associated with the type.
func (fr *FragRef) Directives() []*DirectiveUse {
	return fr.Dirs
}

// SelectionSet returns the selections of the instance.
func (fr *FragRef) SelectionSet() []Selection {
	return fr.Fragment.Sels
}

// Validate a type.
func (fr *FragRef) Validate(root *Root) (errs []error) {
	for _, du := range fr.Directives() {
		errs = append(errs, root.validateDirUse(fr.Fragment.Name, Locate(fr), du)...)
	}
	return
}

// Line the selection was defined on in the request.
func (fr *FragRef) Line() int {
	return fr.line
}

// Column of the start of the selection in the request.
func (fr *FragRef) Column() int {
	return fr.col
}

func (fr *FragRef) write(buf *bytes.Buffer, depth int) {
	_, _ = buf.WriteString(strings.Repeat("  ", depth))
	_, _ = buf.WriteString("...")
	_, _ = buf.WriteString(fr.Fragment.Name)
}
