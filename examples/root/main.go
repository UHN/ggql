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

// Any is an empty struct used for attaching methods.
type Any struct {
}

// Resolve the field into the value stored in the obj.
func (ar *Any) Resolve(obj interface{}, field *ggql.Field, args map[string]interface{}) (result interface{}, err error) {
	if m, _ := obj.(map[string]interface{}); m != nil {
		result = m[field.Name]
	} else {
		err = fmt.Errorf("expected a map[string]interface{}, not a %T", obj)
	}
	return
}

// Len returns the length of the list.
func (ar *Any) Len(list interface{}) int {
	if tlist, ok := list.([]interface{}); ok {
		return len(tlist)
	}
	return 0
}

// Nth returns the nth element in a list.
func (ar *Any) Nth(list interface{}, i int) (result interface{}, err error) {
	if i < 0 {
		return 0, fmt.Errorf("index must be >= 0, not %d", i)
	}
	if tlist, ok := list.([]interface{}); ok {
		if len(tlist) <= i {
			return 0, fmt.Errorf("index must be less than the list length, %d > len %d", i, len(tlist))
		}
		return tlist[i], nil
	}
	return 0, fmt.Errorf("expected a []interface{}, not a %T", list)
}

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
