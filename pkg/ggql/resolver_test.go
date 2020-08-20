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
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

const (
	expectTitle = `{
  "data": {
    "title": "Songs"
  }
}
`
	expectArtistFazerdaze = `{
  "data": {
    "artist": {
      "name": "Fazerdaze"
    }
  }
}
`
	expectArtistsName = `{
  "data": {
    "artists": [
      {
        "name": "Fazerdaze"
      },
      {
        "name": "Viagra Boys"
      }
    ]
  }
}
`
)

func testResolve(t *testing.T, sdl, src, expect string, varPairs ...interface{}) {
	root := setupTestSongs(t, nil)

	if 0 < len(sdl) {
		err := root.ParseString(sdl)
		checkNil(t, err, "extend should not fail. %s", err)
	}
	var b strings.Builder
	var vars map[string]interface{}

	if 0 < len(varPairs) {
		vars = map[string]interface{}{}
		var key string
		for _, x := range varPairs {
			if 0 < len(key) {
				vars[key] = x
				key = ""
			} else if s, _ := x.(string); 0 < len(s) {
				key = s
			}
		}
	}
	result := root.ResolveString(src, "", vars)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expect, b.String(), "result mismatch for %s", src)
}

func TestResolveInterfaceSimpleString(t *testing.T) {
	src := `{title}`
	testResolve(t, "", src, expectTitle)
}

func TestResolveInterfaceSimpleBytes(t *testing.T) {
	root := setupTestSongs(t, nil)

	src := `{ title }`
	var b strings.Builder

	result := root.ResolveBytes([]byte(src), "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expectTitle, b.String(), "result mismatch for %s", src)
}

func TestResolveInterfaceVars(t *testing.T) {
	root := setupTestSongs(t, nil)

	src := `query test($id: String = "Who"){artist(name: $id){name}}`
	exe, err := root.ParseExecutableString(src)
	checkNil(t, err, "parsing executable fail. %s", err)

	var b strings.Builder
	var result interface{}
	vars := map[string]interface{}{"id": "Fazerdaze"}

	result, err = root.ResolveExecutable(exe, "", vars)
	checkNil(t, err, "resolving executable fail. %s", err)

	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expectArtistFazerdaze, b.String(), "result mismatch for %s", src)
}

func TestResolveExecutableErrors(t *testing.T) {
	root := setupTestSongs(t, nil)

	src := `query test($id: String = "Who"){artist(name: $id){name,bad}}`
	exe, err := root.ParseExecutableString(src)
	checkNil(t, err, "parsing executable fail. %s", err)

	vars := map[string]interface{}{"id": "Fazerdaze"}

	_, err = root.ResolveExecutable(exe, "", vars)
	checkEqual(t, `Errors{
  resolve error: bad is not a field in Artist at artist.bad from 1:57
}
`, err.Error(), "result mismatch for %s", src)
}

