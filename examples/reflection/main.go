package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/uhn/ggql/pkg/ggql"
)

// Define the types and data in a generic way without considering the GraphQL
// package then hookup the package as needed. This section is copied into all
// golang framework applications.
////////////////////////////////////////////////////////////////////////////////
// Start of GraphQL package neutral type definitions.

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

// Schema represents the top level of a GraphQL data/resolver graph.
type Schema struct {
	Query    *Query
	Mutation *Mutation
}

// Query represents the query node in a data/resolver graph.
type Query struct {
	Artists []*Artist
}

// Artist returns the artist in the list with the specified name.
func (q *Query) Artist(name string) *Artist {
	for _, a := range q.Artists {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// Mutation represents the query node in a data/resolver graph.
type Mutation struct {
	query *Query // Query is the data store for this example
}

// Like increments likes attribute the song of the artist specified.
func (m *Mutation) Like(artist, song string) *Song {
	if a := m.query.Artist(artist); a != nil {
		if s := a.Song(song); s != nil {
			s.Likes++
			return s
		}
	}
	return nil
}

// Artist represents the GraphQL Artist.
type Artist struct {
	Name   string
	Songs  []*Song
	Origin []string
}

// Song returns the song in the list with the specified name.
func (a *Artist) Song(name string) *Song {
	for _, s := range a.Songs {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Song represents the GraphQL Song.
type Song struct {
	Name     string
	Artist   *Artist
	Duration int
	Release  *Date
	Likes    int
}

// Date represents a date with year, month, and day of the month.
type Date struct {
	Year  int
	Month int
	Day   int
}

// DateFromString parses a string in the format YYYY-MM-DD into a Date
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

// Now the framework specific code.

// DateScalar represents a Date scalar.
type DateScalar struct {
	ggql.Scalar
}

// NewDateScalar returns a new DateScalar as a ggql.Type.
func NewDateScalar() ggql.Type {
	return &DateScalar{ggql.Scalar{Base: ggql.Base{N: "Date"}}}
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

func handleGraphQL(w http.ResponseWriter, req *http.Request, root *ggql.Root) {
	var result map[string]interface{}
	// An example of using CORS headers to allow cross site access for
	// GraphiQL.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "172800")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
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
	case "OPTIONS":
		w.WriteHeader(200)
		return
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
