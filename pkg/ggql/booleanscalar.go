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
	"strconv"
)

type booleanScalar struct {
	Scalar
}

func newBooleanScalar() Type {
	return &booleanScalar{
		Scalar{
			Base: Base{
				N:    "Boolean",
				core: true,
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (*booleanScalar) CoerceIn(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	if b, ok := v.(bool); ok {
		return b, nil
	}
	return nil, newCoerceErr(v, "Boolean")
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *booleanScalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case bool:
		// ok as is
	case float32:
		v = tv != 0.0
	case int32:
		v = tv != 0
	case string:
		var b bool
		if b, err = strconv.ParseBool(tv); err == nil {
			v = b
		}
	default:
		err = newCoerceErr(tv, "Boolean")
		v = nil
	}
	return v, err
}
