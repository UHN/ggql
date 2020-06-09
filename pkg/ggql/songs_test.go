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
	"strconv"
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

const songsSdl = `
type Query {
  title: String
  artist(name: String!): Artist
  artists: [Artist]
  options(misc: Misc!): String
  byID(id: ID!): Artist
}

type Mutation {
  like(artist: String!, song: String!): Song
}

type Subscription {
  like(artist: String): Song!
}

"Song player or singer"
type Artist @go(type: "ggql_test.Artist") {
  name: String!
  songs: [Song]
  origin: [String]
}

type Song {
  name: String!
  artist: Artist
  duration: Int
  release: Date
  likes: Int
}

input Misc {
  width: Int!
  sizes: [Int]
}

scalar Date

directive @example on VARIABLE_DEFINITION
`

////////////////////////////////////////////////////////////////////////////////
// Start of GraphQL package neutral type definitions.

func setupSongs() *Schema {
	fazerdaze := Artist{Name: "Fazerdaze", Origin: []string{"Morningside", "Auckland", "New Zealand"}}
	may5 := &Date{Year: 2017, Month: 5, Day: 5}
	nov2 := &Date{Year: 2015, Month: 11, Day: 2}
	fazerdaze.Songs = SongList{
		{Name: "Jennifer", Artist: &fazerdaze, Duration: 240, Release: may5},
		{Name: "Lucky Girl", Artist: &fazerdaze, Duration: 170, Release: may5},
		{Name: "Friends", Artist: &fazerdaze, Duration: 194, Release: may5},
		{Name: "Reel", Artist: &fazerdaze, Duration: 193, Release: nov2},
	}

	boys := Artist{Name: "Viagra Boys", Origin: []string{"Stockholm", "Sweden"}}
	sep28 := &Date{Year: 2018, Month: 11, Day: 2}
	boys.Songs = SongList{
		{Name: "Down In The Basement", Artist: &boys, Duration: 216, Release: sep28},
		{Name: "Frogstrap", Artist: &boys, Duration: 195, Release: sep28},
		{Name: "Worms", Artist: &boys, Duration: 208, Release: sep28},
		{Name: "Amphetanarchy", Artist: &boys, Duration: 346, Release: sep28},
	}

	query := Query{
		Title:   "Songs",
		Artists: ArtistList{&fazerdaze, &boys},
	}
	return &Schema{
		Query:        &query,
		Mutation:     &Mutation{query: &query},
		Subscription: &Subscription{},
	}
}

// Schema represents the top level of a GraphQL data/resolver graph.
type Schema struct {
	Query        *Query
	Mutation     *Mutation
	Subscription *Subscription
}

// Query represents the query node in a data/resolver graph.
type Query struct {
	Title   string
	Artists ArtistList
}

// Mutation represents the mutation node in a data/resolver graph.
type Mutation struct {
	query *Query // Query is the data store for this example
}

// Subscription represents the subscription node in a data/resolver graph.
type Subscription struct {
	log *strings.Builder
}

// Likes increments likes attribute the song of the artist specified.
func (m *Mutation) Likes(artist, song string) *Song {
	if a := m.query.Artists.GetByName(artist); a != nil {
		if s := a.Songs.GetByName(song); s != nil {
			s.Likes++
			return s
		}
	}
	return nil
}

// Artist returns the artist in the list with the specified name.
func (q *Query) Artist(name string) *Artist {
	return q.Artists.GetByName(name)
}

func (q *Query) ByID(id string) *Artist {
	return q.Artists.GetByName(id)
}

func (q *Query) All() (all []interface{}) {
	for _, a := range q.Artists {
		all = append(all, a)
		for _, s := range a.Songs {
			all = append(all, s)
		}
	}
	return
}

// Artist represents the GraphQL Artist.
type Artist struct {
	Name   string
	Songs  SongList
	Origin []string
}

// Song represents the GraphQL Song.
type Song struct {
	Name     string
	Artist   *Artist
	Duration int
	Release  *Date
	Likes    int
}

// ArtistList is a list of Artists. It exists to allow list members to be
// ordered but still implement map like behavior (not yet implemented).
type ArtistList []*Artist

// Len of the list.
func (al ArtistList) Len() int {
	return len(al)
}

// Nth element in the list.
func (al ArtistList) Nth(i int) interface{} {
	return al[i]
}

