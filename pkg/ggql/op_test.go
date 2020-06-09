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

func TestOp(t *testing.T) {
	hello := ggql.Field{
		Name: "hello",
		Args: []*ggql.ArgValue{{Arg: "name", Value: ggql.Var("name")}},
	}
	op := ggql.Op{
		SelBase: ggql.SelBase{Sels: []ggql.Selection{&hello}},
		Type:    ggql.OpQuery,
	}
	checkEqual(t, `query {
  hello(name: $name)
}
`, op.String(), "incorrect output for blank op")

	op.Name = "greeting"
	checkEqual(t, `query greeting {
  hello(name: $name)
}
`, op.String(), "incorrect output for named op")

	op.Variables = []*ggql.VarDef{{Name: "name", Type: &ggql.Ref{Base: ggql.Base{N: "String"}}, Default: "sailor"}}
	checkEqual(t, `query greeting($name: String = "sailor") {
  hello(name: $name)
}
`, op.String(), "incorrect output for op with variables")

	hello.Alias = "hi"
	checkEqual(t, `query greeting($name: String = "sailor") {
  hi: hello(name: $name)
}
`, op.String(), "incorrect output for op with variables")

	dirt := ggql.Directive{
		Base: ggql.Base{
			N: "dirt",
		},
		On: []ggql.Location{ggql.LocArgumentDefinition, ggql.LocVariableDefinition},
	}
	op.Dirs = []*ggql.DirectiveUse{{Directive: &dirt}}

	checkEqual(t, `query greeting($name: String = "sailor") @dirt {
  hi: hello(name: $name)
}
`, op.String(), "incorrect output for op with variables and directives")

	op.Variables[0].Dirs = []*ggql.DirectiveUse{{Directive: &dirt}}
	checkEqual(t, `query greeting($name: String = "sailor" @dirt) @dirt {
  hi: hello(name: $name)
}
`, op.String(), "incorrect output for op with variables and more directives")

	op.Variables = append(op.Variables, &ggql.VarDef{Name: "age", Type: &ggql.Ref{Base: ggql.Base{N: "Int"}}})
	checkEqual(t, `query greeting($name: String = "sailor" @dirt, $age: Int) @dirt {
  hi: hello(name: $name)
}
`, op.String(), "incorrect output for op with variables and directives")
}
