# Root Resolver

| [Home](../../README.md) | [Examples](../README.md) |
| ----------------------- | ------------------------ |

When the data is relatively unstructured or only known at run time the
root resolver approach can be used. This approach uses a single
resolver function to resolve all requests. Only GraphQL Query operations are supported
with the root resolver. Arguments are supported on fields, but it does
take more effort than demonstrated in this example.

The song schema has been simplified by eliminating the mutation and by
removing fields that took arguments.

```graphql
type Query {
  artists: [Artist]
}

type Artist {
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

scalar Date
```

All data is stored as a tree so the setup becomes:

```golang
func setupSongs() interface{} {
    may5 := &Date{Year: 2017, Month: 5, Day: 5}
    nov2 := &Date{Year: 2015, Month: 11, Day: 2}
    sep28 := &Date{Year: 2018, Month: 11, Day: 2}
    return map[string]interface{}{
        "query": map[string]interface{}{
            "title": "Songs",
            "artists": []interface{}{
                map[string]interface{}{
                    "name":   "Fazerdaze",
                    "origin": []string{"Morningside", "Auckland", "New Zealand"},
                    "songs": []interface{}{
                        map[string]interface{}{"name": "Jennifer", "duration": 240, "release": may5},
                        map[string]interface{}{"name": "Lucky Girl", "duration": 170, "release": may5},
                        map[string]interface{}{"name": "Friends", "duration": 194, "release": may5},
                        map[string]interface{}{"name": "Reel", "duration": 193, "release": nov2},
                    },
                },
                map[string]interface{}{
                    "name":   "Viagra Boys",
                    "origin": []string{"Stockholm", "Sweden"},
                    "songs": []interface{}{
                        map[string]interface{}{"name": "Down In The Basement", "duration": 216, "release": sep28},
                        map[string]interface{}{"name": "Frogstrap", "duration": 195, "release": sep28},
                        map[string]interface{}{"name": "Worms", "duration": 208, "release": sep28},
                        map[string]interface{}{"name": "Amphetanarchy", "duration": 346, "release": sep28},
                    },
                },
            },
        },
    }
}
```

The root resolver is attached to the `ggql.Root` and must implement the `AnyResolver` interface.

```golang
type AnyResolver interface {
	Resolve(obj interface{}, field *Field, args map[string]interface{}) (interface{}, error)
	Len(list interface{}) int
	Nth(list interface{}, i int) (interface{}, error)
}
```

Define a type to use for the root resolver:

```golang
type Any struct {
}

func (ar *Any) Resolve(obj interface{}, field *ggql.Field, args map[string]interface{}) (result interface{}, err error) {
	if m, _ := obj.(map[string]interface{}); m != nil {
		result = m[field.Name]
	} else {
		err = fmt.Errorf("expected a map[string]interface{}, not a %T", obj)
	}
	return
}

func (ar *Any) Len(list interface{}) int {
	switch tlist := list.(type) {
	case []interface{}:
		return len(tlist)
	}
	return 0
}

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
	}
	return 0, fmt.Errorf("expected a []interface{}, not a %T", list)
}
```

Hookup is done by adding an instance of our resolver to the root as
seen in the modified `buildRoot()` function:

```golang
func buildRoot() (root *ggql.Root, err error) {
	schema := setupSongs()
	ggql.Sort = true
	root = ggql.NewRoot(schema)
	root.AnyResolver = &Any{}
	if err = root.AddTypes(NewDateScalar()); err != nil {
		return
	}
	var sdl []byte
	if sdl, err = ioutil.ReadFile("song.graphql"); err == nil {
		err = root.Parse(sdl)
	}
	return
}
```

The HTTP server setup remains the same as the other examples except a
specific artist lookup is not included so a more encompassing query
should be used instead.

```bash
curl -w "\n" 'localhost:3000/graphql?query=\{artists\{name,songs\{name,duration\}\}\}&indent=2'
```
