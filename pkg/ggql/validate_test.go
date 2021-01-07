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
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

type valTestData struct {
	sdl    string
	expect string
}

func valTest(t *testing.T, data []*valTestData, debug bool) {
	root := ggql.NewRoot(nil)
	for _, d := range data {
		err := root.ParseString(d.sdl)
		if debug {
			fmt.Printf("*** %s\n", err)
		}
		checkValErr(t, err, d.sdl, d.expect)
	}
}

func checkValErr(t *testing.T, err error, where string, expect ...string) {
	var errs ggql.Errors
	if !errors.As(err, &errs) {
		t.Fatalf("expected a ggql.Errors error, not a %T (%s)at %s", err, err, where)
	}
	for _, x := range expect {
		ok := false
		for _, e := range errs {
			if strings.Contains(e.Error(), x) {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("%s: did not find error matching '%s' in\n%s", where, x, err)
		}
	}
}

func TestRootValidateName(t *testing.T) {
	// Exists only to stop lint from complaining about the debug argument
	// always being false.
	valTest(t, []*valTestData{}, true)

	valTest(t, []*valTestData{
		{sdl: "scalar 1Scale", expect: "1Scale is not a valid type name, it begins with a number at 1:8"},
		{sdl: "scalar __Scale", expect: "__Scale is not a valid type name, it begins with '__' at 1:8"},
	}, false)
	root := ggql.NewRoot(nil)

	err := root.AddTypes(&ggql.Schema{Object: ggql.Object{Base: ggql.Base{N: "Name"}}})
	checkValErr(t, err, "schema with name", "a schema can not have a name")

	err = root.AddTypes(&ggql.Object{})
	checkValErr(t, err, "horse with no name", "a type name can not be blank", " must have at least one field")

	err = root.AddTypes(&ggql.Object{Base: ggql.Base{N: "My-Name"}})
	checkValErr(t, err, "invalid name", "My-Name is not a valid type name", "My-Name must have at least one field")
}

func TestRootValidateDirUse(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: `
directive @example(a: Int) on SCALAR
scalar Scale @example(b: 3)
`, expect: "directive argument b for directive example on Scale not found at 3:23"},
		{sdl: `
directive @example on OBJECT
scalar Scale @example
`, expect: "directive @example can not be applied to Scale, a SCALAR at 3:14"},
		{sdl: `
directive @example(a: Int) on SCALAR
scalar Scale @example(a: "Not a number")
`, expect: "can not coerce a string into a Int at 3:23"},
	}, false)

	root := ggql.NewRoot(nil)

	err := root.ParseString(`
directive @example(a: ID) on SCALAR
scalar Scale @example(a: 123)
`)
	checkNil(t, err, "validation ID coerce from int should not fail. %s", err)

	err = root.ParseString(`
enum Slick { SLIME SLIDE }
directive @example(a: Slick) on SCALAR
scalar Scale @example(a: "SLIME")
`)
	checkNotNil(t, err, "validation union value coerce from string should fail.")

	dirt := ggql.Enum{Base: ggql.Base{N: "dirt"}}
	err = root.AddTypes(
		&ggql.Object{Base: ggql.Base{N: "Obj", Dirs: []*ggql.DirectiveUse{{Directive: &dirt}}}},
		&dirt,
	)
	checkValErr(t, err, "directive use with non-directive", "invalid directive at Obj a OBJECT")
}

func TestRootValidateUnion(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: `
enum Aa { ROCK }
type Bb { b: Int}
union Uu = Aa | Bb
`, expect: "Aa can not be a union member since it is a *ggql.Enum, not an *ggql.Object at 4:7"},
		{sdl: "union Uu = ", expect: "union Uu must have at least one member at 1:7"},
	}, false)
}

func TestRootValidateEnum(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: "enum Ee { }", expect: "enum Ee must have at least one value at 1:6"},
		{sdl: "enum Ee { true }", expect: "true is not a valid enum value for enum Ee at 1:11"},
	}, false)
}

func TestRootValidateDirective(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: "directive @example on NONSENSE", expect: "NONSENSE is not a valid location at 1:11"},
		{sdl: `
directive @example(a: Obj) on SCALAR
scalar Scale @example(a: 4)
type Obj { a: Int }`, expect: "directive example argument a, a *ggql.Object is not an input type at 2:20"},
		{sdl: `
directive @one(a: Int @one) on ARGUMENT_DEFINITION
`, expect: "directive one has a directive loop - one.a->one at 2:11"},
		{sdl: `
directive @one(a: Int @two) on ARGUMENT_DEFINITION
directive @two(a: Int @one) on ARGUMENT_DEFINITION
`, expect: "directive one has a directive loop - one.a->two.a->one at 2:11"},
		{sdl: `directive @example(a: Int = ABC) on SCALAR`, expect: "can not coerce a ggql.Symbol into a Int at 1:20"},
	}, false)
}

