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

func TestNonNull(t *testing.T) {
	root := ggql.NewRoot(nil)

	nn := ggql.NonNull{
		Base: root.GetType("String"),
	}
	checkEqual(t, false, nn.Core(), "NonNull Core() mismatch")
	checkEqual(t, "", nn.SDL(), "NonNull SDL() mismatch")
	checkEqual(t, "String!", nn.String(), "NonNull String() mismatch")
	checkEqual(t, 0, nn.Rank(), "NonNull Rank() mismatch")
	checkEqual(t, 0, nn.Line(), "NonNull Line() mismatch")
	checkEqual(t, 0, nn.Column(), "NonNull Column() mismatch")
	checkEqual(t, "", nn.Description(), "NonNull Description() mismatch")
	checkEqual(t, 0, len(nn.Directives()), "NonNull Directives() mismatch")
	checkNil(t, nn.Write(nil, false), "NonNull Write() mismatch")
	checkNil(t, nn.Extend(nil), "NonNull Extend() should not return an error")
	checkNil(t, nn.Validate(nil), "NonNull Validate() should not return an error")

	v, err := nn.CoerceOut(nil)
	checkNotNil(t, err, "NotNull.CoerceOut(nil) should return an error")
	checkNil(t, v, "NotNull.CoerceOut(nil) should return nil")

	v, err = nn.CoerceOut("abc")
	checkNil(t, err, `NotNull.CoerceOut("abc") should not return an error`)
	checkEqual(t, "abc", v, `NotNull.CoerceOut("abc") should return 37`)

	v, err = nn.Resolve(&ggql.Field{Name: "kind"}, nil)
	checkNil(t, err, "NonNull.Resolve(kind) should not return an error. %s", err)
	checkEqual(t, "NON_NULL", v, "NON_NULL string should be returned from NonNull.Resolve(kind)")

	v, err = nn.Resolve(&ggql.Field{Name: "interfaces"}, nil)
	checkNil(t, err, "NonNull.Resolve(kind) should not return an error. %s", err)
	checkNil(t, v, "NonNull.Resolve(interfaces) should return nil")

	_, err = nn.Resolve(&ggql.Field{Name: "bad"}, nil)
	checkNotNil(t, err, "NonNull.Resolve(bad) should return an error.")

	v, err = nn.Resolve(&ggql.Field{Name: "ofType"}, nil)
	checkNil(t, err, "NonNull.Resolve(ofType) should not return an error. %s", err)
	checkEqual(t, "String", v.(ggql.Type).Name(), "String type should be returned from NonNull.Resolve(kind)")
}
