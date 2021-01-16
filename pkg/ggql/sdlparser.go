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

package ggql

import (
	"fmt"
	"io"
)

type sdlParser struct {
	parser
}

func parseSDL(root *Root, reader io.Reader) (types []Type, extends []*Extend, err error) {
	p := sdlParser{parser: parser{root: root, reader: reader}}

	var token string
	var b byte
	var x *Extend

	if err = p.skipBOM(); err != nil {
		return
	}
	for {
		if p.eof {
			break
		}
		var t Type
		if b, err = p.skipSpace(); err != nil {
			break // drop down to AddTypes
		}
		var desc string
		if b == '"' {
			// Read a description and continue.
			desc, err = p.readDesc()
		}
	TOP:
		if err == nil {
			token, err = p.readToken()
		}
		if err == nil && 0 < len(token) {
			switch token {
			case directiveStr:
				t, err = p.readDirective(desc)
			case enumStr:
				t, err = p.readEnum(desc)
			case extendStr:
				x = &Extend{}
				goto TOP
			case inputStr:
				t, err = p.readInput(desc)
			case interfaceStr:
				t, err = p.readInterface(desc)
			case scalarStr:
				t, err = p.readScalar(desc)
			case schemaStr:
				t, err = p.readSchema(x != nil)
			case typeStr:
				t, err = p.readObject(desc)
			case unionStr:
				t, err = p.readUnion(desc)
			default:
				err = fmt.Errorf("%w, '%s' is not a valid schema directive at %d:%d", ErrParse, token, p.line, p.col)
			}
		}
		if err != nil {
			break
		}
		if t != nil {
			if x != nil {
				x.Adds = t
				extends = append(extends, x)
				x = nil
			} else {
				types = append(types, t)
			}
		}
	}
	return
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Type-System.Directives
func (p *sdlParser) readDirective(desc string) (Type, error) {
	b, err := p.skipSpace()
	if err == nil && b != '@' {
		err = fmt.Errorf("%w, expected a @ at %d:%d", ErrParse, p.line, p.col)
	}
	var token string
	if err == nil {
		_, _ = p.readByte() // re-read @
		token, err = p.readToken()
	}
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no directive name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var dir *Directive
	if err == nil {
		dir = &Directive{Base: Base{N: token, Desc: desc, line: p.line, col: p.col - len(token) - 2}}
		err = p.readArgs(&dir.args)
	}
	if err == nil {
		token, err = p.readToken()
	}
	if err == nil && token != "on" {
		err = fmt.Errorf("%w, directives must have an 'on' keyword at %d:%d", ErrParse, p.line, p.col)
	}
	if err == nil {
		for {
			if b, err = p.skipSpace(); err != nil {
				break
			}
			if b == '|' {
				_, _ = p.readByte() // re-read |
			} else if 0 < len(dir.On) {
				break
			}
			if token, err = p.readToken(); err != nil {
				return nil, err
			}
			if 0 < len(token) {
				dir.On = append(dir.On, Location(token))
			} else {
				break
			}
		}
	}
	return dir, err
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Enums
func (p *sdlParser) readEnum(desc string) (Type, error) {
	token, err := p.readToken()
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no enum name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var enum *Enum
	if err == nil {
		enum = &Enum{Base: Base{N: token, Desc: desc, line: p.line, col: p.col - len(token) - 1}}
		enum.Dirs, err = p.readDirUses()
	}
	var b byte
	if err == nil {
		b, err = p.skipSpace()
	}
	if err == nil && b != '{' {
		err = fmt.Errorf("%w, expected { at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read {
	if err == nil {
		var ev *EnumValue
		for {
			b, err = p.skipSpace()
			if b == '}' {
				_, _ = p.readByte() // re-read }
				break
			}
			if err == nil {
				ev, err = p.readEnumValue()
				if err == nil {
					err = enum.values.add(ev)
				}
			}
			if err != nil {
				break
			}
		}
	}
	return enum, err
}

// https://graphql.github.io/graphql-spec/June2018/#EnumValueDefinition
func (p *sdlParser) readEnumValue() (ev *EnumValue, err error) {
	var desc string
	var token string

	if desc, err = p.readDesc(); err == nil {
		token, err = p.readToken()
	}
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, invalid enum value name at %d:%d", ErrParse, p.line, p.col)
	}
	if err == nil {
		ev = &EnumValue{Value: Symbol(token), Description: desc, line: p.line, col: p.col - len(token) - 1}
		var du *DirectiveUse
		for {
			if du, err = p.readDirUse(); du == nil {
				break
			}
			ev.Directives = append(ev.Directives, du)
		}
	}
	return
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Input-Objects
func (p *sdlParser) readInput(desc string) (Type, error) {
	token, err := p.readToken()
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no input name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var input *Input
	if err == nil {
		input = &Input{
			Base: Base{
				N:    token,
				Desc: desc,
				line: p.line,
				col:  p.col - len(token) - 1,
			},
			fields: inputFieldList{
				dict: map[string]*InputField{},
			},
		}
		input.Dirs, err = p.readDirUses()
	}
	var b byte
	if err == nil {
		b, err = p.skipSpace()
	}
	if err == nil && b != '{' {
		err = fmt.Errorf("%w, expected { at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read {
	if err == nil {
		// Validation phase will ensure that only fields valid for an input object
		// were specified here.
		// https://graphql.github.io/graphql-spec/June2018/#sec-Input-Objects:
		err = p.readInputFields(&input.fields)
	}
	return input, err
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Interfaces
func (p *sdlParser) readInterface(desc string) (Type, error) {
	token, err := p.readToken()
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no interface name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var inf *Interface
	if err == nil {
		inf = &Interface{
			Base: Base{
				N:    token,
				Desc: desc,
				line: p.line,
				col:  p.col - len(token) - 1,
			},
			fields: fieldList{
				dict: map[string]*FieldDef{},
			},
			Root: p.root,
		}
		inf.Dirs, err = p.readDirUses()
	}
	var b byte
	if err == nil {
		b, err = p.skipSpace()
	}
	if err == nil && b != '{' {
		err = fmt.Errorf("%w, expected { at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read {
	if err == nil {
		err = p.readFields(&inf.fields)
	}
	return inf, err
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Scalars
func (p *sdlParser) readScalar(desc string) (Type, error) {
	token, err := p.readToken()
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no scalar name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var scalar *stringScalar
	if err == nil {
		scalar = &stringScalar{
			Scalar: Scalar{
				Base: Base{
					N:    token,
					Desc: desc,
					line: p.line,
					col:  p.col - len(token),
				},
			},
		}
		scalar.Dirs, err = p.readDirUses()
	}
	if err != nil {
		scalar = nil
	}
	return scalar, err
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Schema
func (p *sdlParser) readSchema(extend bool) (Type, error) {
	schema := &Schema{
		Object: Object{
			Base: Base{
				line: p.line,
				col:  p.col,
			},
			fields: fieldList{
				dict: map[string]*FieldDef{},
			},
		},
	}
	var err error
	schema.Dirs, err = p.readDirUses()

	var b byte
	if err == nil {
		b, err = p.skipSpace()
	}
	if err == nil && b != '{' {
		err = fmt.Errorf("%w, expected { at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read {
	if err == nil {
		err = p.readFields(&schema.fields)
	}
	if err == nil && !extend {
		p.root.schema = schema
	}
	return schema, err
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Objects
func (p *sdlParser) readObject(desc string) (Type, error) {
	token, err := p.readToken()
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no type name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var obj *Object
	if err == nil {
		obj = &Object{
			Base: Base{
				N:    token,
				Desc: desc,
				line: p.line,
				col:  p.col - len(token) - 1,
			},
			fields: fieldList{
				dict: map[string]*FieldDef{},
			},
		}
		obj.Interfaces, err = p.readImplements()
	}
	if err == nil {
		obj.Dirs, err = p.readDirUses()
	}
	var b byte
	if err == nil {
		b, err = p.skipSpace()
	}
	if err == nil && b != '{' {
		err = fmt.Errorf("%w, expected { at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read {
	if err == nil {
		err = p.readFields(&obj.fields)
	}
	return obj, err
}

// https://graphql.github.io/graphql-spec/June2018/#sec-Unions
func (p *sdlParser) readUnion(desc string) (Type, error) {
	token, err := p.readToken()
	if err == nil && len(token) == 0 {
		err = fmt.Errorf("%w, no union name provided at %d:%d", ErrParse, p.line, p.col)
	}
	var union *Union
	if err == nil {
		union = &Union{Base: Base{N: token, Desc: desc, line: p.line, col: p.col - len(token) - 1}}
		union.Dirs, err = p.readDirUses()
	}
	var b byte
	if err == nil {
		b, err = p.skipSpace()
	}
	if err == nil && b != '=' {
		err = fmt.Errorf("%w, expected = at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read =
	if err == nil {
		for {
			if b, err = p.skipSpace(); err != nil {
				break
			}
			if b == '|' {
				_, _ = p.readByte() // re-read |
			} else if 0 < len(union.Members) {
				break
			}
			t, _ := p.readType()
			if t == nil {
				break
			}
			union.Members = append(union.Members, t)
		}
	}
	return union, err
}

// https://graphql.github.io/graphql-spec/June2018/#Arguments
func (p *sdlParser) readArgs(args *argList) (err error) {
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return
	}
	if b != '(' {
		return
	}
	_, _ = p.readByte() // re-read (
	var a *Arg
	for {
		if b, err = p.skipSpace(); err != nil {
			return
		}
		switch b {
		case ')':
			_, _ = p.readByte() // re-read )
			return
		case 0:
			return fmt.Errorf("%w, arguments not closed with a ')' at %d:%d", ErrParse, p.line, p.col)
		default:
			if a, err = p.readArg(); err != nil {
				return
			}
			if err = args.add(a); err != nil {
				return
			}
		}
	}
}

// https://graphql.github.io/graphql-spec/June2018/#Argument
func (p *sdlParser) readArg() (arg *Arg, err error) {
	var desc string
	if desc, err = p.readDesc(); err != nil {
		return
	}
	var token string
	if token, err = p.readToken(); err != nil {
		return
	}
	arg = &Arg{Base: Base{N: token, Desc: desc, line: p.line, col: p.col - len(token) - 1}}

	var b byte
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b != ':' {
		return nil, fmt.Errorf("%w, argument name not followed by a : at %d:%d", ErrParse, p.line, p.col)
	}
	_, _ = p.readByte() // re-read :

	if arg.Type, err = p.readType(); err != nil {
		return
	}
	if arg.Type == nil {
		return nil, fmt.Errorf("%w, argument type missing at %d:%d", ErrParse, p.line, p.col)
	}
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b == '=' {
		_, _ = p.readByte() // re-read =
		if arg.Default, err = p.readValue(); err != nil {
			return
		}
	}
	if err == nil {
		var du *DirectiveUse
		for {
			if du, err = p.readDirUse(); du == nil {
				break
			}
			arg.Dirs = append(arg.Dirs, du)
		}
	}
	return
}

// https://graphql.github.io/graphql-spec/June2018/#FieldsDefinition
func (p *sdlParser) readFields(fields *fieldList) (err error) {
	var f *FieldDef
	var b byte
	for {
		b, err = p.skipSpace()
		if b == '}' {
			_, _ = p.readByte() // re-read }
			break
		}
		if err == nil {
			if b == 0 {
				err = fmt.Errorf("%w, fields not terminated with a '}' at %d:%d", ErrParse, p.line, p.col)
				break
			}
		}
		if err == nil {
			f, err = p.readField()
		}
		if err == nil {
			if f == nil {
				err = fmt.Errorf("%w, missing field at %d:%d", ErrParse, p.line, p.col)
				break
			}
			if err = fields.add(f); err != nil {
				return
			}
		}
		if err != nil {
			break
		}
	}
	return
}

// https://graphql.github.io/graphql-spec/June2018/#FieldDefinition
func (p *sdlParser) readField() (f *FieldDef, err error) {
	var desc string
	var token string
	var b byte

	if desc, err = p.readDesc(); err == nil {
		token, err = p.readToken()
	}
	f = &FieldDef{Base: Base{N: token, Desc: desc, line: p.line, col: p.col - len(token) - 1}}
	if err == nil {
		b, err = p.skipSpace()
	}
	if b == 0 {
		return nil, nil
	}
	if err == nil {
		if b == '(' {
			if err = p.readArgs(&f.args); err == nil {
				b, err = p.skipSpace()
			}
		}
	}
	if err == nil && b != ':' {
		err = fmt.Errorf("%w, field name not followed by a : at %d:%d", ErrParse, p.line, p.col)
	}
	if err == nil {
		_, _ = p.readByte() // re-read :
		if f.Type, err = p.readType(); f.Type == nil {
			err = fmt.Errorf("%w, field type missing at %d:%d", ErrParse, p.line, p.col)
		}
	}
	if err == nil {
		var du *DirectiveUse
		for {
			if du, err = p.readDirUse(); du == nil {
				break
			}
			f.Dirs = append(f.Dirs, du)
		}
	}
	if err != nil {
		f = nil
	}
	return
}

func (p *sdlParser) readInputFields(fields *inputFieldList) (err error) {
	var f *InputField
	var b byte
	for {
		b, err = p.skipSpace()
		if b == '}' {
			_, _ = p.readByte() // re-read }
			break
		}
		if err == nil {
			if b == 0 {
				err = fmt.Errorf("%w, fields not terminated with a '}' at %d:%d", ErrParse, p.line, p.col)
			}
		}
		if err == nil {
			f, err = p.readInputField()
		}
		if err == nil {
			if f == nil {
				err = fmt.Errorf("%w, missing field at %d:%d", ErrParse, p.line, p.col)
			} else {
				err = fields.add(f)
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func (p *sdlParser) readInputField() (f *InputField, err error) {
	var desc string
	var token string
	var b byte

	if desc, err = p.readDesc(); err == nil {
		token, err = p.readToken()
	}
	if err == nil {
		b, err = p.skipSpace()
	}
	if b == 0 {
		return nil, nil
	}
	f = &InputField{Base: Base{N: token, Desc: desc, line: p.line, col: p.col - len(token) - 1}}
	if err == nil && b != ':' {
		err = fmt.Errorf("%w, field name not followed by a ':' at %d:%d", ErrParse, p.line, p.col)
	}
	if err == nil {
		_, _ = p.readByte() // re-read :
		if f.Type, err = p.readType(); f.Type == nil && err == nil {
			err = fmt.Errorf("%w, field type missing at %d:%d", ErrParse, p.line, p.col)
		}
	}
	if err == nil {
		b, err = p.skipSpace()
	}
	if b == '=' {
		_, _ = p.readByte() // re-read =
		f.Default, err = p.readValue()
	}
	if err == nil {
		var du *DirectiveUse
		for {
			if du, err = p.readDirUse(); du == nil {
				break
			}
			f.Dirs = append(f.Dirs, du)
		}
	}
	if err != nil {
		f = nil
	}
	return
}

// https://graphql.github.io/graphql-spec/June2018/#ImplementsInterfaces
func (p *sdlParser) readImplements() (interfaces []Type, err error) {
	var b byte
	if b, err = p.skipSpace(); err != nil || b != 'i' {
		return
	}
	var token string
	token, err = p.readToken()
	if token != "implements" {
		err = fmt.Errorf("%w, expected 'implements' at %d:%d", ErrParse, p.line, p.col)
	}
	if err == nil {
		if b, err = p.skipSpace(); b == '&' {
			_, _ = p.readByte() // re-read &
		}
	}
	var t Type
	for err == nil {
		b, err = p.skipSpace()
		if 0 < len(interfaces) {
			if b != '&' {
				break
			}
			_, _ = p.readByte() // re-read &
		}
		if err == nil {
			if t, err = p.readType(); t != nil {
				interfaces = append(interfaces, t)
			} else {
				break
			}
		}
	}
	if len(interfaces) == 0 && err == nil {
		err = fmt.Errorf("%w, 'implements' keyword but no interfaces listed at %d:%d", ErrParse, p.line, p.col)
	}
	return
}
