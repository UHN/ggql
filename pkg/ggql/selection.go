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

// Selection represents a member of a GraphQL selection set.
type Selection interface {
	fmt.Stringer

	// Directives returns the directive associated with the type.
	Directives() []*DirectiveUse

	// SelectionSet or child Selections.
	SelectionSet() []Selection

	// Validate a type.
	Validate(root *Root) []error

	// Line the selection was defined on in the request.
	Line() int

	// Column of the start of the selection in the request.
	Column() int

	write(buf *bytes.Buffer, depth int)
}
