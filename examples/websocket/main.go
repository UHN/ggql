package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/uhn/ggql/pkg/ggql"
)

const marketSDL = `
type Query {
  price: Float
}

type Mutation {
  setPrice(price: Float!): Float
}

type Subscription {
  listenPrice: Float
}
`

var price = 1.1

// Schema represents the top level of a GraphQL data/resolver graph.
type Schema struct {
	Query        Query
	Mutation     Mutation
	Subscription Subscription
}

// Resolve fields.
func (s *Schema) Resolve(field *ggql.Field, _ map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "query":
		return &s.Query, nil
	case "mutation":
		return &s.Mutation, nil
	case "subscription":
		return &s.Subscription, nil
	}
	return nil, fmt.Errorf("type Schema does not have field %s", field)
}

// Query represents the query node in a data/resolver graph.
type Query struct {
}

// Price returns the price.
func (q *Query) Price() float64 {
	return price
}

// Mutation represents the mutation node in a data/resolver graph.
type Mutation struct {
	root *ggql.Root
}

// SetPrice sets the price.
func (m *Mutation) SetPrice(p float64) (float64, error) {
	price = p
	_, err := m.root.AddEvent("price", price)
	return price, err
}

// Subscription represents the subscription node in a data/resolver graph.
type Subscription struct {
	root *ggql.Root
}

// Resolve fields. The interface approach is needed for subscription as the
// field argument is needed to set up the subscription.
func (s *Subscription) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "listenPrice":
		if ws, _ := field.Context.(*WebSoc); ws != nil {
			ws.sub = ggql.NewSubscription(ws, field, args)
			return ws.sub, nil
		}
		return nil, fmt.Errorf("listenPrice subscription expected an upgradeable connection")
	default:
		return nil, fmt.Errorf("type Schema does not have field %s", field)
	}
}

func prepExe(exe *ggql.Executable, ws *WebSoc) {
	for _, op := range exe.Ops {
		for _, sel := range op.SelectionSet() {
			prepSelection(sel, ws)
		}
	}
}

func prepSelection(selection ggql.Selection, ws *WebSoc) {
	if field, _ := selection.(*ggql.Field); field != nil {
		field.Context = ws
	}
	for _, sel := range selection.SelectionSet() {
		prepSelection(sel, ws)
	}
}

func handleGraphQL(w http.ResponseWriter, req *http.Request, root *ggql.Root) {
	var result map[string]interface{}
	var exe *ggql.Executable
	var err error

	op := req.URL.Query().Get("operationName")
	vars := map[string]interface{}{}
	if jvars := req.URL.Query().Get("variables"); 0 < len(jvars) {
		err = json.Unmarshal([]byte(jvars), &vars)
	}
	// Since the connection has to be hijacked and handed over to the GraphQL
	// resolver a bit more code is required. First an ggql.Executable is
	// created. That executable is then prepared if the connection is
	// upgradeable. Finally the executable is evaluated.
	if err == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "172800")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		switch req.Method {
		case "GET":
			exe, err = root.ParseExecutableReader(strings.NewReader(req.URL.Query().Get("query")))
		case "POST":
			defer func() { _ = req.Body.Close() }()
			var contentType string
			if cta := req.Header["Content-Type"]; 0 < len(cta) {
				contentType = cta[0]
			}
			switch contentType {
			case "application/graphql":
				exe, err = root.ParseExecutableReader(req.Body)
			case "application/json":
				var jmap map[string]interface{}
				var data []byte
				if data, err = ioutil.ReadAll(req.Body); err == nil {
					err = json.Unmarshal(data, &jmap)
				}
				if err == nil {
					if str, _ := jmap["operationName"].(string); 0 < len(str) {
						op = str
					}
					vm, _ := jmap["variables"].(map[string]interface{})
					for k, v := range vm {
						if vars[k] == nil {
							vars[k] = v
						}
					}
					if str, _ := jmap["query"].(string); 0 < len(str) {
						exe, err = root.ParseExecutableString(str)
					}
				}
			default:
				err = fmt.Errorf("%s is not a supported Content-Type", contentType)
			}
		case "OPTIONS":
			return
		default:
			err = fmt.Errorf("%s is not a supported method", req.Method)
		}
	}
	hijacked := false
	if err == nil && req.Header.Get("Upgrade") == "WebSocket" {
		if jack, _ := w.(http.Hijacker); jack != nil {
			var ws *WebSoc
			if ws, err = NewWebSoc(jack); err == nil {
				prepExe(exe, ws)
				hijacked = true
			}
		}
	}
	if err == nil {
		if result, err = root.ResolveExecutable(exe, op, vars); result == nil {
			result = map[string]interface{}{"data": nil}
		}
	}
	if hijacked {
		return
	}
	if err != nil {
		if result == nil {
			result = map[string]interface{}{
				"errors": ggql.FormErrorsResult(err),
			}
		} else {
			result["errors"] = ggql.FormErrorsResult(err)
		}
	}
	indent := -1
	if i, err := strconv.Atoi(req.URL.Query().Get("indent")); err == nil {
		indent = i
	}
	_ = ggql.WriteJSONValue(w, result, indent)
}

func main() {
	ggql.Sort = true
	schema := &Schema{}
	root := ggql.NewRoot(schema)
	schema.Mutation.root = root
	schema.Subscription.root = root
	if err := root.ParseString(marketSDL); err != nil {
		fmt.Printf("*-*-* Failed to build schema. %s\n", err)
		os.Exit(1)
	}
	http.HandleFunc("/graphql/schema", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		full := strings.EqualFold(q.Get("full"), "true")
		desc := strings.EqualFold(q.Get("desc"), "true")
		sdl := root.SDL(full, desc)
		_, _ = w.Write([]byte(sdl))
	})
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		handleGraphQL(w, r, root)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/price.html", func(w http.ResponseWriter, r *http.Request) {
		content, _ := ioutil.ReadFile("price.html")
		_, _ = w.Write(content)
	})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Printf("*-*-* Server failed. %s\n", err)
	}
}
