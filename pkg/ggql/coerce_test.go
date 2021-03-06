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

// CQuery represents the query node in a data/resolver graph.
type CQuery struct {
}

// CSchema represents the top level of a GraphQL data/resolver graph.
type CSchema struct {
	Query CQuery
}

type Numbers struct {
	A int
	B float64
	C *Numbers
	D []int
	E []*Numbers
}

func (n *Numbers) sum() int {
	x := n.A + int(n.B)
	if n.C != nil {
		x += n.C.sum()
	}
	for _, n := range n.D {
		x += n
	}
	for _, n := range n.E {
		x += n.sum()
	}
	return x
}

type Unum struct {
	A uint
	D []int
}

func (s *CSchema) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	if field.Name == "query" {
		return &s.Query, nil
	}
	return nil, fmt.Errorf("type Schema does not have field %s", field)
}

func (q *CQuery) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	if field.Name == "sum" {
		var sum int
		switch numbers := args["numbers"].(type) {
		case *Numbers:
			sum = numbers.sum()
		case *Unum:
			sum = int(numbers.A)
		}
		return sum, nil
	}
	return nil, fmt.Errorf("type Query does not have field %s", field)
}

func TestCoerceInputOk(t *testing.T) {
	sdl := `
type Query {
  sum(numbers: Numbers): Int
}

input Numbers {
  a: Int = 1
  b: Float
  c: Numbers
  d: [Int!]
  e: [Numbers!]
}
`
	ggql.Sort = true
	schema := CSchema{}

	root := ggql.NewRoot(&schema)
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	err = root.RegisterType(&Numbers{}, "Numbers")
	checkNil(t, err, "no error should be returned when registering a type. %s", err)
	err = root.RegisterType(&CQuery{}, "Numbers")
	checkNotNil(t, err, "second register should fail")

	result := root.ResolveString("{sum(numbers: {b:2 c:{a:3 b:4} d:[5 6] e:[{a:7}]})}", "", nil)
	var b strings.Builder
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, `{
  "data": {
    "sum": 28
  }
}
`, b.String(), "result should match")
}

func TestCoerceInputReflectError(t *testing.T) {
	sdl := `
type Query {
  sum(numbers: Numbers): Int
}

input Numbers {
  a: Int = -1
  b: Int
  c: Numbers
  d: [String!]
}
`
	ggql.Sort = true
	schema := CSchema{}

	root := ggql.NewRoot(&schema)
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	err = root.RegisterType(&Unum{}, "Numbers")
	checkNil(t, err, "no error should be returned when registering a type. %s", err)

	result := root.ResolveString("{sum(numbers: {})}", "", nil)
	var b strings.Builder
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, `{
  "data": {
    "sum": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a int64 into a uint at a",
      "path": [
        "sum",
        "numbers"
      ]
    }
  ]
}
`, b.String(), "result should match")

	result = root.ResolveString("{sum(numbers: {a: 1})}", "", nil)
	b.Reset()
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, `{
  "data": {
    "sum": 1
  }
}
`, b.String(), "result should match")

	result = root.ResolveString("{sum(numbers: {a: -1})}", "", nil)
	b.Reset()
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, `{
  "data": {
    "sum": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a int32 into a uint at a",
      "path": [
        "sum",
        "numbers"
      ]
    }
  ]
}
`, b.String(), "result should match")

	result = root.ResolveString(`{sum(numbers: {a:1 d: ["xyz"]})}`, "", nil)
	b.Reset()
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, `{
  "data": {
    "sum": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a string into a int at d",
      "path": [
        "sum",
        "numbers"
      ]
    }
  ]
}
`, b.String(), "result should match")
}

func TestCoerceInputFloatInt(t *testing.T) {
	sdl := `
type Query {
  sum(numbers: Numbers): Int
}

input Numbers {
  a: Int = -1
  b: Int
  c: Numbers
}
`
	ggql.Sort = true
	schema := CSchema{}

	root := ggql.NewRoot(&schema)
	err := root.ParseString(sdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	err = root.RegisterType(&Numbers{}, "Numbers")
	checkNil(t, err, "no error should be returned when registering a type. %s", err)

	result := root.ResolveString("{sum(numbers: {b:-2})}", "", nil)
	var b strings.Builder
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, `{
  "data": {
    "sum": -3
  }
}
`, b.String(), "result should match")
}
