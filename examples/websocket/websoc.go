package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/uhn/ggql/pkg/ggql"
)

// WebSoc is a WebSocket connection.
type WebSoc struct {

	// TBD maybe create and pass down, move to separate file, sock.go or ws.go and call the type WS or Woc

	sub *ggql.Subscription
	con net.Conn
	rw  *bufio.ReadWriter

	// TBD socket
}

func NewWebSoc(jack http.Hijacker) (ws *WebSoc, err error) {
	ws = &WebSoc{}
	if ws.con, ws.rw, err = jack.Hijack(); err != nil {
		return
	}

	// TBD handshake

	return
}

// Send an event on channel associated with the subscription.
func (s *WebSoc) Send(value interface{}) error {
	fmt.Printf("*** Send\n")
	return nil
}

// Match an eventID to the subscriber's ID. Behavior is up to the
// subscriber. An exact match could be used or a system that uses
// wildcards can be used. A return of true indicates a match.
func (s *WebSoc) Match(eventID string) bool {
	// Always return true for this example.
	fmt.Printf("*** match %s\n", eventID)
	return true
}

// Unsubscribe should clean up any resources associated with a
// subscription.
func (s *WebSoc) Unsubscribe() {
	// TBD close connection
}