func TestRootValidateInterface(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: "interface Inty { }", expect: "Inty must have at least one field at 1:11"},
		{sdl: "interface Inty { __a: Int }", expect: "__a is not a valid field name, it begins with '__' at 1:18"},
		{
			sdl:    "interface Inty { a(__x: Int): Int }",
			expect: "__x is not a valid argument name, it begins with '__' at 1:20",
		},
		{
			sdl:    "input Inn { x: Int } interface Inty { a: Inn }",
			expect: "a does not return an output type at 1:39",
		},
		{
			sdl:    "type Obj { x: Int } interface Inty { a(x: Obj): Int }",
			expect: "argument x of a must be an input type at 1:40",
		},
	}, false)
}

func TestRootValidateObject(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: "type Obj { }", expect: "Obj must have at least one field at 1:6"},
		{sdl: "type Obj { __a: Int }", expect: "__a is not a valid field name, it begins with '__' at 1:12"},
		{
			sdl:    "type Obj { a(__x: Int): Int }",
			expect: "__x is not a valid argument name, it begins with '__' at 1:14",
		},
		{sdl: "input Inn { x: Int } type Obj { a: Inn }", expect: "a does not return an output type at 1:33"},
		{sdl: "type Obj { a(x: Obj): Int }", expect: "argument x of a must be an input type at 1:14"},
		{
			sdl:    `input Imp { a: Int } type Obj implements Imp { a: Int }`,
			expect: "Imp is not an interface for Obj at 1:27",
		},
		{
			sdl:    `interface Imp {a: Int} type Obj implements Imp {b: Int}`,
			expect: "Obj is missing field a from interface Imp at 1:29",
		},
		{
			sdl:    `interface Imp {a: Int} type Obj implements Imp {a: Float}`,
			expect: "field a return type Float is not a sub-type of Int at 1:49",
		},
		{
			sdl:    `interface Imp {a(x: Int): Int} type Obj implements Imp {a: Int}`,
			expect: "argument a to x missing at 1:57",
		},
		{
			sdl:    `interface Imp {a(x: Int): Int} type Obj implements Imp {a(y: Int): Int}`,
			expect: "argument a to x missing at 1:57",
		},
		{
			sdl:    `interface Imp {a(x: Int): Int} type Obj implements Imp {a(x: Float): Int}`,
			expect: "argument return for x does not match interface at 1:59",
		},
		{
			sdl:    `interface Imp {a: Int} type Obj implements Imp {a(y: Int!): Int}`,
			expect: "additional argument y to interface field must be optional at 1:51",
		},
		{
			sdl:    `interface Imp {a: Float!} type Obj implements Imp {a: Int!}`,
			expect: "field a return type Int! is not a sub-type of Float! at 1:52",
		},
		{
			sdl:    `interface Imp { a: Imp, b: Int } type Obj implements Imp {a: Obj}`,
			expect: "Obj is missing field b from interface Imp at 1:39",
		},
		{sdl: `type Aa {a: Int}
type Bb {a: Int}
union Uu = Aa | Bb
interface Imp { a: Uu, b: Int }
type Obj implements Imp {a: Aa}
`, expect: "Obj is missing field b from interface Imp at 5:6"},
	}, false)
}

func TestRootValidateInput(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: "input Inn { }", expect: "input object Inn must have at least one field at 1:7"},
		{sdl: "input Inn { __a: Int }", expect: "__a is not a valid field name, it begins with '__' at 1:13"},
		{sdl: "type Obj { x: Int } input Inn { a: Obj }", expect: "a does not return an input type at 1:33"},
	}, false)
}

func TestRootValidateSchema(t *testing.T) {
	valTest(t, []*valTestData{
		{sdl: "schema { bad: Int }", expect: "field bad on Schema not allowed at 1:10"},
	}, false)
}

func TestRootValidateNames(t *testing.T) {
	samples := []string{
		"type %s { id: ID }",
		"type X { %s: ID }",
		"type X { y(%s: String): String }",
		"interface %s { id: ID }",
		"interface X { %s: ID }",
		"type X { id: ID } union %s = X",
		"scalar %s",
		"enum %s { FOO }",
		"enum SomeEnum { %s }",
		"input %s { id: ID }",
		"input X { %s: ID }",
		"directive @%s on SCALAR",
		"directive @X(%s: Int) on SCALAR",
	}
	// name -> valid?
	names := map[string]bool{
		"foo":    true,
		"f123":   true,
		"_xyz":   true,
		"_00abc": true,
		"0abc":   false,
		"%":      false,
		"'":      false,
		"/":      false,
		"__abc":  false,
	}
	for _, s := range samples {
		for n, isValid := range names {
			root := ggql.NewRoot(nil)
			sdl := fmt.Sprintf(s, n)
			err := root.ParseString(sdl)
			if isValid {
				checkNil(t, err, "expected no error but got err %q for sdl: %s\n", err, sdl)
			} else {
				checkNotNil(t, err, "expected error but got none for sdl: %s\n", sdl)
			}
		}
	}
}
