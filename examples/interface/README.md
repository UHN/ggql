# Interface Resolver

| [Home](../../README.md) | [Examples](../README.md) |
| ----------------------- | ------------------------ |

For some applications the easy to use reflection resolver is not
enough. While the reflection resolver takes less code it has some
restrictions that can be worked around with the interface resolver
approach. Both can be used together as well.

* The reflection resolver does not support implementation of the
  GraphQL type with an interface. So if `Artist` in the song schema
  was implemented by multiple Go types but with an `Artist` interface
  the reflection resolver could not be used.

* Non-slice collections are not supported with the reflection
  resolvers. The interface resolver does support arbitrary
  collections.

* The reflection resolver is more fragile in regard to developer
  errors. Of course those errors should be caught in testing but if a
  function such as the `Mutation.Like()` function were coded to take
  three values instead of two then a panic would occur when it was
  called. The Go reflection package does not provide the information
  needed to protect against invalid arguments.

The song schema does not need to change but the implementation
does. Instead of using `[]*Song` and `[]*Artist`, wrapper types are
used and the Go Artist and and Query types are modified with this
code.

```golang
type Query struct {
	Artists ArtistList
}

type Artist struct {
	Name   string
	Songs  SongList
	Origin []string
}

type ArtistList []*Artist

func (al ArtistList) Len() int {
	return len(al)
}

func (al ArtistList) Nth(i int) interface{} {
	return al[i]
}

func (al ArtistList) GetByName(name string) *Artist {
	for _, a := range al {
		if a.Name == name {
			return a
		}
	}
	return nil
}

type SongList []*Song

func (sl SongList) Len() int {
	return len(sl)
}

func (sl SongList) Nth(i int) interface{} {
	return sl[i]
}

func (sl SongList) GetByName(name string) *Song {
	for _, s := range sl {
		if s.Name == name {
			return s
		}
	}
	return nil
}
```

Note the use of `GetByName()` in the list instead of in the Artist and
Query types. Functions are changed accordingly.

```golang
func (q *Query) Artist(name string) *Artist {
	return q.Artists.GetByName(name)
}
```

The graph data setup has to change as well by replacing `[]*Song` with
`SongList` and `[]*Artist` with `ArtistList`.

## GGql Hookup

There are two interfaces that provide the glue for the interface
resolver. One is the `Resolver` for Object types and the other is
`ListResolver` for lists of types. As long as each data type
implements these interfaces the GraphQL API can resolve a query or
mutation into a result set which will be delivered as JSON.

### ListResolver Interface

Each of the lists implement the `ListResover` interface which provides
a means to get the length of the list and the nth element in the list.

```golang
type ListResolver interface {
	Len() int
	Nth(i int) interface{}
}
```

### Resolver Interface

The Resolver interface has one function, `Resolve`.

```golang
type Resolver interface {
	Resolve(field *Field, args map[string]interface{}) (interface{}, error)
}
```

The `field` provides the field name while the `args` have the function
arguments in a map. The function should return the child value for the
field or an error. Lets see what those look like.

```golang
func (s *Schema) Resolve(field *ggql.Field, _ map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "query":
		return s.Query, nil
	case "mutation":
	case "subscription":
	}
	return nil, fmt.Errorf("type Schema does not have field %s", field)
}

func (q *Query) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "title":
		return q.Title, nil
	case "artists":
		return q.Artists, nil
	case "artist":
		if name, _ := args["name"].(string); 0 < len(name) {
			return q.Artist(name), nil
		}
		return nil, fmt.Errorf("name argument not provided to field %s", field.Name)
	}
	return nil, fmt.Errorf("type Query does not have field %s", field)
}

func (m *Mutation) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	if field.Name == "like" {
		artist, _ := args["artist"].(string)
		song, _ := args["artist"].(string)
		return m.Like(artist, song), nil
	}
	return nil, fmt.Errorf("type Query does not have field %s", field)
}

func (a *Artist) Resolve(field *ggql.Field, _ map[string]interface{}) (interface{}, error) {
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

func (s *Song) Resolve(field *ggql.Field, _ map[string]interface{}) (interface{}, error) {
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
```

Each Resolve function just switches on the field name, extracts
arguments as needed, and returns some value.

The final hookup with the `buildRoot()` function and the HTTP server
remains the same as with the reflection hookup.

