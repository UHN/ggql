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
)

// Type is the generic GraphQL type.
type Type interface {
	fmt.Stringer

	// SDL returns an SDL representation of the type with or without the
	// descriptions depending on the desc argument.
	SDL(desc ...bool) string

	// Write the type as SDL.
	Write(w io.Writer, desc bool) error

	// Return the rank of the type. Used in sorting the SDL output.
	Rank() int

	// Core returns true if the type is one of the built in types.
	Core() bool

	// Name returns the name of the type.
	Name() string

	// Description returns the description of the type.
	Description() string

	// Directives returns the directive associated with the type.
	Directives() []*DirectiveUse

	// Extend a type.
	Extend(x Type) error

	// Validate a type.
	Validate(root *Root) []error

	// Line the type was defined on in the schema.
	Line() int

	// Column the type was defined on in the schema.
	Column() int
}

// typeEqual returns true if a and b represent the same schema type. This works
// in cases where equality comparison will not, such as when a & b's underlying
// types are identical but their underlying values are non-nil pointers to
// different structs.
func typeEqual(a, b Type) bool {
	switch ta := a.(type) {
	case *List:
		if tb, ok := b.(*List); ok {
			return typeEqual(ta.Base, tb.Base)
		}
	case *NonNull:
		if tb, ok := b.(*NonNull); ok {
			return typeEqual(ta.Base, tb.Base)
		}
	}
	return a.Name() == b.Name()
}
