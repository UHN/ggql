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

func TestBooleanScalarCoerceIn(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Boolean")
	scalar, _ := tt.(ggql.InCoercer)
	checkNotNil(t, scalar, "Boolean type should be a Coercer")

	b, err := scalar.CoerceIn(true)
	checkEqual(t, true, b, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	_, err = scalar.CoerceIn(3)
	checkNotNil(t, err, "CoerceIn error")

	var r interface{}
	r, err = scalar.CoerceIn(nil)
	checkNil(t, r, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)
}

func TestBooleanScalarCoerceOut(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Boolean")
	scalar, _ := tt.(ggql.OutCoercer)
	checkNotNil(t, scalar, "Boolean type should be a Coercer")

	r, err := scalar.CoerceOut(float32(3.7))
	checkEqual(t, true, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	r, err = scalar.CoerceOut(float32(0.0))
	checkEqual(t, false, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	r, err = scalar.CoerceOut(int32(-3))
	checkEqual(t, true, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	r, err = scalar.CoerceOut(int32(0))
	checkEqual(t, false, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	r, err = scalar.CoerceOut("true")
	checkEqual(t, true, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	r, err = scalar.CoerceOut("false")
	checkEqual(t, false, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	_, err = scalar.CoerceOut("yes")
	checkNotNil(t, err, "CoerceOut error")

	_, err = scalar.CoerceOut([]string{})
	checkNotNil(t, err, "CoerceOut error")

	r, err = scalar.CoerceOut(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)
}
