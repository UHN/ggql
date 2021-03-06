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

type floatScalar struct {
	Scalar
}

func newFloatScalar() Type {
	return &floatScalar{
		Scalar{
			Base: Base{
				N:    "Float",
				core: true,
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (*floatScalar) CoerceIn(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case float64:
		v = float32(tv)
	case float32:
		// ok as is
	case int32:
		v = float32(tv)
	case int64:
		v = float32(tv)
	default:
		v = nil
		err = newCoerceErr(tv, "Float")
	}
	return v, err
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *floatScalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case float32:
		// ok as is
	case float64:
		v = float32(tv)
	case int:
		v = float32(tv)
	case int8:
		v = float32(tv)
	case int16:
		v = float32(tv)
	case int32:
		v = float32(tv)
	case int64:
		v = float32(tv)
	case uint:
		v = float32(tv)
	case uint8:
		v = float32(tv)
	case uint16:
		v = float32(tv)
	case uint32:
		v = float32(tv)
	case uint64:
		v = float32(tv)
	case string:
		var f float64
		if f, err = strconv.ParseFloat(tv, 64); err == nil {
			v = float32(f)
		}
	default:
		v = nil
		err = newCoerceErr(tv, "Float")
	}
	return v, err
}
