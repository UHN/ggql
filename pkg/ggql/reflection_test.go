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

func testReflect(t *testing.T, sdl, src, expect string, pre func(r *ggql.Root)) {
	root := setupTestReflectSongs(t)

	if 0 < len(sdl) {
		err := root.ParseString(sdl)
		checkNil(t, err, "extend should not fail. %s", err)
	}
	if pre != nil {
		pre(root)
	}
	var b strings.Builder

	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, expect, b.String(), "result mismatch for %s", src)
}

func TestReflectSimple(t *testing.T) {
	src := `{title}`
	sdl := `extend type Query { name: String }`
	testReflect(t, sdl, src, expectTitle, nil)
}

func TestReflectFieldName(t *testing.T) {
	src := `{name}`
	sdl := `extend type Query { name: String }`
	expect := `{
  "data": {
    "name": "Songs"
  }
}
`
	pre := func(root *ggql.Root) {
		err := root.RegisterType(&RQuery{}, "Query")
		checkNil(t, err, "RegisterType should not fail. %s", err)
		err = root.RegisterField("Query", "name", "Title")
		checkNil(t, err, "RegisterField should not fail. %s", err)
	}
	testReflect(t, sdl, src, expect, pre)
}

func TestReflectMethod(t *testing.T) {
	src := `{artist(name:"Fazerdaze"){name}}`
	expect := `{
  "data": {
    "artist": {
      "name": "Fazerdaze"
    }
  }
}
`
	testReflect(t, "", src, expect, nil)
}

func TestReflectMethodWvar(t *testing.T) {
	src := `query Fazer($artist: String = "Fazerdaze"){artist(name:$artist){name}}`
	expect := `{
  "data": {
    "artist": {
      "name": "Fazerdaze"
    }
  }
}
`
	testReflect(t, "", src, expect, nil)
}

func TestReflectList(t *testing.T) {
	src := `{artists{name}}`
	testReflect(t, "", src, expectArtistsName, nil)
}

func TestReflectInline(t *testing.T) {
	src := `{artists{...on Artist {name} ...on Song {likes}}}`
	testReflect(t, "", src, expectArtistsName, nil)
}

func TestReflectFragment(t *testing.T) {
	src := `{all{...Arty}}
fragment Arty on Artist {name}
`
	pre := func(root *ggql.Root) {
		err := root.RegisterType(&RArtist{}, "Artist")
		checkNil(t, err, "RegisterType should not fail. %s", err)
		err = root.RegisterType(&RSong{}, "Song")
		checkNil(t, err, "RegisterType should not fail. %s", err)
	}
	expect := `{
  "data": {
    "all": [
      {
        "name": "Fazerdaze"
      },
      {
      },
      {
      },
      {
      },
      {
      },
      {
        "name": "Viagra Boys"
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
}
`
	testReflect(t, "", src, expect, pre)
}

func TestReflectInlineUnion(t *testing.T) {
	src := `{all{...on Artist {name} ...on Song {name,likes}}}`
	pre := func(root *ggql.Root) {
		err := root.RegisterType(&RArtist{}, "Artist")
		checkNil(t, err, "RegisterType should not fail. %s", err)
		err = root.RegisterType(&RSong{}, "Song")
		checkNil(t, err, "RegisterType should not fail. %s", err)
	}
	expect := `{
  "data": {
    "all": [
      {
        "name": "Fazerdaze"
      },
      {
        "likes": 0,
        "name": "Jennifer"
      },
      {
        "likes": 0,
        "name": "Lucky Girl"
      },
      {
        "likes": 0,
        "name": "Friends"
      },
      {
        "likes": 0,
        "name": "Reel"
      },
      {
        "name": "Viagra Boys"
      },
      {
        "likes": 0,
        "name": "Down In The Basement"
      },
      {
        "likes": 0,
        "name": "Frogstrap"
      },
      {
        "likes": 0,
        "name": "Worms"
      },
      {
        "likes": 0,
        "name": "Amphetanarchy"
      }
    ]
  }
}
`
	testReflect(t, "", src, expect, pre)
}

func TestReflectInterface(t *testing.T) {
	src := `{named{name}}`
	sdl := `
interface Namely {name: String!}
extend type Query {named: [Namely]}
extend type Artist implements Namely {}
extend type Song implements Namely {}
`
	pre := func(root *ggql.Root) {
		err := root.RegisterType(&RArtist{}, "Artist")
		checkNil(t, err, "RegisterType should not fail. %s", err)
		err = root.RegisterType(&RSong{}, "Song")
		checkNil(t, err, "RegisterType should not fail. %s", err)
	}
	expect := `{
  "data": {
    "named": [
      {
        "name": "Fazerdaze"
      },
      {
        "name": "Jennifer"
      },
      {
        "name": "Lucky Girl"
      },
      {
        "name": "Friends"
      },
      {
        "name": "Reel"
      },
      {
        "name": "Viagra Boys"
      },
      {
        "name": "Down In The Basement"
      },
      {
        "name": "Frogstrap"
      },
      {
        "name": "Worms"
      },
      {
        "name": "Amphetanarchy"
      }
    ]
  }
}
`
	testReflect(t, sdl, src, expect, pre)
}

