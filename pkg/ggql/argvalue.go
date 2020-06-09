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

// Sort output values in Object type (maps).
var Sort = false

// ArgValue is a GraphQL Arg value.
type ArgValue struct {
	// Arg name associated with the value.
	Arg string

	// Value of the object.
	Value interface{}

	line int
	col  int
}

// Write the type as SDL.
func (av *ArgValue) Write(w io.Writer) (err error) {
	if _, err = w.Write([]byte(av.Arg + ": ")); err == nil {
		_, err = w.Write([]byte(valueString(av.Value)))
	}
	return
}
