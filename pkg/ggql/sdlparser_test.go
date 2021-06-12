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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestRootParseScalar(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
scalar Lizard # lizards are scaly
`
	expect := `
scalar Lizard

scalar Time
`
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	out := root.SDL(false)
	checkEqual(t, expect, out, "parsed vs given")

	err = root.ParseString("scalar ")
	checkNotNil(t, err, "an error should be returned when parsing an invalid valid SDL.")

	err = root.ParseString("scalar Xyz @example(")
	checkNotNil(t, err, "an error should be returned when parsing an invalid valid SDL.")
}

func TestRootParseDirective(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
directive @example("a number"num: Int = 7) on OBJECT | FIELD`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := timeScalarSDL + `
directive @example(
    "a number"
    num: Int = 7) on OBJECT | FIELD
`
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	err = root.Parse([]byte("directive @short on FIELD | "))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	for _, s := range []string{
		"directive ",
		"directive example",
		"directive @ |",
		"directive @example no ",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid directive.")
	}
}

func TestRootParseEnum(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
enum Dir @example {IN OUT @example}
directive @example on ENUM|ENUM_VALUE|FIELD_DEFINITION
`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
enum Dir @example {
  IN
  OUT @example
}
` + timeScalarSDL + `
directive @example on ENUM | ENUM_VALUE | FIELD_DEFINITION
`
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	for _, s := range []string{
		"enum {",
		"enum Dir $",
		"enum Dir { IN ",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid directive.")
	}
}

func TestRootParseInput(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
input Inn @example {a: Int b: Boolean @example c: Int = 3}
directive @example(x: Int = 3) on INPUT_OBJECT
`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
input Inn @example {
  a: Int
  b: Boolean @example
  c: Int = 3
}
` + timeScalarSDL + `
directive @example(x: Int = 3) on INPUT_OBJECT
`
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	for _, s := range []string{
		"input {",
		"input Inn $",
		"input Inn { ",
		"input Inn { a 1",
		"input Inn {a ",
		"input Inn {a: ",
		"input Inn {a: Int @foo(",
		"input Inn {a: Int @foo(a: ",
		"input Inn {a: Int @foo(a: Int, a: Int) ",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid input.")
	}
}

func TestRootParseInterface(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
interface Inn @example {a: Int b: Boolean @example}
directive @example on INTERFACE | FIELD_DEFINITION
`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
interface Inn @example {
  a: Int
  b: Boolean @example
}
` + timeScalarSDL + `
directive @example on INTERFACE | FIELD_DEFINITION
`
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	for _, s := range []string{
		"interface {",
		"interface Inn $",
		"interface Inn {a: ",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid directive.")
	}
}

func TestRootParseSchema(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
type Query { a: Int }
schema @go(type: "MySchema"){query: Query}
`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
schema @go(type: "MySchema") {
  query: Query
}

type Query {
  a: Int
}
` + timeScalarSDL
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	for _, s := range []string{
		"schema {",
		"schema $",
		"schema {a: ",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid directive.")
	}
}

func TestRootParseObject(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
interface Aa {a:Int}
interface Bb {b:Boolean}
type Tie implements & Aa & Bb @go(type: "MyType"){a: Int b: Boolean,
  c: [Int!]
  d(x: Int!, y: Int = 7): String
}
`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
type Tie implements Aa & Bb @go(type: "MyType") {
  a: Int
  b: Boolean
  c: [Int!]
  d(x: Int!, y: Int = 7): String
}

interface Aa {
  a: Int
}

interface Bb {
  b: Boolean
}
` + timeScalarSDL
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	for _, s := range []string{
		"type {",
		"type Tie $",
		"type Tie {a: ",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid directive.")
	}
}

