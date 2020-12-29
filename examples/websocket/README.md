# WebSockets with GraphQL in Go the Easy Way

| [Home](../../README.md) | [Examples](../README.md) |
| ----------------------- | ------------------------ |

Pushing data to a browser makes for lively up-to-date page. GraphQL
provides many advantages as well but we will not get into that
here. This example brings together WebSockets and GraphQL making use
of GraphQL subscriptions. This example is written in Go using just the
[GGql](https://github.com/UHN/ggql) package.

This example assumes familiarity with the GGql package. Simpler, first
time examples for using GGql are:

 - [Reflection](../reflection/README.md)
 - [Interface](../interface/README.md)
 - [Root](../root/README.md)

## Define the API

As always, a GraphQL application needs a GraphQL schema. For
simplicity a very simple API is provided that allows getting, setting,
and subscribing to changes in a price. The price is just a number and
nothing more.

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
scenario is not complicated.

 1. Start the application.
 2. Open the browser for URL: http://localhost:3000/price.html
 3. Curl in a price change with a mutation.

The price change is then pushed to the browser and the new price
displayed. Multiple browsers and multiple mutation calls are all
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

Three files are used, main.go, websoc.go, and frame.go. The websoc.go,
and frame.go files are the implementation of the WebSocket
functionality.

There are a few WebSocket packages available but since the protocol
isn't very complicated and I have already written a server side
package in [Agoo](https://github.com/ohler55/agoo) it seemed better to
keep dependencies down and expose how the WebSocket protocol
works. The code is in the [websoc.go](websoc.go) and
[frame.go](frame.go) files.

The main.go file is the main part of the application that sets up the
server and the resolvers.

#### [main.go](main.go)


#### [websoc.go](websoc.go)


#### [frame.go](frame.go)
