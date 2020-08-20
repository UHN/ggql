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

package ggql_test

import (
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestRootAddType(t *testing.T) {
	root := ggql.NewRoot(nil)
	e := ggql.Enum{Base: ggql.Base{N: "Ee"}}
	e.AddValue(&ggql.EnumValue{Value: "Aa"})
	err := root.AddTypes(&e)
	checkNil(t, err, "no error should be returned whan adding a type. %s", err)
	err = root.AddTypes(&e)
	checkNotNil(t, err, "an error should be returned whan adding a duplicate type. %v was returned", err)
}

func TestRootGetType(t *testing.T) {
	root := ggql.NewRoot(nil)
	ty := root.GetType("Int")
	checkNotNil(t, ty, "GetType(Int)")
	ty = root.GetType("Nope")
	checkNil(t, ty, "GetType(Nope)")
}

func TestRootInit(t *testing.T) {
	root := ggql.NewRoot(nil)

	checkEqual(t, 16, len(root.Types()), "Root types should have all build in types.")

	q := ggql.Object{Base: ggql.Base{N: "Query"}}
	q.AddField(&ggql.FieldDef{Base: ggql.Base{N: "dummy"}, Type: &ggql.Ref{Base: ggql.Base{N: "Int"}}})

	m := ggql.Object{Base: ggql.Base{N: "Mutation"}}
	m.AddField(&ggql.FieldDef{Base: ggql.Base{N: "dummy"}, Type: &ggql.Ref{Base: ggql.Base{N: "Int"}}})

	s := ggql.Object{Base: ggql.Base{N: "Subscription"}}
	s.AddField(&ggql.FieldDef{Base: ggql.Base{N: "dummy"}, Type: &ggql.Ref{Base: ggql.Base{N: "Int"}}})

	root.AddTypes(&q, &m, &s)

	actual := root.SDL(true, true)
	expect := `
type __Schema {
  types: [__Type!]!
  queryType: __Type!
  mutationType: __Type
  subscriptionType: __Type
  directives: [__Directive!]!
}

type Query {
  dummy: Int
}

type Mutation {
  dummy: Int
}

type Subscription {
  dummy: Int
}

type __Directive {
  name: String!
  description: String
  locations: [__DirectiveLocation!]!
  args: [__InputValue!]!
}

type __EnumValue {
  name: String!
  description: String
  isDeprecated: Boolean!
  deprecationReason: String
}

type __Field {
  name: String!
  description: String
  args: [__InputValue!]!
  type: __Type!
  isDeprecated: Boolean!
  deprecationReason: String
}

type __InputValue {
  name: String!
  description: String
  type: __Type!
  defaultValue: String
}

type __Type {
  kind: __TypeKind!
  name: String
  description: String
  fields(includeDeprecated: Boolean = false): [__Field!]
  enumValues(includeDeprecated: Boolean = false): [__EnumValue!]
  inputFields: [__InputValue!]
  interfaces: [__Type!]!
  possibleTypes: [__Type!]!
  ofType: __Type
}

enum __DirectiveLocation {
  QUERY
  MUTATION
  SUBSCRIPTION
  FIELD
  FRAGMENT_DEFINITION
  FRAGMENT_SPREAD
  INLINE_FRAGMENT
  SCHEMA
  SCALAR
  OBJECT
  FIELD_DEFINITION
  ARGUMENT_DEFINITION
  INTERFACE
  UNION
  ENUM
  ENUM_VALUE
  INPUT_OBJECT
  INPUT_FIELD_DEFINITION
}

enum __TypeKind {
  SCALAR
  OBJECT
  INTERFACE
  UNION
  ENUM
  INPUT_OBJECT
  LIST
  NON_NULL
}

scalar Boolean

scalar Float

scalar Float64

scalar ID

scalar Int

scalar Int64

scalar String

"""
Time can be coerced to and from a string that is formatted according to the
RFC3339 specification with nanoseconds.
"""
scalar Time

directive @deprecated(reason: String = "\"No longer supported\"") on FIELD_DEFINITION | ENUM_VALUE

directive @go(type: String!) on SCHEMA | QUERY | MUTATION | SUBSCRIPTION | OBJECT | FIELD_DEFINITION

directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

directive @skip(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT
`
	checkEqual(t, expect, actual, "root SDL() mismatch")
}

func TestRootInitExclude(t *testing.T) {
	root := ggql.NewRoot(nil, "Time", "Int64")

	actual := root.SDL(true, true)
	expect := `
type __Schema {
  types: [__Type!]!
  queryType: __Type!
  mutationType: __Type
  subscriptionType: __Type
  directives: [__Directive!]!
}

type __Directive {
  name: String!
  description: String
  locations: [__DirectiveLocation!]!
  args: [__InputValue!]!
}

type __EnumValue {
  name: String!
  description: String
  isDeprecated: Boolean!
  deprecationReason: String
}

type __Field {
  name: String!
  description: String
  args: [__InputValue!]!
  type: __Type!
  isDeprecated: Boolean!
  deprecationReason: String
}

type __InputValue {
  name: String!
  description: String
  type: __Type!
  defaultValue: String
}

type __Type {
  kind: __TypeKind!
  name: String
  description: String
  fields(includeDeprecated: Boolean = false): [__Field!]
  enumValues(includeDeprecated: Boolean = false): [__EnumValue!]
  inputFields: [__InputValue!]
  interfaces: [__Type!]!
  possibleTypes: [__Type!]!
  ofType: __Type
}

enum __DirectiveLocation {
  QUERY
  MUTATION
  SUBSCRIPTION
  FIELD
  FRAGMENT_DEFINITION
  FRAGMENT_SPREAD
  INLINE_FRAGMENT
  SCHEMA
  SCALAR
  OBJECT
  FIELD_DEFINITION
  ARGUMENT_DEFINITION
  INTERFACE
  UNION
  ENUM
  ENUM_VALUE
  INPUT_OBJECT
  INPUT_FIELD_DEFINITION
}

enum __TypeKind {
  SCALAR
  OBJECT
  INTERFACE
  UNION
  ENUM
  INPUT_OBJECT
  LIST
  NON_NULL
}

scalar Boolean

scalar Float

scalar Float64

scalar ID

scalar Int

scalar String

directive @deprecated(reason: String = "\"No longer supported\"") on FIELD_DEFINITION | ENUM_VALUE

directive @go(type: String!) on SCHEMA | QUERY | MUTATION | SUBSCRIPTION | OBJECT | FIELD_DEFINITION

directive @include(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT

directive @skip(if: Boolean!) on FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT
`
	checkEqual(t, expect, actual, "root SDL() mismatch")
}

func TestRootReplaceRefsOk(t *testing.T) {
	root := ggql.NewRoot(nil)
	rep := ggql.Object{Base: ggql.Base{N: "Replace"}}
	fd := ggql.FieldDef{Base: ggql.Base{N: "list"},
		Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Int"}}},
	}
	fd.AddArg(&ggql.Arg{Base: ggql.Base{N: "a"}, Type: &ggql.Ref{Base: ggql.Base{N: "String"}}})
	fd.AddArg(&ggql.Arg{Base: ggql.Base{N: "b"}, Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "String"}}}})
	rep.AddField(&fd)

	rep.AddField(&ggql.FieldDef{
		Base: ggql.Base{
			N:    "ref",
			Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "dirt"}}}},
		},
		Type: &ggql.Ref{Base: ggql.Base{N: "String"}},
	})
	rep.AddField(&ggql.FieldDef{Base: ggql.Base{N: "nest"},
		Type: &ggql.List{Base: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Int"}}}},
	})

	dir := ggql.Directive{Base: ggql.Base{N: "dirt"}, On: []ggql.Location{ggql.LocUnion}}
	dir.AddArg(&ggql.Arg{Base: ggql.Base{N: "reason"}, Type: &ggql.NonNull{Base: root.GetType("String")}})

	err := root.AddTypes(&dir,
		&ggql.Union{
			Base: ggql.Base{
				N: "Unit",
				Dirs: []*ggql.DirectiveUse{
					{
						Directive: &ggql.Ref{Base: ggql.Base{N: "dirt"}},
						Args: map[string]*ggql.ArgValue{
							"reason": {
								Arg:   "reason",
								Value: "no reason",
							},
						},
					},
				},
			},
			Members: []ggql.Type{&ggql.Ref{Base: ggql.Base{N: "Replace"}}},
		},
		&rep,
	)
	checkNil(t, err, "schema parse failed. %s", err)

	actual := root.SDL(false, false)
	expect := `
type Replace {
  list(a: String, b: [String]): [Int]
  ref: String @dirt
  nest: [[Int]]
}

union Unit @dirt(reason: "no reason") = Replace

scalar Time

directive @dirt(reason: String!) on UNION
`
	checkEqual(t, expect, actual, "root SDL() mismatch")
}

