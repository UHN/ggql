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

package ggql_test

import (
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestFloatScalarCoerceIn(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Float")
	scalar, _ := tt.(ggql.InCoercer)
	checkNotNil(t, scalar, "Float type should be a Coercer")

	for _, v := range []interface{}{
		float32(3.0),
		float64(3.0),
		int32(3),
		int64(3),
	} {
		f, err := scalar.CoerceIn(v)
		checkNil(t, err, "Float.CoerceIn error. %s", err)
		checkEqual(t, float32(3.0), f, "Float.CoerceIn value mismatch for %T %v", v, v)
	}
	v, err := scalar.CoerceIn(nil)
	checkNil(t, err, "Float.CoerceIn(nil) error. %s", err)
	checkNil(t, v, "Float.CoerceIn(nil) should return nil")

	_, err = scalar.CoerceIn(true)
	checkNotNil(t, err, "Float.CoerceIn error")

	_, err = scalar.CoerceIn("3.3")
	checkNotNil(t, err, "Float.CoerceIn error")
}

func TestFloatScalarCoerceOut(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Float")
	scalar, _ := tt.(ggql.OutCoercer)
	checkNotNil(t, scalar, "Float type should be a Coercer")

	for _, v := range []interface{}{
		float32(3.0),
		float64(3.0),
		int(3),
		int8(3),
		int16(3),
		int32(3),
		int64(3),
		uint(3),
		uint8(3),
		uint16(3),
		uint32(3),
		uint64(3),
		"3.0",
	} {
		r, err := scalar.CoerceOut(v)
		checkEqual(t, float32(3.0), r, "Float.CoerceOut value mismatch for %T", v)
		checkNil(t, err, "Float.CoerceOut error. %s", err)
	}
	r, err := scalar.CoerceOut(nil)
	checkNil(t, r, "Float.CoerceOut value mismatch")
	checkNil(t, err, "Float.CoerceOut error. %s", err)

	_, err = scalar.CoerceOut(true)
	checkNotNil(t, err, "Float.CoerceOut error")
}
