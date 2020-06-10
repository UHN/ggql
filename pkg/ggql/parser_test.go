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

func parseAndWrite(t *testing.T, in, expect string) {
	ggql.Sort = true
	value, err := ggql.ParseValueString(in)
	checkNil(t, err, "no error should be returned when parsing '%s'. %s", in, err)

	var b bytes.Buffer
	_ = ggql.WriteSDLValue(&b, value)
	checkEqual(t, expect, b.String(), "parsed vs given")
}

func TestParseValueOk(t *testing.T) {
	parseAndWrite(t, "3", "3")
	parseAndWrite(t, "-3", "-3")
	parseAndWrite(t, "3.7", "3.7")
	parseAndWrite(t, "true", "true")
	parseAndWrite(t, "false", "false")
	parseAndWrite(t, "null", "null")
	parseAndWrite(t, `"hello"`, `"hello"`)
	parseAndWrite(t, "ENUM", "ENUM")
	parseAndWrite(t, "", "null")
	parseAndWrite(t, "[]", "[]")
	parseAndWrite(t, "[ ]", "[]")
	parseAndWrite(t, "[1 2 3]", "[1, 2, 3]")
	parseAndWrite(t, "[1[ 2, 3][4,[5]]]", "[1, [2, 3], [4, [5]]]")
	parseAndWrite(t, "{ }", "{}")
	parseAndWrite(t, "{b:3 a:true }", "{a: true, b: 3}")
	parseAndWrite(t, `{"b":3 a:true}`, "{a: true, b: 3}")
	parseAndWrite(t, `{b:[]a:true}`, "{a: true, b: []}")
	parseAndWrite(t, `"z\b\f\n\r\t\"\\\/\u0005\u000a ぴーたー"`, `"z\b\f\n\r\t\"\\/\u0005\n ぴーたー"`)
	parseAndWrite(t, `"\u000a\u000A"`, `"\n\n"`)
	parseAndWrite(t, `"""\u000a\u000A"""`, `"\n\n"`)
	parseAndWrite(t, `"""abc "xyz"" def"""`, `"abc \"xyz\"\" def"`)
	parseAndWrite(t, `[""1]`, `["", 1]`)
}

func TestParseValueBad(t *testing.T) {
	for _, in := range []string{
		"3x",
		"3.2.1",
		`"hello`,
		"[3.2.1]",
		"[",
		"{",
		`{"a`,
		`{a 1}`,
		`{a: 1.2.3}`,
		`"\u00zz"`,
		`"\`,
		`"`,
		`""" "`,
		`""" ""`,
	} {
		_, err := ggql.ParseValueString(in)
		checkNotNil(t, err, "an error should be returned when parsing '%s'.", in)
	}
	for _, in := range []string{
		"   ",
		"ENUM",
		"1234",
		"[1  ",
		"{   ",
		"{a  ",
		"{ abc",
		`"\z"`,
		`" \n"`,
		`"\u00zz"`,
		` ""x`,
		`  ""`,
		`   ""`,
		`"""`,
	} {
		r := &failReader{max: 3, content: []byte(in)}
		_, err := ggql.ParseValue(r)
		checkNotNil(t, err, "an error should be returned when parsing and error reading.")
	}
	for _, in := range []string{
		`"""   """`,
		`"""  """`,
		`"""\u000a"""`,
	} {
		r := &failReader{max: 7, content: []byte(in)}
		_, err := ggql.ParseValue(r)
		checkNotNil(t, err, "an error should be returned when parsing and error reading.")
	}
}

func TestParseFail(t *testing.T) {
	r := failReader{max: 2, content: []byte(" \n")}
	_, err := ggql.ParseValue(&r)
	checkNotNil(t, err, "an error should be returned when parsing '%s' with a read error at %d.", r.content, r.max)
}

func TestParseBadInput(t *testing.T) {
	_, err := ggql.ParseValueString("[:]")
	checkNotNil(t, err, "an error should be returned when parsing [:].")
}