func TestRootReplaceRefsBadArgs(t *testing.T) {
	root := ggql.NewRoot(nil)
	obj := &ggql.Object{Base: ggql.Base{N: "BadArg"}}
	fd := &ggql.FieldDef{Base: ggql.Base{N: "bad"},
		Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Int"}}},
	}
	fd.AddArg(&ggql.Arg{Base: ggql.Base{N: "a"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})
	obj.AddField(fd)
	err := root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")

	obj = &ggql.Object{Base: ggql.Base{N: "BadArg"}}

	fd = &ggql.FieldDef{Base: ggql.Base{N: "bad"},
		Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Int"}}},
	}
	fd.AddArg(&ggql.Arg{Base: ggql.Base{N: "a"}, Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}})
	obj.AddField(fd)
	err = root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")

	obj = &ggql.Object{Base: ggql.Base{N: "BadArg"}}
	fd = &ggql.FieldDef{Base: ggql.Base{N: "bad"},
		Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Int"}}},
	}
	fd.AddArg(&ggql.Arg{Base: ggql.Base{N: "a"}, Type: &ggql.NonNull{Base: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}})
	obj.AddField(fd)
	err = root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")
}

func TestRootReplaceRefsBadField(t *testing.T) {
	root := ggql.NewRoot(nil)

	obj := &ggql.Object{Base: ggql.Base{N: "BadField"}}
	obj.AddField(&ggql.FieldDef{Base: ggql.Base{N: "bad"}, Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}})
	err := root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")

	obj = &ggql.Object{Base: ggql.Base{N: "BadField"}}
	obj.AddField(&ggql.FieldDef{Base: ggql.Base{N: "bad"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})
	err = root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")
}

