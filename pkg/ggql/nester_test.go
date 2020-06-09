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
	"fmt"
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

type nest struct {
	buf *strings.Builder
}

func (n *nest) Nest(f *ggql.Field) interface{} {
	n.buf.WriteString(fmt.Sprintf("nest.Nest(%s)\n", f.Name))
	return &nest{buf: n.buf}
}

func TestNester(t *testing.T) {
	root := setupTestSongs(t, nil)

	exe, err := root.ParseExecutableString(`{artists{...Frag}}
fragment Frag on Artist {name,songs{...{name}}}
`)
	checkNil(t, err, "ParseExecutableString should not fail. %s", err)

	var b strings.Builder
	n := nest{buf: &b}

	exe.SetContextRecursive(&n)

	expect := `nest.Nest(artists)
nest.Nest(name)
nest.Nest(songs)
nest.Nest(name)
`
	checkEqual(t, expect, b.String(), "nest output mismatch")
}
