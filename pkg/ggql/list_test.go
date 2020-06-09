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

func TestList(t *testing.T) {
	root := ggql.NewRoot(nil)

	list := ggql.List{
		Base: &ggql.NonNull{Base: root.GetType("String")},
	}
	checkEqual(t, false, list.Core(), "List Core() mismatch")
	checkEqual(t, "", list.SDL(), "List SDL() mismatch")
	checkEqual(t, "[String!]", list.String(), "List String() mismatch")
	checkEqual(t, 0, list.Rank(), "List Rank() mismatch")
	checkEqual(t, 0, list.Line(), "List Line() mismatch")
	checkEqual(t, 0, list.Column(), "List Column() mismatch")
	checkEqual(t, "", list.Description(), "List Description() mismatch")
	checkEqual(t, 0, len(list.Directives()), "List Directives() mismatch")
	checkNil(t, list.Write(nil, false), "List Write() mismatch")
	checkNil(t, list.Extend(nil), "List Extend() should not return an error")
	checkNil(t, list.Validate(nil), "List Validate() should return nil")

	v, err := list.CoerceOut(nil)
	checkNil(t, err, "List.CoerceOut(nil) should not return an error")
	checkNil(t, v, "List.CoerceOut(nil) should return nil")

	v, err = list.CoerceOut([]interface{}{"abc", 37})
	checkNil(t, err, "List.CoerceOut([abc 37]) should not return an error")
	checkEqual(t, "[abc 37]", fmt.Sprintf("%v", v), "List.CoerceOut([abc 37]) should return [abc 37]")

	_, err = list.CoerceOut([]interface{}{nil, 37})
	checkNotNil(t, err, "List.CoerceOut([nil 37]) should return an error")

	_, err = list.CoerceOut(37)
	checkNotNil(t, err, "List.CoerceOut(37) should return an error")

	v, err = list.CoerceIn(nil)
	checkNil(t, err, "List.CoerceIn(nil) should not return an error")
	checkNil(t, v, "List.CoerceIn(nil) should return nil")

	_, err = list.CoerceIn([]interface{}{"abc", 37})
	checkNotNil(t, err, "List.CoerceIn([abc 37]) should return an error")

	_, err = list.CoerceIn(37)
	checkNotNil(t, err, "List.CoerceIn(37) should return an error")

	v, err = list.Resolve(&ggql.Field{Name: "ofType"}, nil)
	checkNil(t, err, "NonNull.Resolve(ofType) should not return an error. %s", err)
	checkEqual(t, "String!", v.(ggql.Type).Name(), "String type should be returned from List.Resolve(kind)")

	base := ggql.BaseType(&list)
	checkNotNil(t, base, "BaseType should not return nil")
	checkEqual(t, "String", base.Name(), "Base name of String! should be String")

	list = ggql.List{Base: &ggql.Object{Base: ggql.Base{N: "NotInput"}}}
	_, err = list.CoerceIn([]interface{}{})
	checkNotNil(t, err, "List.CoerceIn() with a non-Coercer base should return an error")

	v, err = list.Resolve(&ggql.Field{Name: "kind"}, nil)
	checkNil(t, err, "List.Resolve(kind) should not return an error. %s", err)
	checkEqual(t, "LIST", v, "LIST should be returned from List.Resolve(kind)")

	v, err = list.Resolve(&ggql.Field{Name: "interfaces"}, nil)
	checkNil(t, err, "List.Resolve(kind) should not return an error. %s", err)
	checkNil(t, v, "List.Resolve(interfaces) should return nil")

	_, err = list.Resolve(&ggql.Field{Name: "bad"}, nil)
	checkNotNil(t, err, "List.Resolve(bad) should return an error.")
}
