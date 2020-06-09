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

type stringScalar struct {
	Scalar
}

func newStringScalar() Type {
	return &stringScalar{
		Scalar{
			Base: Base{
				N:    "String",
				core: true,
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (t *stringScalar) CoerceIn(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	if s, ok := v.(string); ok && v != nil {
		return s, nil
	}
	return nil, newCoerceErr(v, t.N)
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *stringScalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
	// remains nil
	case bool:
		if tv {
			v = trueStr
		} else {
			v = falseStr
		}
	case float32:
		v = strconv.FormatFloat(float64(tv), 'g', -1, 32)
	case float64:
		v = strconv.FormatFloat(tv, 'g', -1, 64)
	case int:
		v = strconv.Itoa(tv)
	case int8:
		v = strconv.Itoa(int(tv))
	case int16:
		v = strconv.Itoa(int(tv))
	case int32:
		v = strconv.Itoa(int(tv))
	case int64:
		v = strconv.FormatInt(tv, 10)
	case uint:
		v = strconv.Itoa(int(tv))
	case uint8:
		v = strconv.Itoa(int(tv))
	case uint16:
		v = strconv.Itoa(int(tv))
	case uint32:
		v = strconv.Itoa(int(tv))
	case uint64:
		v = strconv.FormatInt(int64(tv), 10)
	case string:
		// ok as is
	default:
		err = newCoerceErr(tv, "String")
		v = nil
	}
	return v, err
}
