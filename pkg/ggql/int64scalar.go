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

type int64Scalar struct {
	Scalar
}

func newInt64Scalar() Type {
	return &int64Scalar{
		Scalar{
			Base: Base{
				N:    "Int64",
				core: true,
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (*int64Scalar) CoerceIn(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// remains nil
	case int64:
		// ok as is
	case int32:
		v = tv
	case string:
		var i int64
		if i, err = strconv.ParseInt(tv, 10, 64); err == nil {
			v = i
		}
	default:
		err = newCoerceErr(v, "Int64")
		v = nil
	}
	return v, err
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *int64Scalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
	// remains nil
	case float32:
		v = int64(tv)
	case float64:
		v = int64(tv)
	case int:
		v = int64(tv)
	case int8:
		v = int64(tv)
	case int16:
		v = int64(tv)
	case int32:
		v = int64(tv)
	case int64:
		// ok as is
	case uint:
		v = int64(tv)
	case uint8:
		v = int64(tv)
	case uint16:
		v = int64(tv)
	case uint32:
		v = int64(tv)
	case uint64:
		v = int64(tv)
	case string:
		var i int64
		if i, err = strconv.ParseInt(tv, 10, 64); err == nil {
			v = i
		}
	default:
		err = newCoerceErr(tv, "Int64")
		v = nil
	}
	return v, err
}
