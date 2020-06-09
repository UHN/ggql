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

func TestSchema(t *testing.T) {
	root := ggql.NewRoot(nil)
	query := &ggql.Object{Base: ggql.Base{N: "Query"}}
	query.AddField(&ggql.FieldDef{Base: ggql.Base{N: "dummy"}, Type: &ggql.Ref{Base: ggql.Base{N: "Int"}}})

	schema := &ggql.Schema{Object: ggql.Object{}}
	schema.AddField(&ggql.FieldDef{Base: ggql.Base{N: "query"}, Type: query})

	err := root.AddTypes(schema, query)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	expectStr := `schema {
  query: Query
}
`
	checkEqual(t, expectStr, schema.String(), "Schema String() mismatch")
	checkEqual(t, 1, schema.Rank(), "Schema Rank() mismatch")

	x := &ggql.Schema{Object: ggql.Object{Base: ggql.Base{N: "Schema"}}}
	x.AddField(&ggql.FieldDef{Base: ggql.Base{N: "mutation"}, Type: &ggql.Object{Base: ggql.Base{N: "Mutation"}}})
	err = schema.Extend(x)
	checkNil(t, err, "schema.Extend failed. %s", err)

	x = &ggql.Schema{Object: ggql.Object{Base: ggql.Base{N: "Schema"}}}
	x.AddField(&ggql.FieldDef{Base: ggql.Base{N: "mutation"}, Type: &ggql.Object{Base: ggql.Base{N: "Mutation"}}})
	err = schema.Extend(x)
	checkNotNil(t, err, "duplicate field for schema.Extend should have failed.")
}