func TestResolveInterfaceParseError(t *testing.T) {
	src := `{title`
	expect := `{
  "errors": [
    {
      "locations": [
        {
          "column": 7,
          "line": 1
        }
      ],
      "message": "parse error: selection set not terminated with a '}'"
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceOpError(t *testing.T) {
	src := `fragment Frag on Artist {title}`
	expect := `{
  "data": null,
  "errors": [
    {
      "message": "resolve error, could not determine operation to evaluate"
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceAlias(t *testing.T) {
	src := `query test{header:title}`
	expect := `{
  "data": {
    "header": "Songs"
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceInputObject(t *testing.T) {
	src := `{options(misc:{width:3 sizes:[1,2,3]})}`
	expect := `{
  "data": {
    "options": "{misc: {sizes: [1, 2, 3], width: 3}}"
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceArgs(t *testing.T) {
	src := `query test($id:String="Fazerdaze"){artist(name: $id){name}}`
	testResolve(t, "", src, expectArtistFazerdaze)
}

func TestResolveInterfaceNotOutput(t *testing.T) {
	src := `query test{artist(name: "Fazerdaze")}`
	expect := `{
  "data": {
    "artist": {
    }
  },
  "errors": [
    {
      "message": "resolve error: Artist is not a valid output leaf type",
      "path": [
        "artist"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceList(t *testing.T) {
	src := `query test{artists{name,origin}}`
	expect := `{
  "data": {
    "artists": [
      {
        "name": "Fazerdaze",
        "origin": [
          "Morningside",
          "Auckland",
          "New Zealand"
        ]
      },
      {
        "name": "Viagra Boys",
        "origin": [
          "Stockholm",
          "Sweden"
        ]
      }
    ]
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUtypename(t *testing.T) {
	src := `query test{artist(name:"Fazerdaze"){songs{name,__typename},__typename}__typename}`
	expect := `{
  "data": {
    "__typename": "Query",
    "artist": {
      "__typename": "Artist",
      "songs": [
        {
          "__typename": "Song",
          "name": "Jennifer"
        },
        {
          "__typename": "Song",
          "name": "Lucky Girl"
        },
        {
          "__typename": "Song",
          "name": "Friends"
        },
        {
          "__typename": "Song",
          "name": "Reel"
        }
      ]
    }
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceInline(t *testing.T) {
	src := `query test{artists{...{name}}}`
	testResolve(t, "", src, expectArtistsName)
}

func TestResolveInterfaceFragment(t *testing.T) {
	src := `query test{artists{...frag}} fragment frag on Artist {name}`
	testResolve(t, "", src, expectArtistsName)
}

func TestResolveInterfaceFragmentError(t *testing.T) {
	src := `query test{artists{...frag}} fragment frag on Artist {name,bad}`
	expect := `{
  "data": {
    "artists": [
      {
        "name": "Fazerdaze"
      },
      {
        "name": "Viagra Boys"
      }
    ]
  },
  "errors": [
    {
      "locations": [
        {
          "column": 61,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in Artist",
      "path": [
        "artists",
        0,
        "fragment at 1:28",
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 61,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in Artist",
      "path": [
        "artists",
        1,
        "fragment at 1:28",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUschemaTypes(t *testing.T) {
	sdl := `
schema {query: Query, mutation: Mutation}
extend schema {subscription: Subby}
type Subby { dummy: Int }
`
	src := `query {__schema{types{name}}}`
	expect := `{
  "data": {
    "__schema": {
      "types": [
        {
          "name": "schema"
        },
        {
          "name": "__Schema"
        },
        {
          "name": "Query"
        },
        {
          "name": "Mutation"
        },
        {
          "name": "Subscription"
        },
        {
          "name": "Artist"
        },
        {
          "name": "Song"
        },
        {
          "name": "Subby"
        },
        {
          "name": "__Directive"
        },
        {
          "name": "__EnumValue"
        },
        {
          "name": "__Field"
        },
        {
          "name": "__InputValue"
        },
        {
          "name": "__Type"
        },
        {
          "name": "Misc"
        },
        {
          "name": "__DirectiveLocation"
        },
        {
          "name": "__TypeKind"
        },
        {
          "name": "Boolean"
        },
        {
          "name": "Date"
        },
        {
          "name": "Float"
        },
        {
          "name": "Float64"
        },
        {
          "name": "ID"
        },
        {
          "name": "Int"
        },
        {
          "name": "Int64"
        },
        {
          "name": "String"
        },
        {
          "name": "Time"
        }
      ]
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceUUschemaError(t *testing.T) {
	src := `query {__schema{bad}}`
	expect := `{
  "data": {
    "__schema": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 18,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Schema",
      "path": [
        "__schema",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUschemaDirs(t *testing.T) {
	src := `query {__schema{directives{name}}}`
	expect := `{
  "data": {
    "__schema": {
      "directives": [
        {
          "name": "deprecated"
        },
        {
          "name": "example"
        },
        {
          "name": "go"
        },
        {
          "name": "include"
        },
        {
          "name": "skip"
        }
      ]
    }
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUschemaQueryType(t *testing.T) {
	src := `query {__schema{queryType{name}}}`
	expect := `{
  "data": {
    "__schema": {
      "queryType": {
        "name": "Query"
      }
    }
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUschemaMutationType(t *testing.T) {
	src := `query {__schema{mutationType{name}}}`
	expect := `{
  "data": {
    "__schema": {
      "mutationType": {
        "name": "Mutation"
      }
    }
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUschemaSubscriptionType(t *testing.T) {
	sdl := `
schema {query: Query, mutation: Mutation}
extend schema {subscription: Subby}
type Subby { dummy: Int }
`
	src := `query {__schema{subscriptionType{name}}}`
	expect := `{
  "data": {
    "__schema": {
      "subscriptionType": {
        "name": "Subby"
      }
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceUUtype(t *testing.T) {
	sdl := `
interface Namely {
  name: String!
}

extend type Artist implements Namely {
  cool(age: Int): Boolean @deprecated
}
`
	src := `query {
  __type(name:"Artist"){
    name
    kind
    description
    fields{
      name
      description
      args{
        name
        description
        type{name}
        defaultValue
      }
      type{name}
      isDeprecated
      deprecationReason
    },
    interfaces{name}
    possibleTypes{name}
    enumValues{name}
    inputFields{name}
    ofType
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "Song player or singer",
      "enumValues": null,
      "fields": [
        {
          "args": [
          ],
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "name",
          "type": {
            "name": "String!"
          }
        },
        {
          "args": [
          ],
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "songs",
          "type": {
            "name": "[Song]"
          }
        },
        {
          "args": [
          ],
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "origin",
          "type": {
            "name": "[String]"
          }
        }
      ],
      "inputFields": null,
      "interfaces": [
        {
          "name": "Namely"
        }
      ],
      "kind": "OBJECT",
      "name": "Artist",
      "ofType": null,
      "possibleTypes": null
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceUUtypeWdeprecated(t *testing.T) {
	sdl := `
interface Namely {
  name: String!
}

extend type Artist implements Namely {
  cool(age: Int): Boolean @deprecated
}
`
	src := `query {
  __type(name:"Artist"){
    name
    kind
    description
    fields(includeDeprecated: true){
      name
      description
      args{
        name
        description
        type{name}
        defaultValue
      }
      type{name}
      isDeprecated
      deprecationReason
    },
    interfaces{name}
    possibleTypes{name}
    enumValues(includeDeprecated: true) {name}
    inputFields{name}
    ofType
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "Song player or singer",
      "enumValues": null,
      "fields": [
        {
          "args": [
          ],
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "name",
          "type": {
            "name": "String!"
          }
        },
        {
          "args": [
          ],
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "songs",
          "type": {
            "name": "[Song]"
          }
        },
        {
          "args": [
          ],
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "origin",
          "type": {
            "name": "[String]"
          }
        },
        {
          "args": [
            {
              "defaultValue": null,
              "description": "",
              "name": "age",
              "type": {
                "name": "Int"
              }
            }
          ],
          "deprecationReason": "\"No longer supported\"",
          "description": "",
          "isDeprecated": true,
          "name": "cool",
          "type": {
            "name": "Boolean"
          }
        }
      ],
      "inputFields": null,
      "interfaces": [
        {
          "name": "Namely"
        }
      ],
      "kind": "OBJECT",
      "name": "Artist",
      "ofType": null,
      "possibleTypes": null
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceArgFail(t *testing.T) {
	sdl := `
extend type Artist {
  cool(age: Int): Boolean
}
`
	src := `query {__type(name:"Artist"){fields{name,args{bad}}}}`
	expect := `{
  "data": {
    "__type": {
      "fields": [
        {
          "args": [
          ],
          "name": "name"
        },
        {
          "args": [
          ],
          "name": "songs"
        },
        {
          "args": [
          ],
          "name": "origin"
        },
        {
          "args": [
            {
            }
          ],
          "name": "cool"
        }
      ]
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 48,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __InputValue",
      "path": [
        "__type",
        "fields",
        3,
        "args",
        0,
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceFieldFail(t *testing.T) {
	src := `query {__type(name:"Artist"){name,fields{bad}}}`
	expect := `{
  "data": {
    "__type": {
      "fields": [
        {
        },
        {
        },
        {
        }
      ],
      "name": "Artist"
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 43,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Field",
      "path": [
        "__type",
        "fields",
        0,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 43,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Field",
      "path": [
        "__type",
        "fields",
        1,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 43,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Field",
      "path": [
        "__type",
        "fields",
        2,
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceDir(t *testing.T) {
	src := `query {__schema{directives{name,description,locations,args{name}}}}`
	expect := `{
  "data": {
    "__schema": {
      "directives": [
        {
          "args": [
            {
              "name": "reason"
            }
          ],
          "description": "",
          "locations": [
            "FIELD_DEFINITION",
            "ENUM_VALUE"
          ],
          "name": "deprecated"
        },
        {
          "args": [
          ],
          "description": "",
          "locations": [
            "VARIABLE_DEFINITION"
          ],
          "name": "example"
        },
        {
          "args": [
            {
              "name": "type"
            }
          ],
          "description": "",
          "locations": [
            "SCHEMA",
            "QUERY",
            "MUTATION",
            "SUBSCRIPTION",
            "OBJECT",
            "FIELD_DEFINITION"
          ],
          "name": "go"
        },
        {
          "args": [
            {
              "name": "if"
            }
          ],
          "description": "",
          "locations": [
            "FIELD",
            "FRAGMENT_SPREAD",
            "INLINE_FRAGMENT"
          ],
          "name": "include"
        },
        {
          "args": [
            {
              "name": "if"
            }
          ],
          "description": "",
          "locations": [
            "FIELD",
            "FRAGMENT_SPREAD",
            "INLINE_FRAGMENT"
          ],
          "name": "skip"
        }
      ]
    }
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceDirError(t *testing.T) {
	src := `query {__schema{directives{bad}}}`
	expect := `{
  "data": {
    "__schema": {
      "directives": [
        {
        },
        {
        },
        {
        },
        {
        },
        {
        }
      ]
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 29,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Directive",
      "path": [
        "__schema",
        "directives",
        0,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 29,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Directive",
      "path": [
        "__schema",
        "directives",
        1,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 29,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Directive",
      "path": [
        "__schema",
        "directives",
        2,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 29,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Directive",
      "path": [
        "__schema",
        "directives",
        3,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 29,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Directive",
      "path": [
        "__schema",
        "directives",
        4,
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceEnum(t *testing.T) {
	sdl := `
enum Genre {INDIE JAZZ @deprecated(reason: "no reason") ROCK}
`
	src := `query {
  __type(name:"Genre"){
    name
    kind
    description
    enumValues{
      name
      description
      isDeprecated
      deprecationReason
    }
    fields
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "",
      "enumValues": [
        {
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "INDIE"
        },
        {
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "ROCK"
        }
      ],
      "fields": null,
      "kind": "ENUM",
      "name": "Genre"
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceEnumWdeprecated(t *testing.T) {
	sdl := `
enum Genre {
  INDIE
  JAZZ @deprecated(reason: "no reason")
  ROCK
}
`
	src := `query {
  __type(name:"Genre"){
    name
    kind
    description
    enumValues(includeDeprecated: true){
      name
      description
      isDeprecated
      deprecationReason
    }
    fields
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "",
      "enumValues": [
        {
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "INDIE"
        },
        {
          "deprecationReason": "no reason",
          "description": "",
          "isDeprecated": true,
          "name": "JAZZ"
        },
        {
          "deprecationReason": null,
          "description": "",
          "isDeprecated": false,
          "name": "ROCK"
        }
      ],
      "fields": null,
      "kind": "ENUM",
      "name": "Genre"
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceEnumError(t *testing.T) {
	src := `query {__type(name:"__TypeKind"){bad}}`
	expect := `{
  "data": {
    "__type": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 35,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Type",
      "path": [
        "__type",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceEnumValueError(t *testing.T) {
	sdl := `
enum Genre {
  INDIE
  JAZZ @deprecated(reason: "no reason")
  ROCK
}
`
	src := `query {__type(name:"Genre"){enumValues{bad}}}`
	expect := `{
  "data": {
    "__type": {
      "enumValues": [
        {
        },
        {
        }
      ]
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 41,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __EnumValue",
      "path": [
        "__type",
        "enumValues",
        0,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 41,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __EnumValue",
      "path": [
        "__type",
        "enumValues",
        1,
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceInput(t *testing.T) {
	sdl := `
input FindInfo {
  artist: String = "Who",
  song: String
}
`
	src := `query {
  __type(name:"FindInfo"){
    kind
    name
    description
    inputFields{
      name
      description
      type{name}
      defaultValue
    }
    ofType
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "",
      "inputFields": [
        {
          "defaultValue": "Who",
          "description": "",
          "name": "artist",
          "type": {
            "name": "String"
          }
        },
        {
          "defaultValue": null,
          "description": "",
          "name": "song",
          "type": {
            "name": "String"
          }
        }
      ],
      "kind": "INPUT_OBJECT",
      "name": "FindInfo",
      "ofType": null
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceInputError(t *testing.T) {
	sdl := `
input Inner {
  artist: String = "Who",
  song: String
}
`
	src := `query {__type(name:"Inner"){bad}}`
	expect := `{
  "data": {
    "__type": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 30,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Type",
      "path": [
        "__type",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceInputFieldError(t *testing.T) {
	sdl := `
input InTo {
  artist: String = "Who",
  song: String
}
`
	src := `query {__type(name:"InTo"){inputFields{bad}}}`
	expect := `{
  "data": {
    "__type": {
      "inputFields": [
        {
        },
        {
        }
      ]
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 41,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __InputValue",
      "path": [
        "__type",
        "inputFields",
        0,
        "bad"
      ]
    },
    {
      "locations": [
        {
          "column": 41,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __InputValue",
      "path": [
        "__type",
        "inputFields",
        1,
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceInterface(t *testing.T) {
	sdl := `
interface Named {
  name: String!
}

extend type Artist implements Named {
  cool(age: Int): Boolean @deprecated
}
`
	src := `query {
  __type(name:"Named"){
    name
    kind
    description
    fields{
      name
    },
    interfaces
    possibleTypes{name}
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "",
      "fields": [
        {
          "name": "name"
        }
      ],
      "interfaces": null,
      "kind": "INTERFACE",
      "name": "Named",
      "possibleTypes": [
        {
          "name": "Artist"
        }
      ]
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceInterfaceError(t *testing.T) {
	sdl := `
interface Namie {
  name: String!
}

extend type Artist implements Namie {
  cool(age: Int): Boolean @deprecated
}
`
	src := `query {__type(name:"Namie"){bad}}`
	expect := `{
  "data": {
    "__type": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 30,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Type",
      "path": [
        "__type",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceObjectError(t *testing.T) {
	src := `query Q($id: String = "Artist") {__type(name:$id){bad}}`
	expect := `{
  "data": {
    "__type": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 52,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Type",
      "path": [
        "__type",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUtypeNotQuery(t *testing.T) {
	src := `query {artists{__type(name: "Who"){name}}}`
	expect := `{
  "data": {
    "artists": [
      {
      },
      {
      }
    ]
  },
  "errors": [
    {
      "locations": [
        {
          "column": 17,
          "line": 1
        }
      ],
      "message": "resolve error: __type meta-field is only on the query object",
      "path": [
        "artists",
        0,
        "__type"
      ]
    },
    {
      "locations": [
        {
          "column": 17,
          "line": 1
        }
      ],
      "message": "resolve error: __type meta-field is only on the query object",
      "path": [
        "artists",
        1,
        "__type"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUschemaNotQuery(t *testing.T) {
	src := `query {artists{__schema{types{name}}}}`
	expect := `{
  "data": {
    "artists": [
      {
      },
      {
      }
    ]
  },
  "errors": [
    {
      "locations": [
        {
          "column": 17,
          "line": 1
        }
      ],
      "message": "resolve error: __schema meta-field is only on the query object",
      "path": [
        "artists",
        0,
        "__schema"
      ]
    },
    {
      "locations": [
        {
          "column": 17,
          "line": 1
        }
      ],
      "message": "resolve error: __schema meta-field is only on the query object",
      "path": [
        "artists",
        1,
        "__schema"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUnion(t *testing.T) {
	sdl := `union Any = Artist | Song`
	src := `query {
  __type(name:"Any"){
    name
    kind
    description
    fields{name}
    possibleTypes{name}
    interfaces
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "",
      "fields": null,
      "interfaces": null,
      "kind": "UNION",
      "name": "Any",
      "possibleTypes": [
        {
          "name": "Artist"
        },
        {
          "name": "Song"
        }
      ]
    }
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceUnionError(t *testing.T) {
	sdl := `union Any = Artist | Song`
	src := `query {__type(name:"Any"){bad}}`
	expect := `{
  "data": {
    "__type": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 28,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Type",
      "path": [
        "__type",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceUnionTypename(t *testing.T) {
	sdl := `union Any = Artist | Song
extend type Query {
  all: [Any]
}
`
	src := `query {
  all{
    __typename
    ... on Artist {
      name
    }
    ... on Song {
      duration
    }
  }
}`
	expect := `{
  "data": {
    "all": [
      {
        "__typename": "Artist",
        "name": "Fazerdaze"
      },
      {
        "__typename": "Song",
        "duration": 240
      },
      {
        "__typename": "Song",
        "duration": 170
      },
      {
        "__typename": "Song",
        "duration": 194
      },
      {
        "__typename": "Song",
        "duration": 193
      },
      {
        "__typename": "Artist",
        "name": "Viagra Boys"
      },
      {
        "__typename": "Song",
        "duration": 216
      },
      {
        "__typename": "Song",
        "duration": 195
      },
      {
        "__typename": "Song",
        "duration": 208
      },
      {
        "__typename": "Song",
        "duration": 346
      }
    ]
  }
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceUnionNoMeta(t *testing.T) {
	sdl := `union Any = Artist | Song
extend type Query {
  all: [Any]
}

extend type Song @go(type: "Bad") {
}
`
	src := `query {
  all{
    __typename
    ... on Artist {
      name
    }
    ... on Song {
      duration
    }
  }
}`
	expect := `{
  "data": {
    "all": [
      {
        "__typename": "Artist",
        "name": "Fazerdaze"
      },
      null,
      null,
      null,
      null,
      {
        "__typename": "Artist",
        "name": "Viagra Boys"
      },
      null,
      null,
      null,
      null
    ]
  },
  "errors": [
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        1
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        2
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        3
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        4
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        6
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        7
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        8
      ]
    },
    {
      "locations": [
        {
          "column": 6,
          "line": 25
        }
      ],
      "message": "resolve error: failed to determine union member Song implementation type. Use @go directive",
      "path": [
        "all",
        9
      ]
    }
  ]
}
`
	testResolve(t, sdl, src, expect)
}

func TestResolveInterfaceScalar(t *testing.T) {
	src := `query {
  __type(name:"Int"){
    name
    kind
    description
    interfaces
  }
}`
	expect := `{
  "data": {
    "__type": {
      "description": "",
      "interfaces": null,
      "kind": "SCALAR",
      "name": "Int"
    }
  }
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceScalarError(t *testing.T) {
	src := `query {__type(name:"Int"){bad}}`
	expect := `{
  "data": {
    "__type": {
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 28,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in __Type",
      "path": [
        "__type",
        "bad"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceUUtypeNoName(t *testing.T) {
	src := `query {__type{name}}`
	expect := `{
  "data": {
    "__type": null
  },
  "errors": [
    {
      "locations": [
        {
          "column": 9,
          "line": 1
        }
      ],
      "message": "resolve error: __type meta-field is missing a name argument",
      "path": [
        "__type"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceSkip(t *testing.T) {
	src := `query Skip($test: Boolean = true){
  artist(name:"Fazerdaze"){
    name
    origin @skip(if: $test)
  }
}`
	testResolve(t, "", src, expectArtistFazerdaze)
}

func TestResolveInterfaceSkipBool(t *testing.T) {
	src := `query Skip {
  artist(name:"Fazerdaze"){
    name
    origin @skip(if: true)
  }
}`
	testResolve(t, "", src, expectArtistFazerdaze)
}

func TestResolveInterfaceSkipError(t *testing.T) {
	src := `query Skip($test: Int = 3){
  artist(name:"Fazerdaze"){
    name
    origin @skip(if: $test)
  }
}`
	expect := `{
  "data": {
    "artist": {
      "name": "Fazerdaze"
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 6,
          "line": 4
        }
      ],
      "message": "resolve error: test is not a valid 'if' value for @skip",
      "path": [
        "artist",
        "origin"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

func TestResolveInterfaceInclude(t *testing.T) {
	src := `query Include($test: Boolean = false){
  artist(name:"Fazerdaze"){
    name
    origin @include(if: $test)
  }
}`
	testResolve(t, "", src, expectArtistFazerdaze)
}

func TestResolveInterfaceIncludeBool(t *testing.T) {
	src := `query Include {
  artist(name:"Fazerdaze"){
    name
    origin @include(if: false)
  }
}`
	testResolve(t, "", src, expectArtistFazerdaze)
}

func TestResolveInterfaceIncludeError(t *testing.T) {
	src := `query Include($test: Int = 3){
  artist(name:"Fazerdaze"){
    name
    origin @include(if: $test)
  }
}`
	expect := `{
  "data": {
    "artist": {
      "name": "Fazerdaze"
    }
  },
  "errors": [
    {
      "locations": [
        {
          "column": 6,
          "line": 4
        }
      ],
      "message": "resolve error: test is not a valid 'if' value for @include",
      "path": [
        "artist",
        "origin"
      ]
    }
  ]
}
`
	testResolve(t, "", src, expect)
}

// Resolve validation tests

func TestResolveValidateExtraArg(t *testing.T) {
	testResolve(t, "", `{artist(id: 123, name: "X"){name}}`, `{
  "data": {
  },
  "errors": [
    {
      "locations": [
        {
          "column": 9,
          "line": 1
        }
      ],
      "message": "validation: id is not an argument to artist",
      "path": [
        "artist"
      ]
    }
  ]
}
`)
}

func TestResolveValidateWrongVarType(t *testing.T) {
	testResolve(t, "", `query Arg($name: Int) {artist(name: $name){name}}`, `{
  "data": null,
  "errors": [
    {
      "locations": [
        {
          "column": 13,
          "line": 1
        }
      ],
      "message": "resolve error: can not coerce a string into a Int for name"
    }
  ]
}
`, "name", "Fazerdaze")
}

func TestResolveValidateVarTypeMismatch(t *testing.T) {
	testResolve(t, "", `query Arg($name: Int) {artist(name: $name){name}}`, `{
  "data": {
    "artist": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a int32 into a String",
      "path": [
        "artist",
        "name"
      ]
    }
  ]
}
`, "name", 123)
}

func TestResolveValidateCoerceIDFromIDInt(t *testing.T) {
	testResolve(t, "", `query Arg($name: ID) {byID(id: $name){name}}`, `{
  "data": {
    "byID": null
  }
}
`, "name", 123)
}

func TestResolveValidateCoerceIDFromID(t *testing.T) {
	testResolve(t, "", `query Arg($name: ID) {byID(id: $name){name}}`, `{
  "data": {
    "byID": {
      "name": "Fazerdaze"
    }
  }
}
`, "name", "Fazerdaze")
}

func TestResolveValidateCoerceIDFromString(t *testing.T) {
	testResolve(t, "", `query Arg($name: String) {byID(id: $name){name}}`, `{
  "data": {
    "byID": {
      "name": "Fazerdaze"
    }
  }
}
`, "name", "Fazerdaze")
}

func TestResolveValidateCoerceIDFromInt(t *testing.T) {
	testResolve(t, "", `query Arg($name: Int) {byID(id: $name){name}}`, `{
  "data": {
    "byID": null
  }
}
`, "name", 123)
}

func TestResolveValidateCoerceIDFromFloat(t *testing.T) {
	testResolve(t, "", `query Arg($name: Float) {byID(id: $name){name}}`, `{
  "data": {
    "byID": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a float32 into a ID",
      "path": [
        "byID",
        "id"
      ]
    }
  ]
}
`, "name", 1.23)
}

func TestResolveValidateCoerceObjectGood(t *testing.T) {
	testResolve(t, "", `query Arg($wide: Int) {options(misc: {width: $wide, sizes: [1,2,3]})}`, `{
  "data": {
    "options": "{misc: {sizes: [1, 2, 3], width: 31}}"
  }
}
`, "wide", 31)
}

func TestResolveValidateCoerceObjectWrongType(t *testing.T) {
	testResolve(t, "", `query Arg($wide: String) {options(misc: {width: $wide, sizes: [1,2,3]})}`, `{
  "data": {
    "options": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a string into a Int",
      "path": [
        "options",
        "misc"
      ]
    }
  ]
}
`, "wide", "abc")
}

func TestResolveValidateCoerceListWrongType(t *testing.T) {
	testResolve(t, "", `query Arg($size: String) {options(misc: {width: 31, sizes: [$size,2,3]})}`, `{
  "data": {
    "options": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a string into a Int",
      "path": [
        "options",
        "misc"
      ]
    }
  ]
}
`, "size", "small")
}

func TestResolveValidateArgIntID(t *testing.T) {
	testResolve(t, "", `query {byID(id: 123){name}}`, `{
  "data": {
    "byID": null
  }
}
`)
}

func TestResolveValidateArgStringID(t *testing.T) {
	testResolve(t, "", `query {byID(id: "hello"){name}}`, `{
  "data": {
    "byID": null
  }
}
`)
}

func TestResolveValidateArgBooleanID(t *testing.T) {
	testResolve(t, "", `query {byID(id: true){name}}`, `{
  "data": {
    "byID": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a bool into a ID",
      "path": [
        "byID",
        "id"
      ]
    }
  ]
}
`)
}

func TestResolveValidateArgInput(t *testing.T) {
	testResolve(t, "", `query {options(misc: {width: true sizes: [1,2]})}`, `{
  "data": {
    "options": null
  },
  "errors": [
    {
      "message": "resolve error: can not coerce a bool into a Int",
      "path": [
        "options",
        "misc"
      ]
    }
  ]
}
`)
}

func TestResolveValidateArgInputExtra(t *testing.T) {
	testResolve(t, "", `query {options(misc: {width: 33 sizes: [1,2] height: 66 })}`, `{
  "data": {
    "options": null
  },
  "errors": [
    {
      "message": "resolve error: height not a field in Misc",
      "path": [
        "options",
        "misc"
      ]
    }
  ]
}
`)
}

func TestResolveValidateArgInputMissing(t *testing.T) {
	testResolve(t, "", `query {options(misc: {sizes: [1,2]})}`, `{
  "data": {
    "options": null
  },
  "errors": [
    {
      "message": "resolve error: width is required but missing",
      "path": [
        "options",
        "misc"
      ]
    }
  ]
}
`)
}

func TestResolveValidateArgExtra(t *testing.T) {
	testResolve(t, "", `query {title(a: 3)}`, `{
  "data": {
  },
  "errors": [
    {
      "locations": [
        {
          "column": 14,
          "line": 1
        }
      ],
      "message": "validation: a is not an argument to title",
      "path": [
        "title"
      ]
    }
  ]
}
`)
}

func TestResolveValidateArgMissing(t *testing.T) {
	testResolve(t, "", `query {options}`, `{
  "data": {
    "options": null
  },
  "errors": [
    {
      "locations": [
        {
          "column": 9,
          "line": 1
        }
      ],
      "message": "resolve error: misc is required but missing",
      "path": [
        "options"
      ]
    }
  ]
}
`)
}

func TestResolveValidateFieldExists(t *testing.T) {
	testResolve(t, "", `query {bad}`, `{
  "data": {
  },
  "errors": [
    {
      "locations": [
        {
          "column": 9,
          "line": 1
        }
      ],
      "message": "resolve error: bad is not a field in Query",
      "path": [
        "bad"
      ]
    }
  ]
}
`)
}
