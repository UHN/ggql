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

func TestIntScalarCoerceIn(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Int")
	scalar, _ := tt.(ggql.InCoercer)
	checkNotNil(t, scalar, "Int type should be a Coercer")

	for _, v := range []interface{}{
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
		float64(3.0), // special case to deal with the golang parser behavior.
	} {
		r, err := scalar.CoerceIn(v)
		checkEqual(t, 3, r, "Int.CoerceIn(%v) value mismatch", v)
		checkNil(t, err, "Int.CoerceIn(%v) error. %s", v, err)
	}
	b, err := scalar.CoerceIn(int32(3))
	checkEqual(t, 3, b, "Int.CoerceIn(3) value mismatch")
	checkNil(t, err, "Int.CoerceIn(3) error. %s", err)

	_, err = scalar.CoerceIn(true)
	checkNotNil(t, err, "Int.CoerceIn(true) error")

	_, err = scalar.CoerceIn(3.2)
	checkNotNil(t, err, "Int.CoerceIn(3.2) error")

	v, err := scalar.CoerceIn(nil)
	checkNil(t, err, "Int.CoerceIn(nil) error. %s", err)
	checkNil(t, v, "Int.CoerceIn(nil) should return nil")
}

func TestIntScalarCoerceOut(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Int")
	scalar, _ := tt.(ggql.OutCoercer)
	checkNotNil(t, scalar, "Int type should be a OutCoercer")

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
		checkEqual(t, 3, r, "Int.CoerceOut(%v) value mismatch", v)
		checkNil(t, err, "Int.CoerceOut(%v) error. %s", v, err)
	}
	r, err := scalar.CoerceOut(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	_, err = scalar.CoerceOut("1.23")
	checkNotNil(t, err, `Int.CoerceOut("1.23") error`)

	_, err = scalar.CoerceOut(true)
	checkNotNil(t, err, "Int.CoerceOut(true) error")
}
