# :notes: An Instrumental Introduction to GraphQL with Go :notes:

| [Home](../../README.md) | [Examples](../README.md) |
| ----------------------- | ------------------------ |

You may have heard developers sing praises of GraphQL. In this example
we will go through an example application that uses GraphQL and the Go
GGql package. This example is derived from [An instrumental intro to
GraphQL with Ruby](https://blog.appsignal.com/2019/01/29/graphql.html)
with permission from the author (me).

## What is GraphQL

GraphQL is a query language and runtime that can be used to build
APIs. It holds a similar position in the development stack as a REST
API but with more flexibility. Unlike REST, GraphQL allows response
formats and content to be specified by the client. Just as SQL SELECT
statements allow query results to be specified, GraphQL allows
returned JSON data structures to be specified. Following the SQL
analogy, GraphQL does not provide a WHERE clause but identifies fields
on application objects that should provide the data for the response.

GraphQL, as the name suggests, models APIs as though the application
is a graph of data. While this description may not be how you view
your application, it is a model used in most systems. Data that can be
represented by JSON is a graph since JSON is just a directed
graph. Thinking about the application as presenting a graph model
through the API will make GraphQL much easier to understand.

## Using GraphQL in an Application

Now that we’ve described GraphQL in the abstract, let’s get down to
actually building an application that uses GraphQL by starting with a
definition of the data model or the graph. Last year I picked up a new
hobby. I’m learning to play the electric upright bass as well as
learning about music in general, so a music-related example came to
mind when coming up with a demo app.

The object types in the example are Artist and Song. Artists have
multiple Song and a Song is associated with an Artist. Each object
type has attributes such as a `name`.

## Define the API

GraphQL uses SDL (Schema Definition Language) which is sometimes
referred to as "type system definition language" in the GraphQL
specification. GraphQL types can, in theory, be defined in any
language but the most common agnostic language is SDL so let’s use SDL
to define the API.

```graphql
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

An **Artist** has a `name` that is a `String`, `songs` that is an
array of **Song** objects, and `origin` which is a `String`
array. **Song** is similar but with one odd field. The `release` field
should be a date type but GraphQL does not have that type defined as a
scalar in the specification. To be completely portable between any
GraphQL implementation a `String` is used. The GGql GraphQL
implementation allows scalar types to be added so that's what we will
do. The returned value will be a `String` but by setting the type to
`Date` we document the API more accurately.

The last step is to describe how to get one or more of the
objects. This is referred to as the root or for queries the query
root. Our root will have just two fields or methods called `artist`
and `artists`. The `artist` field will require an artist `name`
argument.

```graphql
type Query {
  artist(name: String!): Artist
  artists: [Artist]
}
```

## Writing the Application

Let’s look at how we would use this in an application. There are a few
implementations of GraphQL in Go. The most common one is
[graphql-go](https://github.com/graphql-go/graphql) which requires the
SDL above to be translated into a set of Go struct equivalents. GGql
takes a different approach. Three different approaches actually as
there are three options that can be used. The reflection approach is
used in this example.

To start, the types involved are modeled without concern as to the
GraphQL API. That's kind of the way it ought to be anyway, right? For
clarity the Go type names will match the GraphQL types.

```golang
type Artist struct {
	Name   string
	Songs  []*Song
	Origin []string
}

type Song struct {
	Name     string
	Artist   *Artist
	Duration int
	Release  *Date
	Likes    int
}

type Date struct {
	Year  int
	Month int
	Day   int
}
```

Next add some functions that will be needed to navigate and retreive
data from our model. The comments have been removed but they can be
viewed in the [source file](main.go).

```golang
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

```

## Where's the GraphQL?

With the model and basic behavior defined it is time to consider how
to put a GraphQL frontend on the model. First we need to define an
entry point for queries to match the Query type defined in our schema.

```golang
type Query struct {
	Artists ArtistList
}
```

The GraphQL root (not to be confused with the query root) sits above
the query root and is implemented with a Go **Schema** type. GraphQL
defines it to optionally have three fields. The Schema type in this
case implements the `query` and `mutation` fields. More about the
mutation a little later.

```golang
type Schema struct {
	Query    *Query
	Mutation *Mutation
}
```

With all the types defined it is possible to create sample data to
power the API. An instance of the Query type will hold the Artists and
Songs that form our in-memory store. An enhancement might be to read
all the information from your music library but that is a task for
another day. A function that creates the data and returns a top level
Schema instance handles the creation of a data graph.

```golang
func setupSongs() *Schema {
	fazerdaze := Artist{Name: "Fazerdaze", Origin: []string{"Morningside", "Auckland", "New Zealand"}}
	may5 := &Date{Year: 2017, Month: 5, Day: 5}
	nov2 := &Date{Year: 2015, Month: 11, Day: 2}
	fazerdaze.Songs = []*Song{
		{Name: "Jennifer", Artist: &fazerdaze, Duration: 240, Release: may5},
		{Name: "Lucky Girl", Artist: &fazerdaze, Duration: 170, Release: may5},
		{Name: "Friends", Artist: &fazerdaze, Duration: 194, Release: may5},
		{Name: "Reel", Artist: &fazerdaze, Duration: 193, Release: nov2},
	}

	boys := Artist{Name: "Viagra Boys", Origin: []string{"Stockholm", "Sweden"}}
	sep28 := &Date{Year: 2018, Month: 11, Day: 2}
	boys.Songs = []*Song{
		{Name: "Down In The Basement", Artist: &boys, Duration: 216, Release: sep28},
		{Name: "Frogstrap", Artist: &boys, Duration: 195, Release: sep28},
		{Name: "Worms", Artist: &boys, Duration: 208, Release: sep28},
		{Name: "Amphetanarchy", Artist: &boys, Duration: 346, Release: sep28},
	}

	query := Query{
		Artists: []*Artist{&fazerdaze, &boys},
	}
	return &Schema{
		Query:    &query,
		Mutation: &Mutation{query: &query},
	}
}
```

GraphQL is not just about querying data. An API often includes calls
to create and update the data graph. In GraphQL this is done with
mutations. Let's start by adding a GraphQL mutation description. Note
that according to the GraphQL specification all modifications to the
data graph are done only at the root mutation. It's a somewhat
arbitrary restriction but that's the nature of the beast. Let's start
with a simple mutation of "like", which mutates a song in our data
graph identified by an artist and song name.

```graphql
type Mutation {
  like(artist: String!, song: String!): Song
}
```

We hook up the mutation with a corresponding Go type named
`Mutation`. Give the Mutation access to the data through the Query
instance. Implement the `Like` function without any ties to a specific
GraphQL package.

```golang
type Mutation struct {
	query *Query
}

func (m *Mutation) Like(artist, song string) *Song {
	if a := m.query.Artist(artist); a != nil {
		if s := a.Song(song); s != nil {
			s.Likes++
			return s
		}
	}
	return nil
}
```

## GGql Hookup

What does it take to hook up a GraphQL API with GGql using the
reflection approach? Only a definition of the `Date` scalar
type. Everything else takes care of itself.

We have to implement the logic to coerce an input string into our Date
type, and coerce our Date into a string for outputting results.

```golang
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

func (t *DateScalar) CoerceIn(v interface{}) (interface{}, error) {
	if s, ok := v.(string); ok {
		return DateFromString(s)
	}
	return nil, fmt.Errorf("%w %v into a Date", ggql.ErrCoerce, v)
}

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
```

The final hook-up is done by creating a GGql Root. A Root parses a schema
SDL file and takes the data graph root. It is ready for executable
document evaluation.

```golang
func buildRoot() (root *ggql.Root, err error) {
	schema := setupSongs()
	ggql.Sort = true
	root = ggql.NewRoot(schema)
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

That concludes the GraphQL setup. Using this Root we can now
execute/handle GraphQL queries. How those queries arrive to your
system is up to you. You can use whatever transport you want to expose
your API (e.g. HTTP server). We're going to use an HTTP server.

A handler is needed for the Go `http` package. It must handle both
`GET` for URL queries and `POST` for mutations and more complex
queries. In each case the executable document is resolved by the root
and the result JSON encoded as the response.

```golang
func handleGraphQL(w http.ResponseWriter, req *http.Request, root *ggql.Root) {
	var result map[string]interface{}
	switch req.Method {
	case "GET":
		result = root.ResolveString(req.URL.Query().Get("query"), "", nil)
	case "POST":
		defer func() { _ = req.Body.Close() }()
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		result = root.ResolveBytes(body, "", nil)
	}
	indent := -1
	if i, err := strconv.Atoi(req.URL.Query().Get("indent")); err == nil {
		indent = i
	}
	_ = ggql.WriteJSONValue(w, result, indent)
}

func main() {
	root, err := buildRoot()
	if err != nil {
		fmt.Printf("*-*-* Failed to build schema. %s\n", err)
		os.Exit(1)
	}
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		handleGraphQL(w, r, root)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	if err = http.ListenAndServe(":3000", nil); err != nil {
		fmt.Printf("*-*-* Server failed. %s\n", err)
	}
}
```

## Using the API

To test the API, you can use a web browser, Postman, curl, or any
other client capable of making HTTP requests.

The GraphQL query to try looks like the following:

```
{
  artist(name:"Fazerdaze") {
    name
    songs{
      name
      duration
    }
  }
}
```

The query asks for the **Artist** named `Fazerdaze` and returns the
`name` and `songs` in a JSON document. For each **Song** the `name`
and `duration` of the **Song** is returned in a JSON object for each
**Song**. The output should look like this:

```
{
  "data":{
    "artist":{
      "name":"Fazerdaze",
      "songs":[
        {
          "name":"Jennifer",
          "duration":240
        },
        {
          "name":"Lucky Girl",
          "duration":170
        },
        {
          "name":"Friends",
          "duration":194
        },
        {
          "name":"Reel",
          "duration":193
        }
      ]
    }
  }
}
```

After getting rid of the optional whitespace in the query an HTTP GET made
with curl should return that content.

```bash
curl -w "\n" 'localhost:3000/graphql?query=\{artist(name:"Fazerdaze")\{name,songs\{name,duration\}\}\}&indent=2'
```

Try changing the query and replace `duration` with `release` and note the
conversion of the Date to a JSON string.

Hope you enjoyed the example. Now you can sing the GraphQL song.
