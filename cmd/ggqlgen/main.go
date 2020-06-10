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
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/uhn/ggql/pkg/ggql"
)

const (
	dotdotdot = "..."
	timeStr   = "Time"
)

// Runtime variables.
var (
	verbose = false
	stubDir = ""
	reflect = false
	pkg     = "main"
	embeds  = embedsValue{}
	overs   = oversValue{}
)

func init() {
	flag.BoolVar(&verbose, "v", verbose, "verbose output prints the schema")
	flag.StringVar(&pkg, "p", pkg, "package for output")

	flag.StringVar(&stubDir, "s", stubDir, "directory to write stub files to. If not provided no stubs are written")
	flag.BoolVar(&reflect, "r", reflect, "generate reflection based stubs vs the default of interface based stubs")
	flag.Var(&embeds, "e", "src-file:embed-file:embed-name")
	flag.Var(&overs, "w", "overwrite the input file with a re-formatted version")
}

type embed struct {
	src   string
	dest  string
	name  string
	types map[string]bool
}

type embedsValue struct {
	embeds []*embed
}

func (ev *embedsValue) String() string {
	var b []byte
	for i, e := range ev.embeds {
		if 0 < i {
			b = append(b, ' ')
		}
		b = append(b, e.src...)
		b = append(b, ':')
		b = append(b, e.dest...)
		b = append(b, ':')
		b = append(b, e.name...)
	}
	return string(b)
}

func (ev *embedsValue) Set(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return fmt.Errorf("expect format src-file:dest-file:var-name")
	}
	ev.embeds = append(ev.embeds, &embed{src: parts[0], dest: parts[1], name: parts[2], types: map[string]bool{}})
	return nil
}

type over struct {
	file  string
	types map[string]bool
}

type oversValue struct {
	overs []*over
}

func (ov *oversValue) String() string {
	var b []byte
	for i, o := range ov.overs {
		if 0 < i {
			b = append(b, ' ')
		}
		b = append(b, o.file...)
	}
	return string(b)
}

func (ov *oversValue) Set(s string) error {
	ov.overs = append(ov.overs, &over{file: s, types: map[string]bool{}})
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `

ggqlgen generates go stub files for a GraphQL schema. It can also generate a
go file with the schema encoded as a string that can be used as an embedded
schema for an application. If no schema file is specified ggqlgen will try
to read the schema from stdin. If the specified schema file is a .go file
containing a single declaration which is a string constant defined using
backtick characters, ggqlgen will parse the file and use the SDL defined in
that constant as the input SDL.

If neither a stub directory nor embedded const name is provided only a
validation is performed.

Stub files are created for input, object, and enum types. The stub files
should build in most cases but are not meant to be a complete
application. Some lint errors are to be expected.

Usage: ggqlgen [options] [<schema-file>...]

`)
		flag.PrintDefaults()
	}
	flag.Parse()
	ggql.Sort = true

	var filepath string

	root := ggql.NewRoot(nil)
	exists := map[string]bool{}
	for _, t := range root.Types() {
		if t.Core() {
			continue
		}
		exists[t.Name()] = true
	}
	var files []string

	files = append(files, flag.Args()...)
	for _, o := range overs.overs {
		files = append(files, o.file)
	}
	if len(files) == 0 {
		if err := root.ParseReader(os.Stdin); err != nil {
			log.Fatalf("Failed to parse stdin: %s", err)
		}
	} else {
		for _, filepath = range files {
			var e *embed
			for _, e2 := range embeds.embeds {
				if filepath == e2.src {
					e = e2
					break
				}
			}
			var o *over
			for _, o2 := range overs.overs {
				if filepath == o2.file {
					o = o2
					break
				}
			}
			sdl, err := getSDL(filepath)
			if err != nil {
				log.Fatalf("Failed to read schema file %s: %s", filepath, err)
			}
			if err = root.Parse(sdl); err != nil {
				log.Fatalf("Failed to parse file %s: %s", filepath, err)
			}
			for _, t := range root.Types() {
				if t.Core() || exists[t.Name()] {
					continue
				}
				if e != nil {
					e.types[t.Name()] = true
				}
				if o != nil {
					o.types[t.Name()] = true
				}
				exists[t.Name()] = true
			}
		}
	}
	for _, e := range embeds.embeds {
		var buf []byte
		buf = append(buf, "package "...)
		buf = append(buf, pkg...)
		buf = append(buf, "\n\nconst "...)
		buf = append(buf, e.name...)
		buf = append(buf, " = `"...)
		for _, t := range root.Types() {
			if t.Core() {
				continue
			}
			if e.types[t.Name()] {
				buf = append(buf, '\n')
				buf = append(buf, t.SDL(true)...)
			}
		}
		buf = append(buf, "`\n"...)
		if 0 < len(e.dest) {
			if err := ioutil.WriteFile(e.dest, buf, 0600); err != nil {
				log.Fatalf("Failed to write embedded file %s. %s", e.dest, err)
			}
		} else {
			fmt.Println(string(buf))
		}
	}
	if 0 < len(stubDir) {
		if err := stubGen(root); err != nil {
			log.Fatalf("Failed to generate stub file in %s: %s", stubDir, err)
		}
	}
	for _, o := range overs.overs {
		var buf []byte
		for _, t := range root.Types() {
			if t.Core() {
				continue
			}
			if o.types[t.Name()] {
				buf = append(buf, '\n')
				buf = append(buf, t.SDL(true)...)
			}
		}
		if err := ioutil.WriteFile(o.file, buf, 0600); err != nil {
			log.Fatalf("Failed to overwrite %s: %s", o.file, err)
		}
	}
	if verbose {
		fmt.Println(root.SDL(false, true))
	}
}

