# GGql Examples

| [Home](../README.md) |
| -------------------- |

GGql includes three approaches to linking Go code to a GraphQL schema
document. In this examples directory each of those approaches are
demonstrated. If you are new to GraphQL it would be best to start with
the **reflection** example as it includes a lengthy tutorial style
[README.md](reflection/README.md) file.

All examples use a common schema for simple musical data.

```graphql
type Query {
  artist(name: String!): Artist
  artists: [Artist]
}

type Mutation {
  like(artist: String!, song: String!): Song
}

type Artist {
  name: String!
  songs: [Song]
  origin: [String]
}

type Song {
  name: String!
  artist: Artist
  duration: Int
  release: Date
  likes: Int
}

scalar Date
```

While this schema is not very complex, it does provide the foundation
for extending to more examples with GraphQL Unions, Interfaces, and
Fragments.

The features supported by each approach are summarized in the following
comparison matrix:

| Feature                          | Reflection         | Interface          | Root               |
|:---------------------------------|--------------------|--------------------|--------------------|
| Auto resolve (map GraphQL to Go) | :heavy_check_mark: |                    |                    |
| GraphQL types as Go interfaces   |                    | :heavy_check_mark: |                    |
| Non-slice lists or collections   |                    | :heavy_check_mark: |                    |
| Run time defined type mapping    |                    |                    | :heavy_check_mark: |
| Field arguments                  | :heavy_check_mark: | :heavy_check_mark: |                    |
| Compile time resolver checking   |                    | :heavy_check_mark: |                    |
| Supports directive use           | :heavy_check_mark: | :heavy_check_mark: |                    |
| Fragment condition               | :heavy_check_mark: |                    |                    |
| GraphQL Union and Interfaces     | :heavy_check_mark: |                    |                    |

## Examples

* [Reflection Resolver Example](reflection/README.md)
* [Interface Resolver Example](interface/README.md)
* [Root Resolver Example](root/README.md)
