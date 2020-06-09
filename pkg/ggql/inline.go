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

// Inline is a GraphQL execution request inline fragment.
type Inline struct {
	SelBase

	// Condition is the type that the fragment should be applied to.
	Condition Type
}

// String representation of the instance.
func (in *Inline) String() string {
	var b bytes.Buffer

	in.write(&b, 0)

	return b.String()
}

// Validate a type.
func (in *Inline) Validate(root *Root) (errs []error) {
	errs = append(errs, in.SelBase.Validate(root)...)
	for _, du := range in.Directives() {
		errs = append(errs, root.validateDirUse("...", Locate(in), du)...)
	}
	// Additional argument checks are performed during the resolve phase so no
	// need to attempt to validate argument type matching and coerce success.
	return
}

func (in *Inline) write(buf *bytes.Buffer, depth int) {
	_, _ = buf.WriteString(strings.Repeat("  ", depth))
	_, _ = buf.WriteString("...")
	if in.Condition != nil {
		_, _ = buf.WriteString(" on ")
		_, _ = buf.WriteString(in.Condition.Name())
	}
	_ = writeDirectiveUses(buf, in.Dirs)
	in.writeSels(buf, depth)
}
