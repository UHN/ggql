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

func TestDirective(t *testing.T) {
	root := ggql.NewRoot(nil)

	typeArg := ggql.Arg{
		Base: ggql.Base{
			N: "type",
		},
		Type: &ggql.NonNull{
			Base: root.GetType("String"),
		},
	}
	xDir := ggql.Directive{
		Base: ggql.Base{
			N:    "example",
			Desc: "Associates a GraphQLtype with a golang type.",
		},
		On: []ggql.Location{ggql.LocObject, ggql.LocSchema},
	}
	xDir.AddArg(&typeArg)

	err := root.AddTypes(&xDir)
	checkNil(t, err, "root.AddTypes failed. %s", err)
	actual := root.SDL(false, true)
	expectStr := "directive @example(type: String!) on OBJECT | SCHEMA"
	expect := timeScalarSDL + `
"Associates a GraphQLtype with a golang type."
` + expectStr + "\n"

	checkEqual(t, expect, actual, "Directive SDL() mismatch")
	checkEqual(t, expectStr, xDir.String(), "Directive String() mismatch")
	checkEqual(t, 11, xDir.Rank(), "Directive Rank() mismatch")
	checkEqual(t, "Associates a GraphQLtype with a golang type.", xDir.Description(), "Directive Description() mismatch")

	w := &failWriter{max: len(expectStr) - 5}
	err = xDir.Write(w, false)
	checkNotNil(t, err, "return error on write error")

	checkNotNil(t, xDir.Extend(nil), "return error on extend attempt")
}

func TestDirectiveBadRef(t *testing.T) {
	root := ggql.NewRoot(nil)

	dir := ggql.Directive{
		Base: ggql.Base{N: "example"},
		On:   []ggql.Location{ggql.LocObject, ggql.LocSchema},
	}
	dir.AddArg(&ggql.Arg{Base: ggql.Base{N: "bad"}, Type: &ggql.Ref{Base: ggql.Base{N: "Bad"}}})

	err := root.AddTypes(&dir)
	checkNotNil(t, err, "root.AddTypes with bad ref should fail.")

	dir = ggql.Directive{
		Base: ggql.Base{N: "example"},
		On:   []ggql.Location{ggql.LocObject, ggql.LocSchema},
	}
	dir.AddArg(&ggql.Arg{
		Base: ggql.Base{N: "bad", Dirs: []*ggql.DirectiveUse{{Directive: &ggql.Ref{Base: ggql.Base{N: "Bad"}}}}},
		Type: root.GetType("String"),
	})

	err = root.AddTypes(&dir)
	checkNotNil(t, err, "root.AddTypes with bad arg directive ref should fail.")
}
