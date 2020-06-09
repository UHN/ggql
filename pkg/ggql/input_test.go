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
	"bytes"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestInput(t *testing.T) {
	ggql.Sort = true
	root := ggql.NewRoot(nil)
	imp := ggql.Input{
		Base: ggql.Base{
			N:    "Imp",
			Desc: "Some input.",
		},
	}
	imp.AddField(&ggql.InputField{
		Base: ggql.Base{
			N: "a",
		},
		Type: &ggql.List{Base: root.GetType("String")},
	})
	imp.AddField(&ggql.InputField{
		Base: ggql.Base{
			N: "b",
			Dirs: []*ggql.DirectiveUse{
				{
					Directive: &ggql.Directive{Base: ggql.Base{N: "example"},
						On: []ggql.Location{ggql.LocField},
					},
				},
			},
		},
		Type:    root.GetType("String"),
		Default: "bee",
	})
	err := root.AddTypes(&imp)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	actual := root.SDL(false, true)
	expectStr := `input Imp {
  a: [String]
  b: String = "bee" @example
}
`
	expectRoot := `
"Some input."
` + expectStr + timeScalarSDL

	checkEqual(t, expectRoot, actual, "Input SDL() mismatch")
	checkEqual(t, expectStr, imp.String(), "Input String() mismatch")
	checkEqual(t, 6, imp.Rank(), "Input Rank() mismatch")
	checkEqual(t, "Some input.", imp.Description(), "Input Description() mismatch")
	checkEqual(t, 0, len(imp.Directives()), "Input Directives() mismatch")

	x := &ggql.Input{Base: ggql.Base{N: "Imp"}}
	x.AddField(&ggql.InputField{Base: ggql.Base{N: "c"}, Type: root.GetType("Int")})
	err = imp.Extend(x)
	checkNil(t, err, "input.Extend failed. %s", err)

	x = &ggql.Input{Base: ggql.Base{N: "Imp"}}
	x.AddField(&ggql.InputField{Base: ggql.Base{N: "c"}, Type: root.GetType("Int")})
	err = imp.Extend(x)
	checkNotNil(t, err, "duplicate field for input.Extend should have failed.")

	for i := 0; i < len(expectStr); i++ {
		w := &failWriter{max: i}
		err = imp.Write(w, false)
		checkNotNil(t, err, "return error on write error")
	}
	var out interface{}
	out, err = imp.CoerceIn(map[string]interface{}{"a": []interface{}{"abc"}, "c": 7})

	checkNil(t, err, "Input.CoerceIn({a:[abc],b:def}) should not fail. %s", err)
	var b bytes.Buffer
	_ = ggql.WriteSDLValue(&b, out, 0)
	checkEqual(t, `{a: ["abc"], b: "bee", c: 7}`, b.String(), "output mismatch after Input.CoerceIn")

	_, err = imp.CoerceIn(false)
	checkNotNil(t, err, "can not coerce a scalar into an input")

	_, err = imp.CoerceIn(map[string]interface{}{"d": 7})
	checkNotNil(t, err, "can not coerce a non field in an input")

	_, err = imp.CoerceIn(map[string]interface{}{"a": []interface{}{7}})
	checkNotNil(t, err, "can not coerce a mismatched list field in an input")

	x = &ggql.Input{Base: ggql.Base{N: "Imp"}}
	x.AddField(&ggql.InputField{Base: ggql.Base{N: "x"}, Type: &ggql.Union{Base: ggql.Base{N: "Uu"}}})
	err = imp.Extend(x)
	checkNil(t, err, "input.Extend failed. %s", err)

	_, err = imp.CoerceIn(map[string]interface{}{"x": 7})
	checkNotNil(t, err, "can not coerce a non-coercable type in an input")
}

func TestInputFields(t *testing.T) {
	input := ggql.Input{}
	checkEqual(t, 0, len(input.Fields()), "Input Fields should be empty")
}
