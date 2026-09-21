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
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

const enumVarSdl = `
type Query {
  echo(genre: Genre): String
  echoAll(genres: [Genre!]): String
  echoInput(pick: Pick): String
}

input Pick {
  genre: Genre
  label: String
}

enum Genre { INDIE JAZZ ROCK }

schema { query: Query }
`

type enumVarQuery struct{}

func (q *enumVarQuery) Resolve(f *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch f.Name {
	case "echo":
		return fmt.Sprintf("%v", args["genre"]), nil
	case "echoAll":
		return fmt.Sprintf("%v", args["genres"]), nil
	case "echoInput":
		pick, _ := args["pick"].(map[string]interface{})
		return fmt.Sprintf("%v", pick["genre"]), nil
	}
	return nil, fmt.Errorf("no field %s", f.Name)
}

type enumVarSchema struct{ query enumVarQuery }

func (s *enumVarSchema) Resolve(f *ggql.Field, _ map[string]interface{}) (interface{}, error) {
	if f.Name == "query" {
		return &s.query, nil
	}
	return nil, fmt.Errorf("no field %s", f.Name)
}

func enumVarResolve(t *testing.T, src string, vars map[string]interface{}) string {
	t.Helper()
	root := ggql.NewRoot(&enumVarSchema{})
	if err := root.ParseString(enumVarSdl); err != nil {
		t.Fatalf("parsing the SDL should not fail: %s", err)
	}
	var b strings.Builder
	_ = ggql.WriteJSONValue(&b, root.ResolveString(src, "", vars), 0)
	return b.String()
}

// JSON has no enum type, so a client sends an enum value as a string. The
// specification allows that for a variable value even though a string literal
// written into a document must be refused.
func TestEnumVarAcceptsStringValue(t *testing.T) {
	out := enumVarResolve(t, `query ($g: Genre) { echo(genre: $g) }`, map[string]interface{}{"g": "JAZZ"})

	checkEqual(t, true, strings.Contains(out, `"echo": "JAZZ"`), "enum string in a variable should resolve. %s", out)
	checkEqual(t, false, strings.Contains(out, "errors"), "no errors expected. %s", out)
}

// Accepting the string does not mean skipping the check: it becomes a symbol
// and goes through the same validation as a literal token.
func TestEnumVarValidatesStringValue(t *testing.T) {
	out := enumVarResolve(t, `query ($g: Genre) { echo(genre: $g) }`, map[string]interface{}{"g": "POLKA"})

	checkEqual(t, true, strings.Contains(out, "not a valid enum value in Genre"),
		"an unknown enum value should be refused. %s", out)
}

// The conversion follows the declared type into lists and input objects, so a
// nested enum is handled the same way.
func TestEnumVarInsideListAndInput(t *testing.T) {
	out := enumVarResolve(t, `query ($g: [Genre!]) { echoAll(genres: $g) }`,
		map[string]interface{}{"g": []interface{}{"INDIE", "ROCK"}})
	checkEqual(t, true, strings.Contains(out, "INDIE"), "enum strings in a list should resolve. %s", out)

	out = enumVarResolve(t, `query ($p: Pick) { echoInput(pick: $p) }`,
		map[string]interface{}{"p": map[string]interface{}{"genre": "ROCK", "label": "x"}})
	checkEqual(t, true, strings.Contains(out, `"echoInput": "ROCK"`),
		"an enum string in an input object should resolve. %s", out)

	out = enumVarResolve(t, `query ($p: Pick) { echoInput(pick: $p) }`,
		map[string]interface{}{"p": map[string]interface{}{"genre": "POLKA"}})
	checkEqual(t, true, strings.Contains(out, "not a valid enum value in Genre"),
		"an unknown nested enum value should be refused. %s", out)
}

// A string written into the document is still refused, which is what the
// specification requires and what separates the two cases.
func TestEnumLiteralStringStillRefused(t *testing.T) {
	out := enumVarResolve(t, `query { echo(genre: "JAZZ") }`, nil)

	checkEqual(t, true, strings.Contains(out, "errors"),
		"a string literal in an enum position should be refused. %s", out)
}
