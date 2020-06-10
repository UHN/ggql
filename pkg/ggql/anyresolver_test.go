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

type myList []interface{}
type badList []interface{}

// Any is a resolver for any type of data.
type Any struct {
}

// Resolve a field on an object. The field argument is the name of the field
// to resolve. The args parameter includes the values associated with the
// arguments provided by the caller. The function should return the field's
// object or an error. A return of nil is also possible.
func (ar *Any) Resolve(obj interface{}, field *ggql.Field, args map[string]interface{}) (result interface{}, err error) {
	if m, _ := obj.(map[string]interface{}); m != nil {
		result = m[field.Name]
	} else {
		err = fmt.Errorf("expected a map[string]interface{}, not a %T", obj)
	}
	return
}

// Len of the list.
func (ar *Any) Len(list interface{}) int {
	switch tlist := list.(type) {
	case []interface{}:
		return len(tlist)
	case myList:
		return len(tlist)
	case badList:
		return len(tlist)
	}
	return 0
}

// Nth element in the list. If not a list or out of bounds, nil should be
// returned along with an error.
func (ar *Any) Nth(list interface{}, i int) (result interface{}, err error) {
	if i < 0 {
		return 0, fmt.Errorf("index must be >= 0, not %d", i)
	}
	switch tlist := list.(type) {
	case []interface{}:
		if len(tlist) <= i {
			return 0, fmt.Errorf("index must be less than the list length, %d > len %d", i, len(tlist))
		}
		return tlist[i], nil
	case myList:
		if len(tlist) <= i {
			return 0, fmt.Errorf("index must be less than the list length, %d > len %d", i, len(tlist))
		}
		return tlist[i], nil
	}
	return 0, fmt.Errorf("expected a []interface{}, not a %T", list)
}

func setupAnySongs(t *testing.T) *ggql.Root {
	may5 := &Date{Year: 2017, Month: 5, Day: 5}
	nov2 := &Date{Year: 2015, Month: 11, Day: 2}
	sep28 := &Date{Year: 2018, Month: 11, Day: 2}
	schema := map[string]interface{}{
		"query": map[string]interface{}{
			"title": "Songs",
			"artists": []interface{}{
				map[string]interface{}{
					"name":   "Fazerdaze",
					"origin": []string{"Morningside", "Auckland", "New Zealand"},
					"songs": myList{
						map[string]interface{}{"name": "Jennifer", "duration": 240, "release": may5},
						map[string]interface{}{"name": "Lucky Girl", "duration": 170, "release": may5},
						map[string]interface{}{"name": "Friends", "duration": 194, "release": may5},
						map[string]interface{}{"name": "Reel", "duration": 193, "release": nov2},
					},
				},
				map[string]interface{}{
					"name":   "Viagra Boys",
					"origin": []string{"Stockholm", "Sweden"},
					"songs": myList{
						map[string]interface{}{"name": "Down In The Basement", "duration": 216, "release": sep28},
						map[string]interface{}{"name": "Frogstrap", "duration": 195, "release": sep28},
						map[string]interface{}{"name": "Worms", "duration": 208, "release": sep28},
						map[string]interface{}{"name": "Amphetanarchy", "duration": 346, "release": sep28},
					},
				},
			},
		},
	}
	ggql.Sort = true
	root := ggql.NewRoot(schema)
	root.AnyResolver = &Any{}

	err := root.AddTypes(NewDateScalar())
	checkNil(t, err, "no error should be returned when adding a Date type. %s", err)

	err = root.ParseString(songsSdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	return root
}

func testAnyResolve(t *testing.T, sdl, src, expect string) {
	root := setupAnySongs(t)

	if 0 < len(sdl) {
		err := root.ParseString(sdl)
		checkNil(t, err, "extend should not fail. %s", err)
	}
	var b strings.Builder

	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expect, b.String(), "result mismatch for %s", src)
}

func TestResolveAnySimple(t *testing.T) {
	src := `{title }`
	testAnyResolve(t, "", src, expectTitle)
}

func TestResolveAnyMore(t *testing.T) {
	src := `{artists{name,songs{names,release}}}`
	expect := `{
  "data": {
    "artists": [
      {
        "name": "Fazerdaze",
        "songs": [
          {
            "release": "2017-05-05"
          },
          {
            "release": "2017-05-05"
          },
          {
            "release": "2017-05-05"
          },
          {
            "release": "2015-11-02"
          }
        ]
      },
      {
        "name": "Viagra Boys",
        "songs": [
          {
            "release": "2018-11-02"
          },
          {
            "release": "2018-11-02"
          },
          {
            "release": "2018-11-02"
          },
          {
            "release": "2018-11-02"
          }
        ]
      }
    ]
  },
  "errors": [
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        0,
        "songs",
        0,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        0,
        "songs",
        1,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        0,
        "songs",
        2,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        0,
        "songs",
        3,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        1,
        "songs",
        0,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        1,
        "songs",
        1,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        1,
        "songs",
        2,
        "names"
      ]
    },
    {
      "locations": [
        {
          "column": 22,
          "line": 1
        }
      ],
      "message": "resolve error: names is not a field in Song",
      "path": [
        "artists",
        1,
        "songs",
        3,
        "names"
      ]
    }
  ]
}
`
	testAnyResolve(t, "", src, expect)
}

func TestResolveAnyError(t *testing.T) {
	schema := map[string]interface{}{
		"query": map[string]interface{}{
			"artists": badList{
				map[string]interface{}{"name": "Fazerdaze"},
				map[string]interface{}{"name": "Viagra Boys"},
			},
		},
	}
	ggql.Sort = true
	root := ggql.NewRoot(schema)
	root.AnyResolver = &Any{}

	err := root.AddTypes(NewDateScalar())
	checkNil(t, err, "no error should be returned when adding a Date type. %s", err)

	err = root.ParseString(songsSdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	src := `{artists{name}}`
	expect := `{
  "data": {
    "artists": [
      0,
      0
    ]
  },
  "errors": [
    {
      "message": "resolve error: expected a []interface{}, not a ggql_test.badList",
      "path": [
        "artists",
        0
      ]
    },
    {
      "message": "resolve error: expected a []interface{}, not a ggql_test.badList",
      "path": [
        "artists",
        1
      ]
    }
  ]
}
`
	var b strings.Builder

	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expect, b.String(), "result mismatch for %s", src)
}
