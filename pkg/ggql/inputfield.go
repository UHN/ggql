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

// InputField in a representation of a field in an input object.
type InputField struct {
	Base

	// Type of the field return value.
	Type Type

	// Default value for the field.
	Default interface{}
}

// Write the type as SDL.
func (f *InputField) Write(w io.Writer, desc bool) (err error) {
	err = writeDesc(w, f.Desc, 1, desc)
	if err == nil {
		if !desc || len(f.Desc) == 0 {
			_, err = w.Write([]byte{' ', ' '})
		}
	}
	if err == nil {
		_, err = w.Write([]byte(f.N))
	}
	if err == nil {
		_, err = w.Write([]byte{':', ' '})
	}
	if err == nil {
		_, err = w.Write([]byte(f.Type.Name()))
	}
	if err == nil && f.Default != nil {
		_, err = w.Write([]byte{' ', '=', ' '})
		if err == nil {
			_, err = w.Write([]byte(valueString(f.Default)))
		}
	}
	if err == nil {
		err = writeDirectiveUses(w, f.Dirs)
	}
	if err == nil {
		_, err = w.Write([]byte{'\n'})
	}
	return
}

// Resolve returns one of the following:
//   name: String!
//   description: String
//   type: __Type!
//   defaultValue: String
func (f *InputField) Resolve(field *Field, args map[string]interface{}) (result interface{}, err error) {
	switch field.Name {
	case nameStr:
		result = f.N
	case descriptionStr:
		result = f.Desc
	case typeStr:
		result = f.Type
	case defaultValueStr:
		result = f.Default
	}
	return
}
