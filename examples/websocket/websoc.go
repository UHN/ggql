package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/uhn/ggql/pkg/ggql"
)

const wsMagic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

var idCount int64 = 0

// WebSoc is a WebSocket connection.
type WebSoc struct {
	id   string
	root *ggql.Root
	sub  *ggql.Subscription
	con  net.Conn
	rw   *bufio.ReadWriter

	query string
	op    string
	vars  map[string]interface{}
}

// NewWebSoc creates a new WebSocket connection, performs the connection
// handshake, and reads the GraphQL query.
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
	// The query to evaluate should be the first message sent by the client.
	var msg []byte
	if msg, _, err = ws.read(); err != nil {
		ws.Unsubscribe()
		return
	}
	// Both plain graphql and JSON are supported. If the msg starts with a {
	// it is assumed to be JSON otherwise a GraphQL subscription is expected.
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
	// Listen for a close or a ping.
	go ws.listen()

	return
}

// Send an event on channel associated with the subscription.
func (ws *WebSoc) Send(value interface{}) error {
	// The wrapper is not really needed but it keeps things consistent with
	// regular GraphQL responses.
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

// Match an eventID to the subscriber's ID. Behavior is up to the
// subscriber. An exact match could be used or a system that uses
// wildcards can be used. A return of true indicates a match.
func (ws *WebSoc) Match(eventID string) bool {
	return eventID == ws.id || eventID == "price"
}

// Unsubscribe should clean up any resources associated with a
// subscription.
func (ws *WebSoc) Unsubscribe() {
	if _, err := ws.rw.Write([]byte{0x88, 0x02, 0x03, 0xEB}); err == nil {
		_ = ws.rw.Flush()
	}
	if ws.con != nil {
		_ = ws.con.Close()
		ws.con = nil
	}
}

// Query returns the GraphQL query string.
func (ws *WebSoc) Query() string {
	return ws.query
}

// Op returns the GraphQL operation.
func (ws *WebSoc) Op() string {
	return ws.op
}

// Vars returns the GraphQL operation.
func (ws *WebSoc) Vars() map[string]interface{} {
	return ws.vars
}

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
