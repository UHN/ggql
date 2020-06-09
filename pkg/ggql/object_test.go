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
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestObject(t *testing.T) {
	root := ggql.NewRoot(nil)
	genre := ggql.Enum{
		Base: ggql.Base{
			N: "Genre",
		},
	}
	genre.AddValue(&ggql.EnumValue{Value: "ROCK"})
	genre.AddValue(&ggql.EnumValue{Value: "CLASSICAL"})
	genre.AddValue(&ggql.EnumValue{Value: "JAZZ"})

	xDir := ggql.Directive{
		Base: ggql.Base{
			N: "example",
		},
		On: []ggql.Location{ggql.LocObject, ggql.LocFieldDefinition},
	}
	xDir.AddArg(&ggql.Arg{Base: ggql.Base{N: "type"}, Type: &ggql.NonNull{Base: root.GetType("String")}})

	like := ggql.Interface{Base: ggql.Base{N: "Likable"}, Root: root}
	like.AddField(&ggql.FieldDef{Base: ggql.Base{N: "likes"}, Type: root.GetType("Int")})

	hasSongs := ggql.Interface{Base: ggql.Base{N: "HasSongs"}, Root: root}
	hasSongs.AddField(&ggql.FieldDef{Base: ggql.Base{N: "songs"}, Type: &ggql.List{Base: root.GetType("String")}})

	artist := &ggql.Object{
		Base: ggql.Base{
			N:    "Artist",
			Desc: "Artist of songs",
			Dirs: []*ggql.DirectiveUse{
				{
					Directive: &xDir,
					Args: map[string]*ggql.ArgValue{
						"type": {
							Arg:   "type",
							Value: "github.com/fake/artist",
						},
					},
				},
			},
		},
		Interfaces: []ggql.Type{&like, &hasSongs},
	}
	fd := ggql.FieldDef{
		Base: ggql.Base{
			N: "genreSongs",
		},
		Type: &ggql.List{
			Base: root.GetType("String"),
		},
	}
	fd.AddArg(&ggql.Arg{Base: ggql.Base{N: "genre"}, Type: &genre, Default: "ROCK"})
	artist.AddField(&fd)

	artist.AddField(&ggql.FieldDef{
		Base: ggql.Base{
			N:    "songs",
			Desc: "The songs the artist released.",
		},
		Type: &ggql.List{
			Base: root.GetType("String"),
		},
	})
	artist.AddField(&ggql.FieldDef{
		Base: ggql.Base{
			N:    "likes",
			Desc: "The number of likes\nfor the song.",
			Dirs: []*ggql.DirectiveUse{{Directive: &xDir}},
		},
		Type: root.GetType("Int"),
	})

	err := root.AddTypes(artist, &xDir, &like, &hasSongs)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	actual := root.SDL(false, true)
	expectStr := `type Artist implements Likable & HasSongs @example(type: "github.com/fake/artist") {
  genreSongs(genre: Genre = "ROCK"): [String]
  songs: [String]
  likes: Int @example
}
`
	expectRoot := `
"Artist of songs"
type Artist implements Likable & HasSongs @example(type: "github.com/fake/artist") {
  genreSongs(genre: Genre = "ROCK"): [String]

  "The songs the artist released."
  songs: [String]

  """
  The number of likes
  for the song.
  """
  likes: Int @example
}

interface HasSongs {
  songs: [String]
}

interface Likable {
  likes: Int
}
` + timeScalarSDL + `
directive @example(type: String!) on OBJECT | FIELD_DEFINITION
`

	checkEqual(t, expectRoot, actual, "Object SDL() mismatch")
	checkEqual(t, expectStr, artist.String(), "Object String() mismatch")
	checkEqual(t, 5, artist.Rank(), "Object Rank() mismatch")
	checkEqual(t, "Artist of songs", artist.Description(), "Object Description() mismatch")
	checkEqual(t, 1, len(artist.Directives()), "Object Directives() mismatch")

	x := &ggql.Object{Base: ggql.Base{N: "Artist"},
		Interfaces: []ggql.Type{&ggql.Interface{Base: ggql.Base{N: "Other"}, Root: root}},
	}
	x.AddField(&ggql.FieldDef{Base: ggql.Base{N: "c"}, Type: root.GetType("Int")})
	err = artist.Extend(x)
	checkNil(t, err, "object.Extend failed. %s", err)

	x = &ggql.Object{Base: ggql.Base{N: "Artist"}}
	x.AddField(&ggql.FieldDef{Base: ggql.Base{N: "c"}, Type: root.GetType("Int")})
	err = artist.Extend(x)
	checkNotNil(t, err, "duplicate field for object.Extend should have failed.")

	err = artist.Extend(&ggql.Object{Base: ggql.Base{N: "Artist"},
		Interfaces: []ggql.Type{&ggql.Interface{Base: ggql.Base{N: "Other"}, Root: root}},
	})
	checkNotNil(t, err, "duplicate interface for object.Extend should have failed.")

	for i := 0; i < len(expectStr); i++ {
		w := &failWriter{max: i}
		err = artist.Write(w, false)
		checkNotNil(t, err, "return error on write error")
	}
}

func TestObjectFields(t *testing.T) {
	obj := ggql.Object{}
	checkEqual(t, 0, len(obj.Fields()), "Object Fields should be empty")
}
