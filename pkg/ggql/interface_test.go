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

func TestInterface(t *testing.T) {
	dirt := ggql.Directive{
		Base: ggql.Base{
			N: "dirt",
		},
		On: []ggql.Location{ggql.LocArgumentDefinition},
	}
	root := ggql.NewRoot(nil)
	inty := ggql.Interface{
		Base: ggql.Base{
			N:    "Inty",
			Desc: "Some interface.",
		},
	}
	fd := ggql.FieldDef{
		Base: ggql.Base{
			N: "fun",
		},
		Type: root.GetType("String"),
	}
	fd.AddArg(&ggql.Arg{
		Base: ggql.Base{
			N: "one",
			Dirs: []*ggql.DirectiveUse{
				{
					Directive: &dirt,
				},
			},
		},
		Type: root.GetType("String"),
	})
	fd.AddArg(&ggql.Arg{
		Base: ggql.Base{
			N:    "two",
			Desc: "Second argument",
		},
		Type: root.GetType("String"),
	})
	inty.AddField(&fd)

	inty.AddField(&ggql.FieldDef{
		Base: ggql.Base{
			N: "feel",
		},
		Type: root.GetType("String"),
	})

	err := root.AddTypes(&inty)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	actual := root.SDL(false, true)
	expectStr := `interface Inty {
  fun(one: String @dirt, two: String): String
  feel: String
}
`
	expectRoot := `
"Some interface."
interface Inty {
  fun(one: String @dirt,
    "Second argument"
    two: String): String
  feel: String
}
` + timeScalarSDL

	checkEqual(t, expectRoot, actual, "Interface SDL() mismatch")
	checkEqual(t, expectStr, inty.String(), "Interface String() mismatch")
	checkEqual(t, 8, inty.Rank(), "Interface Rank() mismatch")
	checkEqual(t, "Some interface.", inty.Description(), "Interface Description() mismatch")
	checkEqual(t, 0, len(inty.Directives()), "Interface Directives() mismatch")

	// Exercise the write failure in args.
	w := &failWriter{max: 55}
	err = inty.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	w = &failWriter{max: 56}
	err = inty.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	w = &failWriter{max: 50}
	err = inty.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	x := &ggql.Interface{Base: ggql.Base{N: "Inty"}, Root: root}
	x.AddField(&ggql.FieldDef{Base: ggql.Base{N: "c"}, Type: root.GetType("Int")})
	err = inty.Extend(x)
	checkNil(t, err, "extend should not return an error. %s", err)

	x = &ggql.Interface{Base: ggql.Base{N: "Inty"}}
	x.AddField(&ggql.FieldDef{Base: ggql.Base{N: "c"}, Type: root.GetType("Int")})
	err = inty.Extend(x)

	checkNotNil(t, err, "duplicate field for interface.Extend should have failed.")
}
