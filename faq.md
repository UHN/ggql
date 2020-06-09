# GGql FAQ

1) **What order are query elements resolved in?**

  Executables are resolved in depth first order.

2) **How does the reflection resolver determine which Go function to use?**

  Once a Go type has been identified during resolving a field or
  method is searched for using a case insensitive check of fields
  first and then methods. This allows the GraphQL convention of
  lowercase field names to match captitalized Go public fields and
  methods. To override this matching fields can be registered directly
  with the `Root.RegisterType()` and `Root.RegisterField()` functions.

3) **How can a context or user data be made available in the Resolve functions?**

  Context is attached to the `Field` type. It can be set directly
  during resolving on individual fields or recursively for all fields
  in a query with the `Executable.SetContextRecursive()` function. If
  the `Nester` interface is used each level of fields can be altered
  witha different context. An example of using the `Nester` interface
  would be to build a path that is stored in each field.

4) **When would the reflection resolver approach be used?**

  The reflection resolver approach is a good way to get started. It
  requires very little code and works well when there is a fairly
  direct mapping between GraphQL type and Go types.

5) **When would the interface resolver approach be used?**

  In more complex system where lists are not slices or when GraphQL
  types don't map cleanly to Go types then the interface resolver
  approach is more desireable than the reflection resolver
  approach. If non-global context is needed for resolving then the
  interface resolver approach is the one to use.

6) **When would the root resolver approach be used?**

  The root resolver approach is best when a dynamic mapping from a
  GraphQL schema to a data graph is needed. A good example would be
  loading a schema file and having that map onto JSON loaded from a
  database.

7) **How are Unions used and what is the @go directive?**

  Unions allow a field to return a list or object where the base type is one of a
  set of types that need not have any fields in common.  For example in the song
  schema, a union could be made of Artist and Song. A Query field of _all_ could
  return all Artists and Songs as a single list. Fragments are then used to
  specify what fields of each type to return. See [this
  page](https://graphql.org/learn/schema/#union-types) for more info on unions.

  GGQL requires some extra info to determine the GraphQL type from the 
  implementation type when resolving Unions. The `@go(type: String!)` directive
  associates an implementation type in go with the GraphQL type which is the 
  missing piece. The type argument should be either the full path and name of the 
  go type, the short package name and type name, or just the type name.

