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
	"time"

	"github.com/uhn/ggql/pkg/ggql"
)

var t0 = time.Unix(0, 1570305197*1000000000).In(time.UTC)

func TestTimeScalarCoerceIn(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Time")
	scalar, _ := tt.(ggql.InCoercer)
	checkNotNil(t, scalar, "Time type should be a Coercer")

	for _, v := range []interface{}{
		int64(1570305197),
		float64(1570305197.0),
		"2019-10-05T19:53:17Z",
	} {
		cv, err := scalar.CoerceIn(v)
		tx, ok := cv.(time.Time)
		checkEqual(t, true, ok, "CoerceIn should return a time.Time for a %T", v)
		checkEqual(t, t0.Format(time.RFC3339Nano), tx.Format(time.RFC3339Nano), "CoerceIn value mismatch for a %T", v)
		checkNil(t, err, "CoerceIn error. %s", err)
	}
	r, err := scalar.CoerceIn(nil)
	checkNil(t, r, "CoerceIn value mismatch")
	checkNil(t, err, "CoerceIn error. %s", err)

	_, err = scalar.CoerceIn(true)
	checkNotNil(t, err, "CoerceIn error")
}

func TestTimeScalarCoerceOut(t *testing.T) {
	root := ggql.NewRoot(nil)
	tt := root.GetType("Time")
	scalar, _ := tt.(ggql.OutCoercer)
	checkNotNil(t, scalar, "Time type should be a Coercer")

	for _, v := range []interface{}{
		int64(1570305197),
		float64(1570305197.0),
		"2019-10-05T19:53:17Z",
		t0,
	} {
		r, err := scalar.CoerceOut(v)
		s, ok := r.(string)
		checkEqual(t, true, ok, "CoerceOut should return a time.Time for a %T", v)
		checkEqual(t, t0.Format(time.RFC3339Nano), s, "CoerceOut value mismatch for a %T", v)
		checkNil(t, err, "CoerceOut error. %s", err)
	}
	r, err := scalar.CoerceOut(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	_, err = scalar.CoerceOut(true)
	checkNotNil(t, err, "CoerceOut error")
}
