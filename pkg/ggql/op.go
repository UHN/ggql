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

// Op is a GraphQL execution request operation.
type Op struct {
	SelBase

	// Type is the type of operation.
	Type OpType

	// Name of the operation if any.
	Name string

	// Variables that can be passed to the operation.
	Variables []*VarDef
}

// String representation of the instance.
func (op *Op) String() string {
	var b bytes.Buffer

	op.write(&b)

	return b.String()
}

// Validate an operation.
func (op *Op) Validate(root *Root) (errs []error) {
	errs = append(errs, op.SelBase.Validate(root)...)
	for _, du := range op.Directives() {
		errs = append(errs, root.validateDirUse(op.Name, Locate(op), du)...)
	}
	for _, v := range op.Variables {
		errs = append(errs, v.Validate(root)...)
	}
	return
}

func (op *Op) write(buf *bytes.Buffer) {
	_, _ = buf.WriteString(string(op.Type))
	if 0 < len(op.Name) {
		_, _ = buf.WriteString(" ")
		_, _ = buf.WriteString(op.Name)
	}
	if 0 < len(op.Variables) {
		writeVarDefs(buf, op.Variables)
	}
	_ = writeDirectiveUses(buf, op.Dirs)
	op.writeSels(buf, 0)
	_, _ = buf.WriteString("\n")
}
