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
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

// -----------------------------------------------------------------------------
// Start of GraphQL package neutral type definitions.

func setupReflectSongs() *RSchema {
	fazerdaze := RArtist{Name: "Fazerdaze", Origin: []string{"Morningside", "Auckland", "New Zealand"}}
	may5 := &Date{Year: 2017, Month: 5, Day: 5}
	nov2 := &Date{Year: 2015, Month: 11, Day: 2}
	fazerdaze.Songs = RSongList{
		{Name: "Jennifer", Artist: &fazerdaze, Duration: 240, Release: may5},
		{Name: "Lucky Girl", Artist: &fazerdaze, Duration: 170, Release: may5},
		{Name: "Friends", Artist: &fazerdaze, Duration: 194, Release: may5},
		{Name: "Reel", Artist: &fazerdaze, Duration: 193, Release: nov2},
	}

	boys := RArtist{Name: "Viagra Boys", Origin: []string{"Stockholm", "Sweden"}}
	sep28 := &Date{Year: 2018, Month: 11, Day: 2}
	boys.Songs = RSongList{
		{Name: "Down In The Basement", Artist: &boys, Duration: 216, Release: sep28},
		{Name: "Frogstrap", Artist: &boys, Duration: 195, Release: sep28},
		{Name: "Worms", Artist: &boys, Duration: 208, Release: sep28},
		{Name: "Amphetanarchy", Artist: &boys, Duration: 346, Release: sep28},
	}

	query := RQuery{
		Title:   "Songs",
		Artists: RArtistList{&fazerdaze, &boys},
	}
	return &RSchema{
		Query:    &query,
		Mutation: &RMutation{query: &query},
	}
}

// RSchema represents the top level of a GraphQL data/resolver graph.
type RSchema struct {
	Query    *RQuery
	Mutation *RMutation
}

// Query represents the query node in a data/resolver graph.
type RQuery struct {
	Title   string
	Artists RArtistList
}

// RMutation represents the mutation node in a data/resolver graph.
type RMutation struct {
	query *RQuery // Query is the data store for this example
}

// Likes increments likes attribute the song of the artist specified.
func (m *RMutation) Likes(artist, song string) *RSong {
	if a := m.query.Artists.GetByName(artist); a != nil {
		if s := a.Songs.GetByName(song); s != nil {
			s.Likes++
			return s
		}
	}
	return nil
}

// Artist returns the artist in the list with the specified name.
func (q *RQuery) Artist(name string) *RArtist {
	return q.Artists.GetByName(name)
}

// All returns all artists and songs.
func (q *RQuery) All() (all []interface{}) {
	for _, a := range q.Artists {
		all = append(all, a)
		for _, s := range a.Songs {
			all = append(all, s)
		}
	}
	return
}

// Named returns all artists and songs.
func (q *RQuery) Named() (all []interface{}) {
	for _, a := range q.Artists {
		all = append(all, a)
		for _, s := range a.Songs {
			all = append(all, s)
		}
	}
	return
}

// Song searches for and returns a song.
func (q *RQuery) Song(artist, song string) *RSong {
	if a := q.Artists.GetByName(artist); a != nil {
		if s := a.Songs.GetByName(song); s != nil {
			return s
		}
	}
	return nil
}

// ArtistCount searches for and returns a song.
func (q *RQuery) ArtistCount() (int, error) {
	return len(q.Artists), nil
}

// ArtistCountErr searches for and returns a song.
func (q *RQuery) ArtistCountErr() (int, error) {
	return len(q.Artists), fmt.Errorf("dummy error")
}

// Three searches for and returns a song.
func (q *RQuery) Three() (int, int, error) {
	return len(q.Artists), 1, nil
}

// NotList searches for and returns a song.
func (q *RQuery) NotList() int {
	return len(q.Artists)
}

// RArtist represents the GraphQL Artist.
type RArtist struct {
	Name   string
	Songs  RSongList
	Origin []string
}

// RSong represents the GraphQL Song.
type RSong struct {
	Name     string
	Artist   *RArtist
	Duration int
	Release  *Date
	Likes    int
}

// RArtistList is a list of Artists. It exists to allow list members to be
// ordered but still implement map like behavior (not yet implemented).
type RArtistList []*RArtist

// RGetByName retrieves the element with the specified name.
func (al RArtistList) GetByName(name string) *RArtist {
	for _, a := range al {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// RSongList is a list of Songs. It exists to allow list members to be
// ordered but still implement map like behavionr (not yet implemented).
type RSongList []*RSong

// GetByName retrieves the element with the specified name.
func (sl RSongList) GetByName(name string) *RSong {
	for _, s := range sl {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// End of GraphQL package neutral type definitions.
// -----------------------------------------------------------------------------

func setupTestReflectSongs(t *testing.T) *ggql.Root {
	schema := setupReflectSongs()

	ggql.Sort = true
	root := ggql.NewRoot(schema)
	err := root.AddTypes(NewDateScalar())
	checkNil(t, err, "no error should be returned when adding a Date type. %s", err)

	err = root.ParseString(songsSdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	err = root.ParseString(`
union Both = Artist | Song
extend type Query { all: [Both] song(artist: String! song: String!): Song }
`)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	return root
}
