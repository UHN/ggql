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

func TestExecutable(t *testing.T) {
	dirt := ggql.Directive{
		Base: ggql.Base{N: "dirt"},
		On:   []ggql.Location{ggql.LocInlineFragment, ggql.LocField},
	}
	friend := ggql.Fragment{
		Name: "friend",
		Inline: ggql.Inline{
			Condition: &ggql.Ref{Base: ggql.Base{N: "User"}},
			SelBase: ggql.SelBase{
				Dirs: []*ggql.DirectiveUse{{Directive: &dirt}},
				Sels: []ggql.Selection{
					&ggql.Field{Name: "name"},
					&ggql.Field{
						Name: "picture",
						Args: []*ggql.ArgValue{{Arg: "size", Value: ggql.Var("size")}},
					},
				},
			},
		},
	}
	ex := ggql.Executable{
		Ops: map[string]*ggql.Op{
			"fraggle": {
				Name: "fraggle",
				Type: ggql.OpQuery,
				Variables: []*ggql.VarDef{
					{Name: "size", Type: &ggql.Ref{Base: ggql.Base{N: "Int"}}, Default: 100},
				},
				SelBase: ggql.SelBase{
					Sels: []ggql.Selection{
						&ggql.Field{
							Name: "user",
							Args: []*ggql.ArgValue{{Arg: "name", Value: "Gobo"}},
							SelBase: ggql.SelBase{
								Sels: []ggql.Selection{
									&ggql.Field{Name: "id"},
									&ggql.FragRef{Fragment: &friend},
									&ggql.Inline{
										Condition: &ggql.Ref{Base: ggql.Base{N: "User"}},
										SelBase: ggql.SelBase{
											Dirs: []*ggql.DirectiveUse{{Directive: &dirt}},
											Sels: []ggql.Selection{
												&ggql.Field{Name: "height"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Fragments: map[string]*ggql.Fragment{"friend": &friend},
	}
	checkEqual(t, `query fraggle($size: Int = 100) {
  user(name: "Gobo") {
    id
    ...friend
    ... on User @dirt {
      height
    }
  }
}

fragment friend on User @dirt {
  name
  picture(size: $size)
}
`, ex.String(), "incorrect output for blank op")

}