func TestRootParseUnion(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
directive @example on UNION
type Aaa { a: Int }
type Bbb { b: Int }
union Uuu @example=Aaa|Bbb
`
	err := root.Parse([]byte(sdl))
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
type Aaa {
  a: Int
}

type Bbb {
  b: Int
}

union Uuu @example = Aaa | Bbb
` + timeScalarSDL + `
directive @example on UNION
`
	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")

	for _, s := range []string{
		"union @",
		"union Uuu $",
	} {
		err = root.ParseString(s)
		checkNotNil(t, err, "an error should be returned when parsing an invalid directive.")
	}
}

func TestRootParseDirUse(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
enum Slick { SLIME SLIDE }
directive @example(s: Slick) on SCALAR
type Lizard @go(type: "MyType") { a: Int }
scalar Fish @example(s: SLIME)
`
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	expect := `
type Lizard @go(type: "MyType") {
  a: Int
}

enum Slick {
  SLIME
  SLIDE
}

scalar Fish @example(s: SLIME)

scalar Time

directive @example(s: Slick) on SCALAR
`
	out := root.SDL(false)
	checkEqual(t, expect, out, "parsed vs given")
}

func TestRootParseDesc(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
"This is a one liner."
scalar Lizard
`
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	out := root.SDL(false, true)
	checkEqual(t, sdl+timeScalarSDL, out, "parsed vs given")

	root = ggql.NewRoot(nil)
	sdl = `
"""
This is a multi line
description.
"""
scalar Fish
` + timeScalarSDL
	err = root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	out = root.SDL(false, true)
	checkEqual(t, sdl, out, "parsed vs given")

	root = ggql.NewRoot(nil)
	sdl = `
type Lizard {

  """
  This is an indented
  multiline description.
  """
  size: Int
}
`
	err = root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	out = root.SDL(false, true)
	checkEqual(t, sdl+timeScalarSDL, out, "parsed vs given")

	err = root.ParseString(`""" not terminated`)
	checkNotNil(t, err, "an error should be returned when parsing a non terminated string. %s")
}

