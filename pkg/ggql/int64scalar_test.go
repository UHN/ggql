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

func TestInt64ScalarCoerceIn(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Int64")
	scalar, _ := tt.(ggql.InCoercer)
	checkNotNil(t, scalar, "Int64 type should be a Coercer")

	i, err := scalar.CoerceIn(int32(3))
	checkEqual(t, 3, i, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	i, err = scalar.CoerceIn(int64(3))
	checkEqual(t, 3, i, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	i, err = scalar.CoerceIn("321")
	checkEqual(t, 321, i, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	_, err = scalar.CoerceIn(true)
	checkNotNil(t, err, "CoerceIn error")

	i, err = scalar.CoerceIn(nil)
	checkNil(t, i, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)
}

func TestInt64ScalarCoerceOut(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Int64")
	scalar, _ := tt.(ggql.OutCoercer)
	checkNotNil(t, scalar, "Int64 type should be a Coercer")

	for _, v := range []interface{}{
		float32(3.1),
		float64(3.3),
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
		"3",
	} {
		r, err := scalar.CoerceOut(v)
		checkEqual(t, 3, r, "CoerceOut value mismatch")
		checkNil(t, err, "CoerceOut error. %s", err)
	}
	r, err := scalar.CoerceOut(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	_, err = scalar.CoerceOut("1.23")
	checkNotNil(t, err, "CoerceOut error")

	_, err = scalar.CoerceOut(true)
	checkNotNil(t, err, "CoerceOut error")
}
