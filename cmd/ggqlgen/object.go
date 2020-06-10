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

func stubObject(t *ggql.Object) (err error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", pkg))
	b.WriteString("import (\n")
	if reflect {
		for _, f := range t.Fields() {
			if 0 < len(f.Args()) {
				b.WriteString("\t\"fmt\"\n")
				break
			}
		}
	} else {
		b.WriteString("\t\"fmt\"\n")
	}
TOP:
	for _, f := range t.Fields() {
		if f.Type.Name() == timeStr {
			b.WriteString("\t\"time\"\n")
			break
		}
		for _, a := range f.Args() {
			if a.Type.Name() == timeStr {
				b.WriteString("\t\"time\"\n")
				break TOP
			}
		}
	}
	if !reflect {
		b.WriteString("\n\t\"github.com/uhn/ggql/pkg/ggql\"\n")
	}
	b.WriteString(")\n\n")

	if err = stubObjectStruct(&b, t); err != nil {
		return
	}
	if err = stubObjectFuncs(&b, t); err != nil {
		return
	}
	if !reflect {
		if err = stubObjectResolve(&b, t); err != nil {
			return
		}
	}
	path := filepath.Join(stubDir, strings.ToLower(t.Name())+".go")

	return ioutil.WriteFile(path, []byte(b.String()), 0600)
}

func stubObjectStruct(b *strings.Builder, t *ggql.Object) (err error) {
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
		if 0 < len(f.Args()) {
			continue
		}
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
	b.WriteString("}\n\n")

	return
}

func stubObjectFuncs(b *strings.Builder, t *ggql.Object) (err error) {
	for _, f := range t.Fields() {
		if len(f.Args()) == 0 {
			continue
		}
		public := publicName(f.N)
		desc := f.Desc
		if len(desc) == 0 {
			desc = dotdotdot
		}
		if strings.HasPrefix(desc, public+" ") {
			b.WriteString(fmt.Sprintf("// %s\n", strings.ReplaceAll(desc, "\n", "\n\t// ")))
		} else {
			b.WriteString(fmt.Sprintf("// %s %s\n", public, strings.ReplaceAll(desc, "\n", "\n\t// ")))
		}
		b.WriteString(fmt.Sprintf("func (t *%s) %s(", t.Name(), public))
		for i, a := range f.Args() {
			if 0 < i {
				b.WriteString(fmt.Sprintf(", %s %s", a.N, typeStr(a.Type)))
			} else {
				b.WriteString(fmt.Sprintf("%s %s", a.N, typeStr(a.Type)))
			}
		}
		b.WriteString(fmt.Sprintf(") (result %s, err error) {\n\n", typeStr(f.Type)))

		b.WriteString("\t// FIXME\n")
		b.WriteString("\terr = fmt.Errorf(\"not implemented yet\")\n\n")

		b.WriteString("\treturn\n")
		b.WriteString("}\n\n")
	}
	return
}

func stubObjectResolve(b *strings.Builder, t *ggql.Object) (err error) {
	b.WriteString("// Resolve a field into a value.\n")
	b.WriteString(fmt.Sprintf("func (t *%s) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {\n", t.Name()))
	b.WriteString("\tswitch field.Name {\n")
	for _, f := range t.Fields() {
		public := publicName(f.Name())
		b.WriteString(fmt.Sprintf("\tcase \"%s\":\n", f.Name()))
		if 0 < len(f.Args()) {
			for _, a := range f.Args() {
				b.WriteString(fmt.Sprintf("\t\t%s, _ := args[\"%s\"].(%s)\n", a.Name(), a.Name(), typeStr(a.Type)))
			}
			b.WriteString(fmt.Sprintf("\n\t\treturn t.%s(", public))
			for i, a := range f.Args() {
				if 0 < i {
					b.WriteString(fmt.Sprintf(", %s", a.Name()))
				} else {
					b.WriteString(a.Name())
				}
			}
			b.WriteString(")\n")
		} else {
			b.WriteString(fmt.Sprintf("\t\treturn t.%s, nil\n", public))
		}
	}
	b.WriteString("\t}\n")

	b.WriteString(fmt.Sprintf("\treturn nil, fmt.Errorf(\"type %s does not have field %%s\", field)\n", t.Name()))
	b.WriteString("}\n")

	return
}