func TestRootReplaceRefsBadInterface(t *testing.T) {
	root := ggql.NewRoot(nil)

	obj := &ggql.Object{Base: ggql.Base{N: "BadField"},
		Interfaces: []ggql.Type{&ggql.Ref{Base: ggql.Base{N: "Bad"}}},
	}
	obj.AddField(&ggql.FieldDef{Base: ggql.Base{N: "a"}, Type: root.GetType("Int")})
	err := root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")
}

func TestRootReplaceRefsDoubleNonNull(t *testing.T) {
	root := ggql.NewRoot(nil)

	obj := &ggql.Object{Base: ggql.Base{N: "BadField"}}
	obj.AddField(&ggql.FieldDef{Base: ggql.Base{N: "bad"},
		Type: &ggql.NonNull{Base: &ggql.NonNull{Base: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}},
	})
	err := root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")
}

func TestRootReplaceRefsBadDir(t *testing.T) {
	root := ggql.NewRoot(nil)
	err := root.AddTypes(
		&ggql.Object{
			Base: ggql.Base{
				N:    "BadDir",
				Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "bad"}}}},
			},
		},
	)
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(
		&ggql.Union{
			Base: ggql.Base{
				N:    "BadDir",
				Dirs: []*ggql.DirectiveUse{{Directive: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "bad"}}}}},
			},
		},
	)
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(
		&ggql.Enum{
			Base: ggql.Base{
				N:    "NotDir",
				Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "Boolean"}}}},
			},
		},
	)
	checkNotNil(t, err, "check error was returned")

	dir := ggql.Directive{
		Base: ggql.Base{N: "dirt"},
		On:   []ggql.Location{ggql.LocEnumValue},
	}
	dir.AddArg(&ggql.Arg{Base: ggql.Base{N: "bad"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})

	enum := ggql.Enum{
		Base: ggql.Base{
			N: "NoDirArg",
		},
	}
	enum.AddValue(&ggql.EnumValue{
		Value:      "BAD",
		Directives: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}},
	})
	err = root.AddTypes(&dir, &enum)
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(
		&ggql.Schema{
			Object: ggql.Object{
				Base: ggql.Base{
					N:    "BadDir",
					Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "bad"}}}},
				},
			},
		},
	)
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(
		&ggql.Interface{
			Base: ggql.Base{
				N:    "BadDir",
				Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "bad"}}}},
			},
			Root: root,
		},
	)
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(
		&ggql.Input{
			Base: ggql.Base{
				N:    "BadDir",
				Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "bad"}}}},
			},
		},
	)
	checkNotNil(t, err, "check error was returned")

	input := &ggql.Input{Base: ggql.Base{N: "BadListField"}}
	input.AddField(&ggql.InputField{Base: ggql.Base{N: "a"}, Type: &ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}})
	err = root.AddTypes(input)
	checkNotNil(t, err, "check error was returned")

	input = &ggql.Input{Base: ggql.Base{N: "BadNonNullField"}}
	input.AddField(&ggql.InputField{Base: ggql.Base{N: "a"}, Type: &ggql.NonNull{Base: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}})
	err = root.AddTypes(input)
	checkNotNil(t, err, "check error was returned")
}

