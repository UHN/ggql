# WebSockets with GraphQL in Go the Easy Way

| [Home](../../README.md) | [Examples](../README.md) |
| ----------------------- | ------------------------ |

TBD
 -

This example assumes familiarity with the GGql package. The base
examples for using GGql are:

 - [Reflection](../reflection/README.md)
 - [Interface](../interface/README.md)
 - [Root](../root/README.md)


 - this example demonstrates subscriptions over websockets
   - describe how it works, grab connection, etc


## Define the API

A simple API is provided that allows getting, setting, and subscribing
to changes in a price.

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

## The Code


TBD

There are a few WebSocket packages available but since the protocol
isn't very complicated and since I have already written a server side
package in [Agoo](https://github.com/ohler55/agoo) it seemed better to
keep dependencies down and expose how the WebSocket protocol
works. The code is in the [websoc.go](websoc.go) and
[frame.go](frame.go) files.
