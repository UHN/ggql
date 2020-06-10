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

func TestInline(t *testing.T) {
	in := ggql.Inline{
		Condition: &ggql.Ref{Base: ggql.Base{N: "User"}},
		SelBase: ggql.SelBase{
			Sels: []ggql.Selection{
				&ggql.Field{Name: "height"},
			},
		},
	}
	checkEqual(t, `... on User {
  height
}`, in.String(), "Inline.String() mismatch")
}
