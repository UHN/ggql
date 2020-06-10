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
	"sort"
)

// DirectiveUse is a GraphQL Directive in use.
type DirectiveUse struct {

	// Directive associated with the use.
	Directive Type

	// Args to the directive use.
	Args map[string]*ArgValue

	line int
	col  int
}

// Write the type as SDL.
func (dir *DirectiveUse) Write(w io.Writer) (err error) {
	if _, err = w.Write([]byte("@" + dir.Directive.Name())); err == nil {
		if 0 < len(dir.Args) {
			if _, err = w.Write([]byte{'('}); err == nil {
				keys := make([]string, 0, len(dir.Args))
				for _, a := range dir.Args {
					keys = append(keys, a.Arg)
				}
				sort.Strings(keys)
				for i, k := range keys {
					a := dir.Args[k]
					if 0 < i {
						if _, err = w.Write([]byte{',', ' '}); err != nil {
							break
						}
					}
					if err = a.Write(w); err != nil {
						break
					}
				}
				if err == nil {
					_, err = w.Write([]byte{')'})
				}
			}
		}
	}
	return
}

func writeDirectiveUses(w io.Writer, dirs []*DirectiveUse) (err error) {
	for _, du := range dirs {
		if _, err = w.Write([]byte{' '}); err == nil {
			err = du.Write(w)
		}
		if err != nil {
			break
		}
	}
	return
}
