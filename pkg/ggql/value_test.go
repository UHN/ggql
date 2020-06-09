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
	"time"

	"github.com/uhn/ggql/pkg/ggql"
)

var sample = map[string]interface{}{
	"i": 17,
	"f": 1.23,
	"b": true,
	"n": nil,
	"s": "a string",
	"t": time.Date(2019, time.October, 18, 17, 16, 15, 123456789, time.UTC),
	"a": []interface{}{int16(1), int32(2), int64(3), float32(7.1), false, byte('A')},
}

func TestRootValueWriteJSON(t *testing.T) {
	ggql.Sort = true
	var b bytes.Buffer

	expect := `{"a":[1,2,3,7.1,false,"65"],"b":true,"f":1.23,"i":17,"n":null,"s":"a string","t":"2019-10-18T17:16:15.123456789Z"}`
	err := ggql.WriteJSONValue(&b, sample, -1)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, expect, b.String(), "write value content")

	b.Reset()
	expect = `{"a": [1, 2, 3, 7.1, false, "65"], "b": true, "f": 1.23, "i": 17, "n": null, "s": "a string", "t": "2019-10-18T17:16:15.123456789Z"}`
	err = ggql.WriteJSONValue(&b, sample, 0)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, expect, b.String(), "write value content")

	ggql.Sort = false
	b.Reset()
	err = ggql.WriteJSONValue(&b, map[string]interface{}{"x": 5})
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, `{"x": 5}`, b.String(), "write value content")
}

func TestRootValueWriteSDL(t *testing.T) {
	ggql.Sort = true
	var b bytes.Buffer

	expect := `{a:[1,2,3,7.1,false,"65"]b:true,f:1.23,i:17,n:null,s:"a string",t:"2019-10-18T17:16:15.123456789Z"}`
	err := ggql.WriteSDLValue(&b, sample, -1)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, expect, b.String(), "write value content")

	b.Reset()
	err = ggql.WriteSDLValue(&b, []interface{}{1, []interface{}{2}, 3}, -1)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, "[1[2]3]", b.String(), "write value content")

	b.Reset()
	err = ggql.WriteSDLValue(&b, []interface{}{1, map[string]interface{}{"x": map[string]interface{}{}, "y": 3}}, -1)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, "[1{x:{}y:3}]", b.String(), "write value content")
}

func TestRootValueWriteSDLIndent(t *testing.T) {
	ggql.Sort = true
	var b bytes.Buffer

	expect := `{
  a: [
    1
    2
    3
    7.1
    false
    "65"
  ]
  b: true
  f: 1.23
  i: 17
  n: null
  s: "a string"
  t: "2019-10-18T17:16:15.123456789Z"
}
`
	err := ggql.WriteSDLValue(&b, sample, 2)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, expect, b.String(), "write value content")

	b.Reset()
	err = ggql.WriteSDLValue(&b, []interface{}{1, 2, 3}, 2)
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, `[
  1
  2
  3
]
`, b.String(), "write value content")
}

func TestRootValueWriteEscape(t *testing.T) {
	ggql.Sort = true
	var b bytes.Buffer

	expect := `"\b\f\n\r\t\"\u0005 ぴーたー"`
	err := ggql.WriteSDLValue(&b, "\b\f\n\r\t\"\x05 ぴーたー")
	checkNil(t, err, "write value error - %s", err)
	checkEqual(t, expect, b.String(), "write value content")

	w := &failWriter{max: 2}
	err = ggql.WriteSDLValue(w, "abcd")
	checkNotNil(t, err, "write to failure should have an error")
}
