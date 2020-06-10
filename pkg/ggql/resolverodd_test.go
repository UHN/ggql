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
	"time"

	"github.com/uhn/ggql/pkg/ggql"
)

const oddSdl = `
type Query {
  ints: [Int]
  int64s: [Int64]
  any: [String]
  badAny: [String]
  strs: [String]
  bools: [Boolean]
  float32s: [Float]
  float64s: [Float]
  times: [Time]
  nestErrors: Int
  eat(what: Food!): String
  size(list: [Int!]): Int
  lunch(what: Bento!): String
}

"Japanese lunch box"
input Bento {
  main: Food!
  ordered: Time
}

enum Food {
  ANCHOVIES
  BEETS
  CORN
}
`

type oddSchema struct {
	Query oddQuery
}

type oddQuery struct {
}

func (s *oddSchema) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "query":
		return &s.Query, nil
	case "mutation":
	case "subscription":
	}
	return nil, fmt.Errorf("type Schema does not have field %s", field)
}

func (q *oddQuery) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "ints":
		return []int{1, 2, 3}, nil
	case "int64s":
		return []int64{1, 2, 3}, nil
	case "any":
		return []interface{}{1, 2, 3}, nil
	case "badAny":
		return []interface{}{1, []byte{}, 3}, nil
	case "strs":
		return []string{"a", "b", "c"}, nil
	case "bools":
		return []bool{true, false, true}, nil
	case "float32s":
		return []float32{1.1, 2.2, 3.3}, nil
	case "float64s":
		return []float64{1.1, 2.2, 3.3}, nil
	case "times":
		return []time.Time{
			time.Date(2019, 11, 11, 10, 9, 8, 7, time.UTC),
		}, nil
	case "nestErrors":
		return 2, ggql.Errors{
			fmt.Errorf("error one"),
			fmt.Errorf("error two"),
		}
	case "eat":
		food, _ := args["what"].(ggql.Symbol)
		return string(food), nil
	case "lunch":
		what := args["what"]
		return fmt.Sprintf("%T", what), nil
	case "size":
		list, _ := args["list"].([]interface{})
		return len(list), nil
	}
	return nil, fmt.Errorf("type Query does not have field %s", field)
}

func testOddResolve(t *testing.T, src, expect string) {
	ggql.Sort = true
	root := ggql.NewRoot(&oddSchema{})

	err := root.ParseString(oddSdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	var b strings.Builder

	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expect, b.String(), "result mismatch for %s", src)
}

func TestResolveInterfaceIntArray(t *testing.T) {
	src := `{ints}`
	expect := `{
  "data": {
    "ints": [
      1,
      2,
      3
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceInt64Array(t *testing.T) {
	src := `{int64s}`
	expect := `{
  "data": {
    "int64s": [
      1,
      2,
      3
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceAnyArray(t *testing.T) {
	src := `{any}`
	expect := `{
  "data": {
    "any": [
      "1",
      "2",
      "3"
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceBadAnyArray(t *testing.T) {
	src := `{badAny}`
	expect := `{
  "data": {
    "badAny": [
      "1",
      null,
      "3"
    ]
  },
  "errors": [
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: can not coerce a []uint8 into a String",
      "path": [
        "badAny",
        1
      ]
    }
  ]
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceStrsArray(t *testing.T) {
	src := `{strs}`
	expect := `{
  "data": {
    "strs": [
      "a",
      "b",
      "c"
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceBoolsArray(t *testing.T) {
	src := `{bools}`
	expect := `{
  "data": {
    "bools": [
      true,
      false,
      true
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceFloat32sArray(t *testing.T) {
	src := `{float32s}`
	expect := `{
  "data": {
    "float32s": [
      1.1,
      2.2,
      3.3
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceFloat64sArray(t *testing.T) {
	src := `{float64s}`
	expect := `{
  "data": {
    "float64s": [
      1.1,
      2.2,
      3.3
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceTimesArray(t *testing.T) {
	src := `{times}`
	expect := `{
  "data": {
    "times": [
      "2019-11-11T10:09:08.000000007Z"
    ]
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceNestedErrors(t *testing.T) {
	src := `{nestErrors}`
	expect := `{
  "data": {
    "nestErrors": 2
  },
  "errors": [
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: error one",
      "path": [
        "nestErrors"
      ]
    },
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: error two",
      "path": [
        "nestErrors"
      ]
    }
  ]
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceEnumArg(t *testing.T) {
	src := `{eat(what: ANCHOVIES)}`
	expect := `{
  "data": {
    "eat": "ANCHOVIES"
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceEnumBadArg(t *testing.T) {
	src := `{eat(what: DUCK)}`
	expect := `{
  "data": {
    "eat": null
  },
  "errors": [
    {
      "message": "resolve error: DUCK is not a valid enum value in Food",
      "path": [
        "eat",
        "what"
      ]
    }
  ]
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceInputEnumBad(t *testing.T) {
	src := `{lunch(what: {main: DUCK})}`
	expect := `{
  "data": {
    "lunch": null
  },
  "errors": [
    {
      "message": "resolve error: DUCK is not a valid enum value in Food",
      "path": [
        "lunch",
        "what"
      ]
    }
  ]
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceListArg(t *testing.T) {
	src := `{size(list: [1,2,3])}`
	expect := `{
  "data": {
    "size": 3
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceListNullArg(t *testing.T) {
	src := `{size(list: [1,null,3])}`
	expect := `{
  "data": {
    "size": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce null into a Int!",
      "path": [
        "size",
        "list"
      ]
    }
  ]
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceListEmptyArg(t *testing.T) {
	src := `{size(list: [])}`
	expect := `{
  "data": {
    "size": 0
  }
}
`
	testOddResolve(t, src, expect)
}

func TestResolveInterfaceNilTime(t *testing.T) {
	src := `{lunch(what: {main: CORN, ordered: null})}`
	expect := `{
  "data": {
    "lunch": "map[string]interface {}"
  }
}
`
	testOddResolve(t, src, expect)
}
