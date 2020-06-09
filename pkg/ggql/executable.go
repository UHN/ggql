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
	"sort"
)

// Executable is a GraphQL execution request document. It does not include the
// operations name and variables values of a request but does define the
// elements of the executable.
type Executable struct {
	// Root schema to be used for the evaluation.
	Root *Root

	// Ops are the operations in the executable. Most commonly, only one.
	Ops map[string]*Op

	// Fragments supporting the operations.
	Fragments map[string]*Fragment
}

// String representation of the instance.
func (ex *Executable) String() string {
	var b bytes.Buffer

	ex.write(&b)

	return b.String()
}

// Validate an executable.
func (ex *Executable) Validate(root *Root) (errs []error) {
	for _, op := range ex.Ops {
		errs = append(errs, op.Validate(root)...)
	}
	for _, f := range ex.Fragments {
		errs = append(errs, f.Validate(root)...)
	}
	return
}

// SetContextRecursive sets the context for each field in the request tree
// below this instance. If the ctx argument implements the Nester interface
// then a new context is created by calling the New function on the Nester
// with the current field.
func (ex *Executable) SetContextRecursive(ctx interface{}) {
	for _, op := range ex.Ops {
		op.SetContextRecursive(ctx)
	}
}

func (ex *Executable) write(buf *bytes.Buffer) {
	for _, op := range ex.Ops {
		op.write(buf)
	}
	if 0 < len(ex.Fragments) {
		keys := make([]string, 0, len(ex.Fragments))
		for k := range ex.Fragments {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			frag := ex.Fragments[k]
			_, _ = buf.WriteString("\n")
			frag.write(buf)
		}
	}
}