func TestRootParseMultipleDesc(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
"Lizards are cold"
scalar Lizard

scalar Worm
`
	expect := `
"Lizards are cold"
scalar Lizard
` + timeScalarSDL + `
scalar Worm
`
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	out := root.SDL(false, true)
	checkEqual(t, expect, out, "parsed vs given")
}

func TestRootParseBOM(t *testing.T) {
	root := ggql.NewRoot(nil)

	err := root.ParseString("\xef\xbb\xbfscalar Lizard")
	checkNil(t, err, "no error should be returned when parsing an SDL with a BOM. %s", err)

	err = root.ParseString("\xef\xba\xbfscalar Fish")
	checkNotNil(t, err, "an error should be returned when parsing an SDL with a bad BOM. %s", err)
}

func TestRootParseBad(t *testing.T) {
	root := ggql.NewRoot(nil)
	for _, sdl := range []string{
		"nonsense",
		"nonsense ",
		"scalar Lizard @ ",
		"scalar Lizard @example(a 1)",
		"scalar Lizard @example(:1)",
		"directive @example(a: Int",
		"directive @example(a Int)",
		"directive @example(a: [Int)",
		"directive @example(a: )",
		"type Ooo { : 1 }",
		"type Ooo { a 1 }",
		"type Ooo { a: 1",
		"type Ooo itchy {a:1}",
		"type Ooo implements {a:1}",
		"type Ooo {a(x:Int, x:Int)}",
		"type Ooo {a:Int,a:Int}",
	} {
		err := root.ParseString(sdl)
		checkNotNil(t, err, "an error should be returned when parsing '%s'.", sdl)
	}
}

func failOnRead(t *testing.T, sdl string, pos int) {
	root := ggql.NewRoot(nil)
	r := failReader{max: pos, content: []byte(sdl)}
	err := root.ParseReader(&r)
	checkNotNil(t, err, "an error should be returned when parsing '%s' with a read error at %d.", sdl, pos)
}

func TestRootParseFail(t *testing.T) {
	sdl := "  scalar Lizard @example  (  abc : 1234) # comment"
	for i := 0; i < len(sdl); i++ {
		failOnRead(t, sdl, i)
	}
	sdl = `directive @example ( "hi" abc  :  [Int!]!  =  3 ) on OBJECT | FIELD`
	for i := 0; i < len(sdl); i++ {
		failOnRead(t, sdl, i)
	}
	sdl = `type Ooo { a  : Int = 33 }`
	for i := 0; i < len(sdl); i++ {
		failOnRead(t, sdl, i)
	}
	failOnRead(t, "enum Dir {IN OUT}", 12)
	failOnRead(t, "union Uuu =   Aaa | Bbb", 12)
}

func TestRootParseExtend(t *testing.T) {
	root := ggql.NewRoot(nil)
	sdl := `
directive @example on SCALAR
scalar Lizard
extend scalar Lizard @example
`
	expect := `
scalar Lizard @example

scalar Time

directive @example on SCALAR
`
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	out := root.SDL(false)
	checkEqual(t, expect, out, "parsed vs given")

	err = root.ParseString("extend scalar Lizard @example")
	checkNotNil(t, err, "an error should be returned when extending an existing directive.")
}

func TestRootParseExtendBad(t *testing.T) {
	for _, sdl := range []string{
		"extend scalar Lizard @example directive @example on SCALAR",
		"extend scalar Lizard @example directive @example on SCALAR enum Lizard{IN,OUT}",
		"extend scalar Lizard @example scalar Lizard",
	} {
		root := ggql.NewRoot(nil)
		err := root.ParseString(sdl)
		checkNotNil(t, err, "an error should be returned when parsing '%s'.", sdl)
	}
}

func TestRootParseFS(t *testing.T) {
	tempdir := t.TempDir()
	err := os.WriteFile(filepath.Join(tempdir, "other.txt"), []byte("scalar Bad"), 0600) // should not be parsed.
	checkNil(t, err, "no error should be returned")
	err = os.WriteFile(filepath.Join(tempdir, "email.graphql"), []byte("scalar Email"), 0600)
	checkNil(t, err, "no error should be returned")
	err = os.WriteFile(filepath.Join(tempdir, "phone.gql"), []byte("scalar Phone"), 0600)
	checkNil(t, err, "no error should be returned")

	root := ggql.NewRoot(nil)
	err = root.ParseFS(os.DirFS(tempdir), "*.graphql", "*.gql")
	checkNil(t, err, "no error should be returned.")

	email := root.GetType("Email")
	checkNotNil(t, email, "email type should not be nil")
	phone := root.GetType("Phone")
	checkNotNil(t, phone, "phone type should not be nil")
	bad := root.GetType("Bad")
	checkNil(t, bad, "bad type should not have been parsed")
}

func TestRootParseFSBadPattern(t *testing.T) {
	root := ggql.NewRoot(nil)
	err := root.ParseFS(os.DirFS(t.TempDir()), "\\")
	checkNotNil(t, err, "an error should be returned")
}

func TestRootParseFSErr(t *testing.T) {
	tempdir := t.TempDir()
	const noPermissions = 0000
	err := os.WriteFile(filepath.Join(tempdir, "bad.graphql"), []byte("scalar Bad"), noPermissions)
	checkNil(t, err, "no error should be returned writing the file")

	root := ggql.NewRoot(nil)
	err = root.ParseFS(os.DirFS(tempdir), "*.graphql")
	checkNotNil(t, err, "an error should be returned")
	checkEqual(t, true, strings.Contains(err.Error(), "permission denied"), "error states file could not be opened")
}

func TestRootParseFSNoPatterns(t *testing.T) {
	root := ggql.NewRoot(nil)
	err := root.ParseFS(os.DirFS(t.TempDir()))
	checkNil(t, err, "no error should be returned")
}
