# WebSockets with GraphQL in Go the Easy Way

| [Home](../../README.md) |
| ----------------------- |

Pushing data to a browser makes for lively up-to-date page. GraphQL
provides many advantages as well but we will not get into that
here. This example brings together WebSockets and GraphQL through the
use of GraphQL subscriptions. This example is written in Go using just
the [GGql](https://github.com/UHN/ggql) package.

This example assumes familiarity with the GGql package. Simpler, first
time examples for using GGql are:

 - [Examples](../README.md)
   - [Reflection](../reflection/README.md)
   - [Interface](../interface/README.md)
   - [Root](../root/README.md)

## Define the API

As always, a GraphQL application needs a GraphQL schema. A very simple
API is provided that allows getting, setting, and subscribing to
changes in a price. The price is just a number and nothing more.

```graphql
type Query {
  price: Float
}

type Mutation {
  setPrice(price: Float!): Float
}

type Subsciption {
  listenPrice: Float
}
```

## Use Case

The players involved are the go application (of course), a browser for
viewing updated prices, and a terminal for changing the price. The
scenario is just 3 steps.

 1. Start the application.
 2. Open the browser for URL: http://localhost:3000/price.html
 3. Curl in a price change with a mutation.

The price change is then pushed to the browser and the new price
displayed. Multiple browsers and multiple mutation calls are
supported.

In this case a single float is the data being pushed. It could just as
easily be a more complex object.

## The Code

For a WebSocket demonstration there need to be a client and a
server. The client is a web browser that opens an HTML page with
embedded Javascript. The server is in go.

### HTML and Javascript

The page shows a connection status and the price. On loading the page
makes a GraphQL request with a request that asks to upgrade the
connection to WebSockets.

On a successful connection the `onopen` function is called. The
GraphQL query is then sent to the server. In this case just the query
is sent. Many servers might expect a JSON wrapper around the query but
this example allows for either.

When published events arrive the `onmessage` function is called with
the message. The message `data` is what was published by the server.

``` html
<html>
  <body>
    <p id="status"> ... </p>
    <p id="price"> ... waiting ... </p>

    <script type="text/javascript">
      var sock;
      var url = "ws://" + document.URL.split('/')[2] + '/graphql'
      if (typeof MozWebSocket != "undefined") {
        sock = new MozWebSocket(url);
      } else {
        sock = new WebSocket(url);
      }
      sock.onopen = function() {
        document.getElementById("status").textContent = "connected";
        sock.send("subscription{listenPrice}")
      }
      sock.onmessage = function(msg) {
        data = JSON.parse(msg.data)
        document.getElementById("price").textContent = "price: " + data["data"];
      }
    </script>
  </body>
</html>
```

### Golang Server

Three files are used, [main.go](main.go), [websoc.go](websoc.go) and
[frame.go](frame.go). The [websoc.go](websoc.go) and
[frame.go](frame.go) files are the implementation of the WebSocket
functionality.

There are a few WebSocket packages available but since the protocol
isn't very complicated it seemed better to keep dependencies down and
expose how the WebSocket protocol works. The code is in the
[websoc.go](websoc.go) and [frame.go](frame.go) files.

The main.go file is the main part of the application that sets up the
server and the resolvers.

#### [main.go](main.go)

[main.go](main.go) implements the server and sets up the
resolvers. Four types are used for the resolvers. As expected the type
names match the GraphQL fields. Note the addition of the `ggql.Root`
in the Mutation and Subscription types. The root will be used to set
up a subscription and to publish events.

``` golang
type Schema struct {
	Query        Query
	Mutation     Mutation
	Subscription Subscription
}

type Query struct {
}

type Mutation struct {
	root *ggql.Root
}

type Subscription struct {
	root *ggql.Root
}
```

The price is just a global variable. A more expansive application
might use a cache or database.

``` golang
var price = 1.1
```

The implementation of the resolvers is a mix of the GGql reflection
and interface approaches. Reflection is used when possible for
simplicity and the interface approach is used when needed. For getting
and setting the price the reflection approach is employed. Right after
setting the price in the Mutation an event is added that will be
published to any subscribers.

``` golang
func (q *Query) Price() float64 {
	return price
}

func (m *Mutation) SetPrice(p float64) float64 {
	price = p
	if _, err := m.root.AddEvent("price", price); err != nil {
		fmt.Println(err.Error())
	}
	return price
}
```

The Subscription resolver uses the interface approach to resolve the
`listenPrice` subscription. A new subscription needs the field
information to build the query result when it is triggered. The
interface approach includes that information when it is invoked so it
must be used instead of the reflection approach.

``` golang
func (s *Subscription) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "listenPrice":
		if ws, _ := field.Context.(*WebSoc); ws != nil {
			ws.sub = ggql.NewSubscription(ws, field, args)
			_ = ws.Send(price) // Sends an initial value.
			return ws.sub, nil
		}
		return nil, fmt.Errorf("listenPrice subscription expected an upgradeable connection")
	default:
		return nil, fmt.Errorf("type Schema does not have field %s", field)
	}
}
```

The HTTP handler for the `/graphql` endpoint needs to handle GET and
POST requests as well as WebSocket upgrade requests. Errors, if they
occur, are returned using the ResponseWriter so instead of returning on
error the flow of the function skips blocks of code if `err` if not
nil. Early in the handler a check is made for whether the connection
should be hijacked or not. That check is whether the ResponseWriter is
also a Hijacker which is always is and whether there is an Upgrade
header set to "WebSocket". If so then a new WebSoc is created.

``` golang
func handleGraphQL(w http.ResponseWriter, req *http.Request, root *ggql.Root) {
	var result map[string]interface{}
	var exe *ggql.Executable
	var err error

	op := req.URL.Query().Get("operationName")
	vars := map[string]interface{}{}
	if jvars := req.URL.Query().Get("variables"); 0 < len(jvars) {
		err = json.Unmarshal([]byte(jvars), &vars)
	}
	var ws *WebSoc
	if err == nil && strings.EqualFold(req.Header.Get("Upgrade"), "WebSocket") {
		if jack, _ := w.(http.Hijacker); jack != nil {
			ws, err = NewWebSoc(root, req, jack)
		}
	}
```

The code in this example is different than the other examples in that
parsing and execution of a query is broken into two steps so that the
WebSoc instance can be passed to the subscription call. Otherwise the
GET and POST handling is similar to the other examples.

``` golang
	if err == nil {
		if ws == nil {
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
```

For a subscription the parsed `ggql.Executable` needs to be prepared
but walking the executable operations and fields and setting the field
Context with the WebSoc instance so that when the `Resolve()` function
is called the WebSoc is available for use.

``` golang
		} else {
			op = ws.Op()
			vars = ws.Vars()
			if exe, err = root.ParseExecutableString(ws.Query()); err == nil {
				prepExe(exe, ws)
			}
		}
	}
```

Stepping out of the handler for a moment and looking at the recursive
prepare functions we see it is composed of two functions. Once walks
the operations in the executable and the other recurses through the
selections. The context on every field is then set to the WebSoc
instance. Not every field needs to have the Context set but it was
easier to set all for the example instead of checking path and field
name.

``` golang
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
```

Back to the handler. With the operation and vars already set the
resolver functions can be called.

``` golang
	if err == nil {
		if result, err = root.ResolveExecutable(exe, op, vars); result == nil {
			result = map[string]interface{}{"data": nil}
		}
	}
```

If the connection has been upgraded a response should not be written
as the connection has been hijacked so just return.

``` golang
	if ws != nil {
		return
	}
```

For a non-hijacked request the results are formed and written as in
other examples.

``` golang
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
```

Putting it all together the resolvers are setup and the schema SDL
used to create a GGql root object. The Mutation and Subscription
resolvers are given the root for use later and then it is on to the
HTTP server setup.

``` golang
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
```

GGql has a feature that allows the schema to be returned without
building a nested introspection query. A handler for that is
registered with just a few lines of code.

``` golang
	http.HandleFunc("/graphql/schema", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		full := strings.EqualFold(q.Get("full"), "true")
		desc := strings.EqualFold(q.Get("desc"), "true")
		sdl := root.SDL(full, desc)
		_, _ = w.Write([]byte(sdl))
	})
```

The GraphQL handler is registered along with a handler to serve the
price.html file and then the server is started.

``` golang

	// The primary endpoint.
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		handleGraphQL(w, r, root)
	})
	// The page with the Javascript that makes a WebSocket call.
	http.HandleFunc("/price.html", func(w http.ResponseWriter, r *http.Request) {
		content, _ := ioutil.ReadFile("price.html")
		_, _ = w.Write(content)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Printf("*-*-* Server failed. %s\n", err)
	}
}
```

#### [websoc.go](websoc.go)

The WebSocket code also includes support needed to interface with
GGql. Near the top of the file is a counter used to generate a server
unique identifier for the connection. It is only incremented and read
atomically.

``` golang
var idCount int64 = 0
```

Examining the WebSoc struct there is an `id` needed later to
unsubscribe and to identify a subscription. Along with the `id` are
the GGql root and subscription handle. These are used to
unsubscribe. When unsubscribing the upgraded connection should be
closed so the `net.Conn` from the HTTP request is included in the
struct along with the buffered `ReadWriter`. The reading and writing
doesn't really need to be buffered for this example but it would help
performance in a non-trivial application.

For query and mutations that make use of GET and POST requests the
parameters needed to resolve a query are either in the URL query or in
the body of a request. For WebSockets, neither the URL query nor the
body are passed from the Javascript call to the server. Instead the
parameters must be in a WebSocket message. The chosen approach to
passing the query parameters was to put them in the WebSoc struct
although they could have just been passed as return values from the
`NewWebSoc()` function.

``` golang
type WebSoc struct {
	id    string
	root  *ggql.Root
	sub   *ggql.Subscription
	con   net.Conn
	rw    *bufio.ReadWriter
	query string
	op    string
	vars  map[string]interface{}
}
```

When upgrading a connection to a WebSocket connection a open handshake
response must be sent by the server. The `NewWebSoc()` function not
only creates a new WebSoc struct but it also sends the accept response
of the handshake. In an attempt to keep as much of the WebSocket code
out of the main part of the code the `NewWebSoc()` function also reads
in GraphQL query parameters.

``` golang
func NewWebSoc(root *ggql.Root, req *http.Request, jack http.Hijacker) (ws *WebSoc, err error) {
	id := atomic.AddInt64(&idCount, 1)
	ws = &WebSoc{
		root: root,
		id:   strconv.FormatInt(id, 10),
	}
	// Hijack the connection.
	if ws.con, ws.rw, err = jack.Hijack(); err != nil {
		return
	}
	// Build the acceptance message as the response in the open handshake.
	h := sha1.New()
	_, _ = h.Write([]byte(req.Header.Get("Sec-WebSocket-Key")))
	_, _ = h.Write([]byte(wsMagic))
	var accept []byte
	accept = append(accept, "HTTP/1.1 101 Switching Protocols\r\n"...)
	accept = append(accept, "Upgrade: websocket\r\n"...)
	accept = append(accept, "Connection: Upgrade\r\n"...)
	accept = append(accept, "Sec-WebSocket-Accept: "...)
	accept = append(accept, base64.StdEncoding.EncodeToString(h.Sum(nil))...)
	accept = append(accept, "\r\n\r\n"...)
	if _, err = ws.rw.Write(accept); err != nil {
		return
	}
	if err = ws.rw.Flush(); err != nil {
		return
	}
```

While there is no strict definition of the content of the WebSocket
exchange, JSON is often the format of the contents. With a bias toward
avoiding the double encoding needed when using JSON (query encode as
GraphQL and then as JSON) this example also support straight GraphQL
as the content. If the message payload starts with a { it is assumed
to be JSON otherwise a GraphQL subscription is expected.

``` golang
	var msg []byte
	if msg, _, err = ws.read(); err != nil {
		ws.Unsubscribe()
		return
	}
	msg = bytes.TrimSpace(msg)
	if 0 < len(msg) && msg[0] == '{' {
		var j map[string]interface{}
		if err = json.Unmarshal(msg, &j); err != nil {
			return
		}
		ws.op, _ = j["operationName"].(string)
		ws.vars, _ = j["variables"].(map[string]interface{})
		ws.query, _ = j["query"].(string)
	} else {
		ws.query = string(msg)
	}
```

Just before returing a listen loop is created that listens for close
and ping messages.

``` golang
	go ws.listen()

	return
}
```

When a new event is generated that data for that event is passed to
the subscriber, the WebSoc instance, using the `Send()` function. The
event value is encoded a JSON and then written to the connection
usings the `ReadWriter`.

``` golang
func (ws *WebSoc) Send(value interface{}) error {
	j, err := json.Marshal(map[string]interface{}{
		"data": value,
	})
	if err != nil {
		return err
	}
	f := newFrame(j)
	if _, err = ws.rw.Write(f); err == nil {
		err = ws.rw.Flush()
	}
	return err
}
```

An unsubscribe is triggered by either the client or the server closing
a connection. In both cases a close handshake should be initiated and
responded to. In practice this doesn't always occur but the
`Unsubscribe()` function makes an attempt to follow RFC 6455 but
ignores failures.

``` golang
func (ws *WebSoc) Unsubscribe() {
	if _, err := ws.rw.Write([]byte{0x88, 0x02, 0x03, 0xEB}); err == nil {
		_ = ws.rw.Flush()
	}
	if ws.con != nil {
		_ = ws.con.Close()
		ws.con = nil
	}
}
```

The `read()` function must handle reads that don't come all at
once. This is particularly important for longer payloads. A working
buffer is created and then `Read()` is called until the frame is fully
read. This is determined by the checking the expected size of the
frame with the number of bytes read.

``` golang
func (ws *WebSoc) read() ([]byte, int, error) {
	var f frame
	var err error
	buf := make([]byte, 4096)
	var n int
	for {
		if n, err = ws.rw.Read(buf); err != nil {
			return nil, 0, err
		}
		f = append(f, buf[:n]...)
		if f.ready() {
			break
		}
	}
	return f.payload(), f.op(), err
}
```

A WebSocket connection is bidirectional. The client should send a
close frame when disconnecting and ping messages could also be sent
which must be responded to. The `listen()` loop function continues to
attempt a reads on the connection and responds to a ping with a pong
and a close with another close. A close response is follwed by an
unsubscribe and the closing of the connection.

``` golang
func (ws *WebSoc) listen() {
	for ws.con != nil {
		_, op, err := ws.read()
		if err == nil {
			switch op {
			case opPing:
				// Send pong if a ping is received. Some browser might send a
				// ping although I didn't come across one in testing.
				if _, err = ws.rw.Write([]byte{0x80 | opPong, 0x00}); err == nil {
					err = ws.rw.Flush()
				}
			case opClose:
				ws.root.Unsubscribe(ws.id)
				if _, err = ws.rw.Write([]byte{0x88, 0x02, 0x03, 0xEB}); err == nil {
					err = ws.rw.Flush()
				}
				if ws.con != nil {
					_ = ws.con.Close()
					ws.con = nil
				}
			}
		}
		if err != nil {
			break
		}
	}
	ws.root.Unsubscribe(ws.id)
	ws.con = nil
}
```

#### [frame.go](frame.go)

WebSockets makes use of frames to package payloads for exchange
between clients and servers. The frame consists of a headers and a
payload. The header includes an opcode such as for text, close, ping,
pong, and others. Next is an encoded length that varies in size
depending on length to be encoded. Following the length is the
optional mask which is followed by the payload.

The frame type is a `[]byte` since only methods are needed to extract
information from a frame. Methods include getting the various parts of
the frame such as the payload, opcode, and expected length. A function
is also included to create a frame for a payload.

The format of a frame is defined by [RFC
6455](https://tools.ietf.org/html/rfc6455) which includes this
diagram.

```
   0                   1                   2                   3
   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
  +-+-+-+-+-------+-+-------------+-------------------------------+
  |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
  |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
  |N|V|V|V|       |S|             |   (if payload len==126/127)   |
  | |1|2|3|       |K|             |                               |
  +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
  |     Extended payload length continued, if payload len == 127  |
  + - - - - - - - - - - - - - - - +-------------------------------+
  |                               |Masking-key, if MASK set to 1  |
  +-------------------------------+-------------------------------+
  | Masking-key (continued)       |          Payload Data         |
  +-------------------------------- - - - - - - - - - - - - - - - +
  :                     Payload Data continued ...                :
  + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
  |                     Payload Data continued ...                |
  +---------------------------------------------------------------+
```
