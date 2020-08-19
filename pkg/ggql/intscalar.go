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

type intScalar struct {
	Scalar
}

func newIntScalar() Type {
	return &intScalar{
		Scalar{
			Base: Base{
				N:    "Int",
				core: true,
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (*intScalar) CoerceIn(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case int:
		v = int32(tv)
	case int8:
		v = int32(tv)
	case int16:
		v = int32(tv)
	case int32:
		// ok as is
	case int64:
		v = int32(tv)
	case uint:
		v = int32(tv)
	case uint8:
		v = int32(tv)
	case uint16:
		v = int32(tv)
	case uint32:
		v = int32(tv)
	case uint64:
		v = int32(tv)
	case float64:
		// Needed for nested types since the go JSON parser always emits float64 even if an integer.
		v = int32(tv)
		if float64(int32(tv)) != tv {
			err = newCoerceErr(v, "Int")
		}
	default:
		err = newCoerceErr(v, "Int")
		v = nil
	}
	return v, err
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *intScalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case float32:
		v = int32(tv)
	case float64:
		v = int32(tv)
	case int:
		v = int32(tv)
	case int8:
		v = int32(tv)
	case int16:
		v = int32(tv)
	case int32:
		// ok as is
	case int64:
		v = int32(tv)
	case uint:
		v = int32(tv)
	case uint8:
		v = int32(tv)
	case uint16:
		v = int32(tv)
	case uint32:
		v = int32(tv)
	case uint64:
		v = int32(tv)
	case string:
		var i int64
		if i, err = strconv.ParseInt(tv, 10, 64); err == nil {
			v = int32(i)
		}
	default:
		err = newCoerceErr(tv, "Int")
		v = nil
	}
	return v, err
}
