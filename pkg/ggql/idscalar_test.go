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
	"fmt"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestIDScalarCoerceIn(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("ID")
	scalar, _ := tt.(ggql.InCoercer)
	checkNotNil(t, scalar, "ID type should be a Coercer")

	id, err := scalar.CoerceIn("abc123")
	checkEqual(t, "abc123", id, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	id, err = scalar.CoerceIn(int32(123456))
	checkEqual(t, "123456", id, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	id, err = scalar.CoerceIn(int64(123456))
	checkEqual(t, "123456", id, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	_, err = scalar.CoerceIn(true)
	checkNotNil(t, err, "CoerceIn error")

	var r interface{}
	r, err = scalar.CoerceIn(nil)
	checkNil(t, r, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)
}

func TestIDScalarCoerceOut(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("ID")
	scalar, _ := tt.(ggql.OutCoercer)
	checkNotNil(t, scalar, "ID type should be a Coercer")

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
		"rope",
	} {
		r, err := scalar.CoerceOut(v)
		checkEqual(t, fmt.Sprintf("%v", v), r, "CoerceOut value mismatch")
		checkNil(t, err, "CoerceOut error. %s", err)
	}
	r, err := scalar.CoerceOut(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	_, err = scalar.CoerceOut([]string{})
	checkNotNil(t, err, "CoerceOut error")
}
