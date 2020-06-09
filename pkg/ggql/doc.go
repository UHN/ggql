// Copyright 2019-2020 University Health Network
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ggql provides a GraphQL implementation.
//
// GraphQL schema defines an API for a backend implementation. Some
// packages require Go code to define the schema. This package keeps the
// GraphQL schema definitions in their original SDL form.
//
// Parsing of a GraphQL schema in SDL builds an internal representation
// of the schema. Requests are also parsed to build a document that will
// be evaluated. As this document is evaluated calls are made to resolve
// the nodes of the query, mutation, or subscription. The resolvers are
// the linked go code.
//
// There are three approaches used to implementing the resolvers.
//
//   - Reflection Resolvers
//   - Interface Resolvers
//   - Root Resolver
//
// All three can be used together. The first priority is given to
// the Interface Resolvers. Failing to find an Interface Resolver, a check
// is made for the Root Resolver. If no other resolvers match then the
// Reflection Resolver function is attempted.
//
// Reflection Resolvers
//
// The type binding approach uses reflection as needed but also caches
// reflection objects when feasible.
//
// Unlike some other packages this implementation does not require a
// separate translation function for each resolver. With some other
// packages structs must be populated to define the schema using a
// proprietary format that is often poorly documented. Documentation
// aside, the functions are often global and not associated with a
// type. The APIs are not at all simple. This package takes the approach
// of encapsulating functionality in individual types along with some
// restrictions on function signatures with restrictions mostly on return
// values. The binding between the GraphQL type can be registered or if
// the names of fields match fields or methods of the Go struct the
// registration is automatic. Simple matches such as case changes are
// used to look for matches between the SDL and Go types.
//
// An overview of using the reflection resolver can be found in this
// [example](examples/reflection/README.md).
//
// Interface Resolvers
//
// In some cases it is preferable to have a more dynamic approach to
// binding GraphQL types to resolvers. This approach relies on two
// interfaces that data elements in the graph must implement. An
// interface resolver only approach does not support conditional fragment,
// unions, or interfaces. It requires a bit more code but does support
// Go interface use and non-slice collections.
//
// An overview of using the interface resolver can be found in this
// [example](examples/interface/README.md).
//
// Root Resolver
//
// If the target does not implement the Resolver interface then the root
// resolver can be used. Starting with the root object the resolve
// function should be able to dig down into the data according to the
// provided key. As an example, if the data is a map[string]interface{}
// then the resolver might simply look like this:
//
//   func Resolve(
//       target interface{},
//       field *Field,
//       args map[string]interface{}) (interface{}, error) {
//
//       if m, _ := target.(map[string]interface{}); m != nil {
//           return m[field.Name], nil
//       }
//       return nil, nil
//   }
//
// Of course the actual resolver might be much more complex if the data
// must be retrieved from a database.
//
// An overview of using the root resolver can be found in this
// [example](examples/root/README.md).
//
// Subscriptions
//
// The GraphQL spec is intentionally vague on what the subscription
// operation does and how it works. Since the GraphQL comes over a
// particular transport (e.g. HTTP) it makes sense to restrict the
// publish and subscribe behavior to the same transport. If an HTTP
// transport is used, then publish and subscribe would be done with Websockets or
// SSE. If NATS is the transport then NATS would be used to publish to a
// NATS listener providing the resolved subscription query results.
//
// The package uses a pluggable API for hooking in a subscription
// implementation.
//
package ggql
