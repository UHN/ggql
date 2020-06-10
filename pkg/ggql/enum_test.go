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

func TestEnum(t *testing.T) {
	root := ggql.NewRoot(nil)

	dir := ggql.Directive{
		Base: ggql.Base{
			N: "ref",
		},
		On: []ggql.Location{ggql.LocEnumValue},
	}
	dir.AddArg(&ggql.Arg{Base: ggql.Base{N: "a"}, Type: root.GetType("String")})
	dir.AddArg(&ggql.Arg{Base: ggql.Base{N: "b"}, Type: root.GetType("String")})

	enum := &ggql.Enum{
		Base: ggql.Base{
			N: "Genre",
		},
	}
	enum.AddValue(&ggql.EnumValue{Value: "ROCK"})
	enum.AddValue(&ggql.EnumValue{
		Value:       "CLASSICAL",
		Description: "Short description.",
		Directives: []*ggql.DirectiveUse{
			{
				Directive: &dir,
				Args: map[string]*ggql.ArgValue{
					"a": {
						Arg:   "a",
						Value: "aaa",
					},
					"b": {
						Arg:   "b",
						Value: "bbb",
					},
				},
			},
		},
	})
	enum.AddValue(&ggql.EnumValue{
		Value:       "JAZZ",
		Description: "A multi-line\ndescription.",
	})

	err := root.AddTypes(enum)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	actual := root.SDL(false, true)
	expectRoot := `
enum Genre {
  ROCK

  "Short description."
  CLASSICAL @ref(a: "aaa", b: "bbb")

  """
  A multi-line
  description.
  """
  JAZZ
}
` + timeScalarSDL
	expectStr := `enum Genre {
  ROCK
  CLASSICAL @ref(a: "aaa", b: "bbb")
  JAZZ
}
`
	checkEqual(t, expectRoot, actual, "Enum SDL() mismatch")
	checkEqual(t, expectStr, enum.String(), "Enum String() mismatch")
	checkEqual(t, 9, enum.Rank(), "Enum Rank() mismatch")

	x := &ggql.Enum{Base: ggql.Base{N: "Genre"}}
	x.AddValue(&ggql.EnumValue{Value: "POP"})
	err = enum.Extend(x)
	checkNil(t, err, "enum.Extend failed. %s", err)

	x = &ggql.Enum{Base: ggql.Base{N: "Genre"}}
	x.AddValue(&ggql.EnumValue{Value: "POP"})
	err = enum.Extend(x)
	checkNotNil(t, err, "duplicate value for enum.Extend should have failed.")

	w := &failWriter{max: 30}
	err = enum.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	w = &failWriter{max: 45}
	err = enum.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	w = &failWriter{max: 47}
	err = enum.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	// Cause writeHeader desc to fail.
	w = &failWriter{max: 20}
	err = enum.Write(w, true)
	checkNotNil(t, err, "return error on write error")

	_, err = enum.CoerceIn(3)
	checkNotNil(t, err, "return error on Enum.CoerceIn(3)")

	var r interface{}
	r, err = enum.CoerceIn(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)

	var v interface{}
	v, err = enum.CoerceOut(ggql.Symbol("abc"))
	checkNil(t, err, "Enum.CoerceOut(abc) should not return an error. %s", err)
	checkEqual(t, "abc", v, "coerce mismatch")

	v, err = enum.CoerceOut("abc")
	checkNil(t, err, "Enum.CoerceOut(abc) should not return an error. %s", err)
	checkEqual(t, "abc", v, "coerce mismatch")

	_, err = enum.CoerceOut(3)
	checkNotNil(t, err, "return error on Enum.CoerceOut(3)")

	r, err = enum.CoerceOut(nil)
	checkNil(t, r, "CoerceOut value mismatch")
	checkNil(t, err, "CoerceOut error. %s", err)
}

func TestEnumBadRef(t *testing.T) {
	root := ggql.NewRoot(nil)

	enum := &ggql.Enum{Base: ggql.Base{N: "Genre"}}
	enum.AddValue(&ggql.EnumValue{Value: "ROCK", Directives: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}}})
	err := root.AddTypes(enum)
	checkNotNil(t, err, "root.AddTypes should fail with bad directive ref.")
}

func TestEnumValues(t *testing.T) {
	enum := ggql.Enum{}
	checkEqual(t, 0, len(enum.Values()), "Enum values should be empty")
}
