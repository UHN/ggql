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
)

// Schema is a GraphQL schema.
type Schema struct {
	Object
}

// Rank of the type.
func (t *Schema) Rank() int {
	return rankSchema
}

// String representation of the type.
func (t *Schema) String() string {
	return t.SDL()
}

// SDL returns an SDL representation of the type.
func (t *Schema) SDL(desc ...bool) string {
	var b bytes.Buffer

	_ = t.Write(&b, 0 < len(desc) && desc[0])

	return b.String()
}

// Write the type as SDL.
func (t *Schema) Write(w io.Writer, desc bool) (err error) {
	return t.write(w, schemaStr, desc)
}

// Extend a type.
func (t *Schema) Extend(x Type) error {
	if ox, ok := x.(*Schema); ok { // Already checked so no need to report an error again.
		for _, fd := range ox.fields.list {
			if err := t.fields.add(fd); err != nil {
				return fmt.Errorf("%w: on %s", err, t.N)
			}
		}
	}
	return t.Object.Base.Extend(x)
}

// Validate a type.
func (t *Schema) Validate(root *Root) (errs []error) {
	for name, f := range t.fields.dict {
		switch name {
		case "query", "mutation", "subscription":
			// ok
		default:
			errs = append(errs, fmt.Errorf("%w: field %s on Schema not allowed at %d:%d",
				ErrValidation, name, f.line, f.col))
		}
	}
	errs = append(errs, t.Object.Validate(root)...)

	return
}
