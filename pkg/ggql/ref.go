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
	"io"
)

// Ref is used as a place holder in a two pass parse and resolve when parsing
// an SDL. It is public so it can be tested for 100% coverage.
type Ref struct {
	Base
}

// String representation of the instance.
func (t *Ref) String() string {
	return t.N
}

// SDL returns an SDL representation of the type.
func (t *Ref) SDL(desc ...bool) string {
	return ""
}

// Write the type as SDL.
func (t *Ref) Write(w io.Writer, desc bool) error {
	return nil
}

// Rank returns the rank of the type. Used in sorting the SDL output.
func (t *Ref) Rank() int {
	return rankHidden
}