// getSDL returns the sdl as a byte slice or an error. If the file path has a
// .go extension, will try to parse the AST, and if there is only a single
// declaration and it's a string constant defined with backticks
// (e.g. const SDL = `my-sdl`), use that as the SDL.
func getSDL(path string) ([]byte, error) {
	if ext := filepath.Ext(path); ext != ".go" {
		return ioutil.ReadFile(path)
	}
	// If path is a go file, parse the AST
	fset := token.NewFileSet()
	var f *ast.File
	var err error
	if f, err = parser.ParseFile(fset, path, nil, parser.AllErrors); err != nil {
		return nil, err
	}
	if len(f.Decls) != 1 {
		return nil, fmt.Errorf("cannot use file with more than one declaration in it")
	}
	decl := f.Decls[0]
	err = fmt.Errorf("sdl declaration must be a const string with backticks")
	gen, ok := decl.(*ast.GenDecl)
	if !ok || gen.Tok != token.CONST {
		return nil, err
	}
	for _, spec := range gen.Specs {
		if vspec, ok := spec.(*ast.ValueSpec); ok {
			if literal, ok := vspec.Values[0].(*ast.BasicLit); ok && literal.Kind == token.STRING && literal.Value[0] == '`' {
				// Strip the literal of its backtics and set as the SDL
				return []byte(strings.Trim(literal.Value, "`")), nil
			}
			break
		}
	}
	return nil, err
}

func stubGen(root *ggql.Root) (err error) {
	if err = os.MkdirAll(stubDir, 0750); err != nil {
		return
	}
	var b strings.Builder

	b.WriteString(fmt.Sprintf("// Package %s was generated by ggqlgen.\n", pkg))
	b.WriteString(fmt.Sprintf("package %s\n", pkg))
	path := filepath.Join(stubDir, "doc.go")

	if err = ioutil.WriteFile(path, []byte(b.String()), 0600); err != nil {
		return err
	}
	for _, t := range root.Types() {
		if t.Core() {
			continue
		}
		switch tt := t.(type) {
		case *ggql.Object:
			err = stubObject(tt)
		case *ggql.Input:
			err = stubInput(tt)
		case *ggql.Enum:
			err = stubEnum(tt)
		}
		if err != nil {
			break
		}
	}
	return
}
