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

func TestUnion(t *testing.T) {
	root := ggql.NewRoot(nil)

	dirt := ggql.Directive{
		Base: ggql.Base{
			N: "dirt",
		},
		On: []ggql.Location{ggql.LocUnion},
	}
	a := &ggql.Object{Base: ggql.Base{N: "Aa"}}
	a.AddField(&ggql.FieldDef{Base: ggql.Base{N: "a"}, Type: root.GetType("String")})

	b := &ggql.Object{Base: ggql.Base{N: "Bb"}}
	b.AddField(&ggql.FieldDef{Base: ggql.Base{N: "b"}, Type: root.GetType("String")})
	u := &ggql.Union{
		Base: ggql.Base{
			N:    "Uu",
			Desc: "Union of A and B.",
			Dirs: []*ggql.DirectiveUse{{Directive: &dirt}},
		},
		Members: []ggql.Type{a, b},
	}
	err := root.AddTypes(u, a, b, &dirt)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	actual := root.SDL(false, true)
	expectStr := `union Uu @dirt = Aa | Bb
`
	expectRoot := `
type Aa {
  a: String
}

type Bb {
  b: String
}

"Union of A and B."
` + expectStr + timeScalarSDL + `
directive @dirt on UNION
`

	checkEqual(t, expectRoot, actual, "Union SDL() mismatch")
	checkEqual(t, expectStr, u.String(), "Union String() mismatch")
	checkEqual(t, 7, u.Rank(), "Union Rank() mismatch")
	checkEqual(t, "Union of A and B.", u.Description(), "Union Description() mismatch")
	checkEqual(t, 1, len(u.Directives()), "Union Directives() mismatch")

	err = u.Extend(&ggql.Union{Base: ggql.Base{N: "Uu"}, Members: []ggql.Type{&ggql.Object{Base: ggql.Base{N: "Cc"}}}})
	checkNil(t, err, "union.Extend failed. %s", err)

	err = u.Extend(&ggql.Union{Base: ggql.Base{N: "Uu"}, Members: []ggql.Type{&ggql.Object{Base: ggql.Base{N: "Cc"}}}})
	checkNotNil(t, err, "duplicate member for union.Extend should have failed.")

	w := &failWriter{max: 8}
	err = u.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	w = &failWriter{max: 10}
	err = u.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	w = &failWriter{max: 18}
	err = u.Write(w, false)
	checkNotNil(t, err, "return error on write error")

}
