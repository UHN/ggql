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
	"time"
)

type timeScalar struct {
	Scalar
}

func newTimeScalar() Type {
	return &timeScalar{
		Scalar{
			Base: Base{
				N: "Time",
				Desc: `Time can be coerced to and from a string that is formatted according to the
RFC3339 specification with nanoseconds.`,
				core: false,
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (*timeScalar) CoerceIn(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case nil:
		// leave as nil
	case float64:
		secs := int64(tv)
		v = time.Unix(0, secs*int64(time.Second)).In(time.UTC).Add(time.Duration((tv - float64(secs)) * float64(time.Second)))
	case int64:
		v = time.Unix(0, tv*int64(time.Second)).In(time.UTC)
	case string:
		var t time.Time
		if t, err = time.Parse(time.RFC3339Nano, tv); err == nil {
			v = t
		}
	case time.Time:
		// Ok as is.
	default:
		err = newCoerceErr(tv, "Time")
		v = nil
	}
	return v, err
}

// CoerceOut coerces a result value into a type for the scalar. The time
// representation is a string.
func (t *timeScalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	var tt time.Time
	// Convert to time first to make sure the conversion to string is
	// correct. Even a string is parsed first to verify.
	switch tv := v.(type) {
	case nil:
		// remains nil
	case float64:
		secs := int64(tv)
		tt = time.Unix(0, secs*int64(time.Second)).In(time.UTC).Add(time.Duration((tv - float64(secs)) * float64(time.Second)))
	case int64:
		tt = time.Unix(0, tv*int64(time.Second)).In(time.UTC)
	case string:
		tt, err = time.Parse(time.RFC3339Nano, tv)
	case time.Time:
		tt = tv
	default:
		err = newCoerceErr(v, "Time")
		v = nil
	}
	if err == nil && v != nil {
		v = tt.In(time.UTC).Format(time.RFC3339Nano)
	}
	return v, err
}
