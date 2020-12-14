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
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

const songSdl = `
type Query {
  title: String
  artist(name: String!): Artist
  artists: [Artist]
}

type Artist {
  name: String!
  songs: [Song]
  origin: [String]
}

type Song {
  name: String!
  artist: Artist
  duration: Int
  release: Time
}

directive @example on VARIABLE_DEFINITION
`

func setupSchema(t *testing.T) *ggql.Root {
	root := ggql.NewRoot(nil)

	err := root.ParseString(songSdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	return root
}

func TestRootParseExecutable(t *testing.T) {
	root := setupSchema(t)

	for src, expect := range map[string]string{
		"{ title }": `query {
  title
}
`,
		"\xef\xbb\xbf{ title }": `query {
  title
}
`,
		"query getTitle { title }": `query getTitle {
  title
}
`,
		"{ alien: title }": `query {
  alien: title
}
`,
		"{artists{name}}": `query {
  artists {
    name
  }
}
`,
		`{artist(name:"Frazerdaze"){name}}`: `query {
  artist(name: "Frazerdaze") {
    name
  }
}
`,
		`query artistName($id: String = "Who"){artist(name:$id){name}}`: `query artistName($id: String = "Who") {
  artist(name: $id) {
    name
  }
}
`,
		`query artistName($id:String="Who"@example){artist(name:$id){name}}`: `query artistName($id: String = "Who" @example) {
  artist(name: $id) {
    name
  }
}
`,
		`{artists{name, ...{songs}}}`: `query {
  artists {
    name
    ... {
      songs
    }
  }
}
`,
		`{artists{name, ...on Artist{songs}}}`: `query {
  artists {
    name
    ... on Artist {
      songs
    }
  }
}
`,
		`{artists{name, ...sung}} fragment sung on Artist {songs}`: `query {
  artists {
    name
    ...sung
  }
}

fragment sung on Artist {
  songs
}
`,
		`fragment sung on Artist {songs} {artists{name, ...sung}}`: `query {
  artists {
    name
    ...sung
  }
}

fragment sung on Artist {
  songs
}
`,
		`{artists{name, ...sung}} fragment sung on Artist {songs} fragment where on Artist {origin}`: `query {
  artists {
    name
    ...sung
  }
}

fragment sung on Artist {
  songs
}

fragment where on Artist {
  origin
}
`,
		`{artists{name, ...sung ...where}} fragment sung on Artist {songs} fragment where on Artist {origin}`: `query {
  artists {
    name
    ...sung
    ...where
  }
}

fragment sung on Artist {
  songs
}

fragment where on Artist {
  origin
}
`,
	} {
		exe, err := root.ParseExecutableString(src)
		checkNil(t, err, "ParseExecutableString(%s) failed. %s", src, err)
		checkEqual(t, expect, exe.String(), "result mismatch for %s", src)
	}
}

func TestRootParseExecutableBytes(t *testing.T) {
	root := setupSchema(t)

	exe, err := root.ParseExecutable([]byte("{title}"))
	checkNil(t, err, "ParseExecutable({title}) failed. %s", err)
	checkEqual(t, `query {
  title
}
`, exe.String(), "result mismatch")
}

type parseTestData struct {
	src  string
	line int
	col  int
}

func TestParseExecutableError(t *testing.T) {
	root := ggql.NewRoot(nil)
	for _, d := range []parseTestData{
		{src: "query dup { title } query dup { origin }", line: 1, col: 28},
		{src: `
{ title }
fragment dup on Artist { title }
fragment dup on Artist { origin }`, line: 4, col: 11},
		{src: "{ title } { artists }", line: 1, col: 12},
		{src: "bogus { title }", line: 1, col: 2},
		{src: "{ (a: Int) }", line: 1, col: 4},
		{src: "{ ..bad } fragment bad on Query { title }", line: 1, col: 6},
		{src: "{ ... on Bad { a }}", line: 1, col: 10},
		{src: "query Qq($a: Int = 3", line: 1, col: 21},
		{src: "query Qq(a: Int = 3) { title }", line: 1, col: 11},
		{src: "query Qq($: Int = 3) { title }", line: 1, col: 12},
		{src: "query Qq($a Int = 3) { title }", line: 1, col: 14},
		{src: `query Qq($a: Int = "x") { artist(name: $a`, line: 1, col: 42},
		{src: `{artists{songs:{name}}}`, line: 1, col: 11},
	} {
		_, err := root.ParseExecutableString(d.src)
		checkNotNil(t, err, "ParseExecutableString(%s) should fail.", d.src)
		var ge *ggql.Error
		var ges ggql.Errors
		switch {
		case errors.As(err, &ge):
			checkEqual(t, d.line, ge.Line, "line number mismatch for %s. %s", d.src, ge)
			checkEqual(t, d.col, ge.Column, "column number mismatch for %s. %s", d.src, ge)
		case errors.As(err, &ges):
			checkEqual(t, 1, len(ges), "ParseExecutableString(%s) should return one error. %s", d.src, err)
			var e2 *ggql.Error
			errors.As(ges[0], &e2)
			checkNotNil(t, e2, "ParseExecutableString(%s) should return a ggql.Errors with one ggql.Error or not a %T. %s",
				d.src, ges[0], ges[0])
			checkEqual(t, d.line, e2.Line, "line number mismatch for %s. %s", d.src, e2)
			checkEqual(t, d.col, e2.Column, "column number mismatch for %s. %s", d.src, e2)
		default:
			t.Fatalf("\nParseExecutableString(%s) should return a *ggql.Error or ggql.Errors not a %T. %s",
				d.src, err, err)
		}
	}
}

func TestRootParseExecutableEmptyFieldName(t *testing.T) {
	root := ggql.NewRoot(nil)
	src := `{artists{songs:{name}}}`
	_, err := root.ParseExecutableString(src)
	checkNotNil(t, err, "ParseExecutableString(%s) should fail.", src)
}

func TestRootParseExecutableFail(t *testing.T) {
	for _, src := range []string{
		"{ title }",
		"\xef\xbb\xbf{ title }",
		"{artists{name, ...sung}} fragment sung on Artist {songs}",
		"query Qq ( $a : Int! = 3 ) { title }",
		`query Qq ($a: String = "x") { artist (name: $id ) }`,
		`{ arty: artists (name: "x") { name }}`,
	} {
		for i := 0; i < len(src); i++ {
			failExeRead(t, src, i)
		}
	}
}

func TestRootParseExecutableNoHang(t *testing.T) {
	root := ggql.NewRoot(nil)
	src := "{artist{name}}}"
	_, err := root.ParseExecutableString(src)
	checkNotNil(t, err, "ParseExecutableString(%s) should fail", src)
}

func failExeRead(t *testing.T, src string, pos int) {
	root := ggql.NewRoot(nil)
	r := failReader{max: pos, content: []byte(src)}
	_, err := root.ParseExecutableReader(&r)
	checkNotNil(t, err, "an error should be returned when parsing '%s' with a read error at %d.", src, pos)
}

func valExeTest(t *testing.T, data []*valTestData, debug bool) {
	root := setupSchema(t)
	for _, d := range data {
		_, err := root.ParseExecutableString(d.sdl)
		if debug {
			fmt.Printf("*** %s\n", err)
		}
		if 0 < len(d.expect) {
			checkValErr(t, err, d.sdl, d.expect)
		} else {
			checkNil(t, err, "no error expected. %s", err)
		}
	}
}

func TestRootExeValidate(t *testing.T) {
	valExeTest(t, []*valTestData{
		{sdl: `{ artist(name: "A" name: "B") { name }}`, expect: "duplicate argument name to artist"},
		{sdl: `{ title @go }`, expect: "directive @go can not be applied to title, a FIELD"},
		{sdl: `{ ...frag @go } fragment frag on Query { title }`,
			expect: "directive @go can not be applied to frag, a FRAGMENT_SPREAD"},
		{sdl: `{ ... @go { title }}`, expect: "directive @go can not be applied to ..., a INLINE_FRAGMENT"},
		{sdl: `query bad @example { title }`, expect: "directive @example can not be applied to bad, a QUERY"},
		{sdl: `mutation bad @example { title }`, expect: "directive @example can not be applied to bad, a MUTATION"},
		{sdl: `subscription bad @example { title }`,
			expect: "directive @example can not be applied to bad, a SUBSCRIPTION"},
		{sdl: `{ title } fragment bad on Query @example {title}`,
			expect: "directive @example can not be applied to bad, a FRAGMENT_DEFINITION"},
		{sdl: `query Arg($art: String!){artist(name: $art){title}}`, expect: ""},
		{sdl: `query Arg($art: Bad){artist(name: $art){title}}`,
			expect: "validation: Bad is not a valid input type for $art"},
	}, false)
}
