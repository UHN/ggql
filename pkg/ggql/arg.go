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

// Arg is a GraphQL Arg or InputValue.
type Arg struct {
	Base

	// Type for the arg value.
	Type Type

	// Default value for the argument.
	Default interface{}
}

// Write the type as SDL.
func (a *Arg) Write(w io.Writer, desc bool) (err error) {
	err = writeDesc(w, a.Desc, 2, desc)
	if err == nil {
		_, err = w.Write([]byte(a.N))
	}
	if err == nil {
		_, err = w.Write([]byte{':', ' '})
	}
	if err == nil {
		_, err = w.Write([]byte(a.Type.Name()))
	}
	if err == nil && a.Default != nil {
		_, err = w.Write([]byte{' ', '=', ' '})
		if err == nil {
			_, err = w.Write([]byte(valueString(a.Default)))
		}
	}
	if err == nil {
		err = writeDirectiveUses(w, a.Dirs)
	}
	return
}

func writeArgs(w io.Writer, args *argList, desc bool) (err error) {
	if 0 < args.Len() {
		if _, err = w.Write([]byte{'('}); err == nil {
			for i, a := range args.list {
				if 0 < i {
					_, err = w.Write([]byte{','})
					if err == nil && !desc || len(a.Desc) == 0 {
						_, err = w.Write([]byte{' '})
					}
				}
				if err == nil {
					err = a.Write(w, desc)
				}
				if err != nil {
					break
				}
			}
			if err == nil {
				_, err = w.Write([]byte{')'})
			}

		}
	}
	return
}

// Resolve returns one of the following:
//   name: String!
//   description: String
//   type: __Type!
//   defaultValue: String
func (a *Arg) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case nameStr:
		result = a.N
	case descriptionStr:
		result = a.Desc
	case typeStr:
		result = a.Type
	case defaultValueStr:
		result = a.Default
	}
	return
}