// GetByName retrieves the element with the specified name.
func (al ArtistList) GetByName(name string) *Artist {
	for _, a := range al {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// SongList is a list of Songs. It exists to allow list members to be
// ordered but still implement map like behavionr (not yet implemented).
type SongList []*Song

// Len of the list.
func (sl SongList) Len() int {
	return len(sl)
}

// Nth element in the list.
func (sl SongList) Nth(i int) interface{} {
	return sl[i]
}

// GetByName retrieves the element with the specified name.
func (sl SongList) GetByName(name string) *Song {
	for _, s := range sl {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Date represents a date with year, month, and day of the month.
type Date struct {
	Year  int
	Month int
	Day   int
}

// DateFromString parses a string in the format YYY-MM-DD into a Date
// instance.
func DateFromString(s string) (d *Date, err error) {
	d = &Date{}
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%s is not a valid date format", s)
	}
	if d.Year, err = strconv.Atoi(parts[0]); err == nil {
		if d.Month, err = strconv.Atoi(parts[1]); err == nil {
			d.Day, err = strconv.Atoi(parts[2])
		}
	}
	return
}

// String returns a YYYY-MM-DD formatted string representation of the Date.
func (d *Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// End of GraphQL package neutral type definitions.
////////////////////////////////////////////////////////////////////////////////

func setupTestSongs(t *testing.T, log *strings.Builder) *ggql.Root {
	schema := setupSongs()

	ggql.Sort = true
	root := ggql.NewRoot(schema)
	schema.Subscription.log = log
	err := root.AddTypes(NewDateScalar())
	checkNil(t, err, "no error should be returned when adding a Date type. %s", err)

	err = root.ParseString(songsSdl)
	checkNil(t, err, "no error should be returned when parsing a valid SDL. %s", err)

	return root
}

func (s *Schema) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "query":
		return s.Query, nil
	case "mutation":
		//return s.Mutation, nil
	case "subscription":
		return s.Subscription, nil
	}
	return nil, fmt.Errorf("type Schema does not have field %s", field)
}

func (q *Query) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "title":
		return q.Title, nil
	case "artists":
		return q.Artists, nil
	case "all":
		return q.All(), nil
	case "artist":
		if name, _ := args["name"].(string); 0 < len(name) {
			return q.Artist(name), nil
		}
		return nil, fmt.Errorf("name argument not provided to field %s", field.Name)
	case "byID":
		if id, _ := args["id"].(string); 0 < len(id) {
			return q.Artist(id), nil
		}
		return nil, fmt.Errorf("id argument not provided to field %s", field.Name)
	case "options":
		var buf strings.Builder

		_ = ggql.WriteSDLValue(&buf, args)

		return buf.String(), nil
	}
	return nil, fmt.Errorf("type Query does not have field %s", field)
}

func (a *Artist) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "name":
		return a.Name, nil
	case "songs":
		return a.Songs, nil
	case "origin":
		return a.Origin, nil
	}
	return nil, fmt.Errorf("type Artist does not have field %s", field)
}

func (s *Song) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "name":
		return s.Name, nil
	case "artist":
		return s.Artist, nil
	case "duration":
		return s.Duration, nil
	case "release":
		return s.Release, nil
	case "likes":
		return s.Likes, nil
	}
	return nil, fmt.Errorf("type Song does not have field %s", field)
}

type DateScalar struct {
	ggql.Scalar
}

func NewDateScalar() ggql.Type {
	return &DateScalar{
		ggql.Scalar{
			Base: ggql.Base{
				N: "Date",
			},
		},
	}
}

// CoerceIn coerces an input value into the expected input type if possible
// otherwise an error is returned.
func (t *DateScalar) CoerceIn(v interface{}) (interface{}, error) {
	if s, ok := v.(string); ok {
		return DateFromString(s)
	}
	return nil, fmt.Errorf("%w %v into a Date", ggql.ErrCoerce, v)
}

// CoerceOut coerces a result value into a type for the scalar.
func (t *DateScalar) CoerceOut(v interface{}) (interface{}, error) {
	var err error
	switch tv := v.(type) {
	case string:
		// ok as is
	case *Date:
		v = tv.String()
	default:
		return nil, fmt.Errorf("%w %v into a Date", ggql.ErrCoerce, v)
	}
	return v, err
}
