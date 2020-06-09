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
)

// Fragment is a GraphQL execution request fragment.
type Fragment struct {
	Inline

	// Name of the fragment.
	Name string
}

// String representation of the instance.
func (f *Fragment) String() string {
	var b bytes.Buffer

	f.write(&b)

	return b.String()
}

// Validate a type.
func (f *Fragment) Validate(root *Root) (errs []error) {
	errs = append(errs, f.SelBase.Validate(root)...)
	for _, du := range f.Directives() {
		errs = append(errs, root.validateDirUse(f.Name, Locate(f), du)...)
	}
	// Additional argument checks are performed during the resolve phase so no
	// need to attempt to validate argument type matching and coerce success.
	return
}

func (f *Fragment) write(buf *bytes.Buffer) {
	_, _ = buf.WriteString("fragment ")
	_, _ = buf.WriteString(f.Name)
	if f.Condition != nil {
		_, _ = buf.WriteString(" on ")
		_, _ = buf.WriteString(f.Condition.Name())
	}
	_ = writeDirectiveUses(buf, f.Dirs)
	f.writeSels(buf, 0)
	_, _ = buf.WriteString("\n")
}
