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

// Symbol represents a enum value and is displayed without quotes.
type Symbol string

// EnumValue is a GraphQL Enum value.
type EnumValue struct {

	// Value of the value.
	Value Symbol

	// Description of the value.
	Description string

	// Directives associated with the value.
	Directives []*DirectiveUse

	line int
	col  int
}

// Write the type as SDL.
func (ev *EnumValue) Write(w io.Writer, desc bool) (err error) {
	if err = writeDesc(w, ev.Description, 1, desc); err == nil {
		if !desc || len(ev.Description) == 0 {
			_, err = w.Write([]byte{' ', ' '})
		}
		if err == nil {
			if _, err = w.Write([]byte(ev.Value)); err == nil {
				err = writeDirectiveUses(w, ev.Directives)
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte{'\n'})
	}
	return
}

// Resolve returns one of the following:
//   name: String!
//   description: String
//   isDeprecated: Boolean!
//   deprecationReason: String
func (ev *EnumValue) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case nameStr:
		result = string(ev.Value)
	case descriptionStr:
		result = ev.Description
	case isDeprecatedStr:
		result = false
		for _, du := range ev.Directives {
			if du.Directive.Name() == deprecatedStr {
				result = true
			}
		}
	case deprecationReasonStr:
		for _, du := range ev.Directives {
			if du.Directive.Name() == deprecatedStr {
				if av := du.Args["reason"]; av != nil {
					result = av.Value
				}
			}
		}
	}
	return
}

func (ev *EnumValue) isDeprecated() (isDep bool) {
	for _, du := range ev.Directives {
		if du.Directive.Name() == deprecatedStr {
			isDep = true
			break
		}
	}
	return
}