func TestReflectArgOrder(t *testing.T) {
	src := `{song(song: "Jennifer", artist: "Fazerdaze"){name}}`
	expect := `{
  "data": {
    "song": {
      "name": "Jennifer"
    }
  }
}
`
	pre := func(root *ggql.Root) {
		err := root.RegisterType(&RQuery{}, "Query")
		checkNil(t, err, "RegisterType should not fail. %s", err)
		err = root.RegisterField("Query", "song", "Song")
		checkNil(t, err, "RegisterField should not fail. %s", err)
	}
	testReflect(t, "", src, expect, pre)
}

func TestReflectRegArgOrder(t *testing.T) {
	src := `{song(song: "Jennifer", artist: "Fazerdaze"){name}}`
	expect := `{
  "data": {
    "song": {
      "name": "Jennifer"
    }
  }
}
`
	pre := func(root *ggql.Root) {
		err := root.RegisterType(&RQuery{}, "Query")
		checkNil(t, err, "RegisterType should not fail. %s", err)
		err = root.RegisterField("Query", "song", "Song", "artist", "song")
		checkNil(t, err, "RegisterField should not fail. %s", err)
	}
	testReflect(t, "", src, expect, pre)
}

func TestReflectRegisterType(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterType(&RQuery{}, "")
	checkNil(t, err, "RegisterType with a blank name should not fail. %s", err)
}

func TestReflectRegisterNonObjectType(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterType(&RQuery{}, "Date")
	checkNotNil(t, err, "RegisterType with a non-object should fail")
}

func TestReflectRegisterTypeDup(t *testing.T) {
	root := setupTestReflectSongs(t)

	_ = root.RegisterType(&RArtist{}, "Artist")
	err := root.RegisterType(&RSong{}, "Artist")
	checkNotNil(t, err, "RegisterType twice with different types should fail")
}

func TestReflectRegisterNonObjectField(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterField("Date", "name", "Name")
	checkNotNil(t, err, "RegisterField with a non-object should fail")
}

func TestReflectRegisterNotRegField(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterField("Artist", "name", "Name")
	checkNotNil(t, err, "RegisterField with a not registered type should fail")
}

func TestReflectRegisterBadField(t *testing.T) {
	root := setupTestReflectSongs(t)

	_ = root.RegisterType(&RArtist{}, "Artist")
	err := root.RegisterField("Artist", "nonsense", "Name")
	checkNotNil(t, err, "RegisterField with an invalid field should fail")
}

func TestReflectRegisterWrongArgField(t *testing.T) {
	root := setupTestReflectSongs(t)

	_ = root.RegisterType(&RArtist{}, "Artist")
	err := root.RegisterField("Artist", "name", "Name", "bad")
	checkNotNil(t, err, "RegisterField with wrong args should fail")
}

func TestReflectRegisterArgFieldLenError(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterType(&RQuery{}, "Query")
	checkNil(t, err, "RegisterType should not fail. %s", err)
	err = root.RegisterField("Query", "song", "Song", "artist")
	checkNotNil(t, err, "RegisterField with not enough argument should fail")
}

func TestReflectRegisterArgFieldWrong(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterType(&RQuery{}, "Query")
	checkNil(t, err, "RegisterType should not fail. %s", err)
	err = root.RegisterField("Query", "song", "Song", "artist", "bad")
	checkNotNil(t, err, "RegisterField with wrong argument should fail")
}

func TestReflectRegisterBadGoField(t *testing.T) {
	root := setupTestReflectSongs(t)

	err := root.RegisterType(&RQuery{}, "Query")
	checkNil(t, err, "RegisterType should not fail. %s", err)
	err = root.RegisterField("Query", "song", "Sing")
	checkNotNil(t, err, "RegisterField with wrong argument should fail")
}

func TestReflectRegMissingField(t *testing.T) {
	src := `{bad}`
	sdl := `extend type Query { bad: String }`
	expect := `{
  "data": {
    "bad": null
  },
  "errors": [
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: reflection error: bad is not a field of *ggql_test.RQuery",
      "path": [
        "bad"
      ]
    }
  ]
}
`
	testReflect(t, sdl, src, expect, nil)
}

func TestReflectMultiReturn(t *testing.T) {
	src := `{artistCount}`
	sdl := `extend type Query { artistCount: Int }`
	expect := `{
  "data": {
    "artistCount": 2
  }
}
`
	testReflect(t, sdl, src, expect, nil)
}

func TestReflectMultiReturnError(t *testing.T) {
	src := `{artistCountErr}`
	sdl := `extend type Query { artistCountErr: Int }`
	expect := `{
  "data": {
    "artistCountErr": 2
  },
  "errors": [
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: dummy error",
      "path": [
        "artistCountErr"
      ]
    }
  ]
}
`
	testReflect(t, sdl, src, expect, nil)
}

func TestReflect3Return(t *testing.T) {
	src := `{three}`
	sdl := `extend type Query { three: Int }`
	expect := `{
  "data": {
    "three": null
  },
  "errors": [
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: *ggql_test.RQuery.three returned more than 2 values",
      "path": [
        "three"
      ]
    }
  ]
}
`
	testReflect(t, sdl, src, expect, nil)
}

func TestReflectNotList(t *testing.T) {
	src := `{notList{name}}`
	sdl := `extend type Query { notList: [Int] }`
	expect := `{
  "data": {
    "notList": null
  },
  "errors": [
    {
      "locations": [
        {
          "column": 3,
          "line": 1
        }
      ],
      "message": "resolve error: int is not a list type",
      "path": [
        "notList"
      ]
    }
  ]
}
`
	testReflect(t, sdl, src, expect, nil)
}
