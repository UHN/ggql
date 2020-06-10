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
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/uhn/ggql/pkg/ggql"
)

func stubInput(t *ggql.Input) (err error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", pkg))
	for _, f := range t.Fields() {
		if f.Type.Name() == timeStr {
			b.WriteString("import \"time\"\n\n")
			break
		}
	}
	desc := t.Description()
	if len(desc) == 0 {
		desc = dotdotdot
	}
	if strings.HasPrefix(desc, t.Name()+" ") {
		b.WriteString(fmt.Sprintf("// %s\n", strings.ReplaceAll(desc, "\n", "\n// ")))
	} else {
		b.WriteString(fmt.Sprintf("// %s %s\n", t.Name(), strings.ReplaceAll(desc, "\n", "\n// ")))
	}
	b.WriteString(fmt.Sprintf("type %s struct {\n", t.Name()))
	for _, f := range t.Fields() {
		public := publicName(f.N)
		desc = f.Desc
		if len(desc) == 0 {
			desc = dotdotdot
		}
		if strings.HasPrefix(desc, public+" ") {
			b.WriteString(fmt.Sprintf("\n\t// %s\n", strings.ReplaceAll(desc, "\n", "\n\t// ")))
		} else {
			b.WriteString(fmt.Sprintf("\n\t// %s %s\n", public, strings.ReplaceAll(desc, "\n", "\n\t// ")))
		}
		b.WriteString(fmt.Sprintf("\t%s %s\n", public, typeStr(f.Type)))
	}
	b.WriteString("}\n")
	path := filepath.Join(stubDir, strings.ToLower(t.Name())+".go")

	return ioutil.WriteFile(path, []byte(b.String()), 0600)
}