func TestRootReplaceRefsBadTop(t *testing.T) {
	root := ggql.NewRoot(nil)

	err := root.AddTypes(&ggql.Ref{Base: ggql.Base{N: "Ref"}})
	checkNotNil(t, err, "check error was returned when adding a ref")

	err = root.AddTypes(&ggql.List{Base: &ggql.Ref{Base: ggql.Base{N: "Ref"}}})
	checkNotNil(t, err, "check error was returned when adding a list type")

	obj := &ggql.Schema{Object: ggql.Object{Base: ggql.Base{N: "schema"}}}
	obj.AddField(&ggql.FieldDef{Base: ggql.Base{N: "bad"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})

	err = root.AddTypes(obj)
	checkNotNil(t, err, "check error was returned")

	inty := &ggql.Interface{Base: ggql.Base{N: "Inty"}, Root: root}
	inty.AddField(&ggql.FieldDef{Base: ggql.Base{N: "bad"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})
	err = root.AddTypes(inty)
	checkNotNil(t, err, "check error was returned")

	input := &ggql.Input{Base: ggql.Base{N: "Putter"}}
	input.AddField(&ggql.InputField{Base: ggql.Base{N: "bad"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})
	err = root.AddTypes(input)
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(
		&ggql.Union{Base: ggql.Base{N: "Unit"},
			Members: []ggql.Type{&ggql.Ref{Base: ggql.Base{N: "Bad"}}},
		})
	checkNotNil(t, err, "check error was returned")

	err = root.AddTypes(&ggql.Ref{Base: ggql.Base{N: "Ref"}})
	checkNotNil(t, err, "check error was returned")
}
