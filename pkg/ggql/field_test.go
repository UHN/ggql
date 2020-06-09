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

func TestField(t *testing.T) {
	dirt := ggql.Directive{
		Base: ggql.Base{N: "dirt"},
		On:   []ggql.Location{ggql.LocInlineFragment, ggql.LocField},
	}
	f := ggql.Field{
		Name: "user",
		Args: []*ggql.ArgValue{
			{Arg: "name", Value: "Gobo"},
			{Arg: "age", Value: 50},
		},
		SelBase: ggql.SelBase{Dirs: []*ggql.DirectiveUse{{Directive: &dirt}}},
	}
	checkEqual(t, `user(name: "Gobo", age: 50) @dirt`, f.String(), "Field.String() mismatch")

	checkNotNil(t, f.Directives(), "directives should not be nil")
	checkNotNil(t, f.SelectionSet(), "selections should not be nil")
}
