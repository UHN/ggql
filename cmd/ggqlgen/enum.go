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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uhn/ggql/pkg/ggql"
)

func stubEnum(t *ggql.Enum) (err error) {
	var b strings.Builder

	fmt.Fprintf(&b, "package %s\n\n", pkg)
	b.WriteString("const (\n")
	for _, ev := range t.Values() {
		name := t.Name() + string(ev.Value)
		desc := ev.Description
		if len(desc) == 0 {
			desc = dotdotdot
		}
		if strings.HasPrefix(desc, name+" ") {
			fmt.Fprintf(&b, "\t// %s\n", strings.ReplaceAll(desc, "\n", "\n\t// "))
		} else {
			fmt.Fprintf(&b, "\t// %s %s\n", name, strings.ReplaceAll(desc, "\n", "\n\t// "))
		}
		fmt.Fprintf(&b, "\t%s = %q\n", name, string(ev.Value))
	}
	b.WriteString(")\n")

	path := filepath.Join(stubDir, strings.ToLower(t.Name())+".go")

	return os.WriteFile(path, []byte(b.String()), 0600)
}
