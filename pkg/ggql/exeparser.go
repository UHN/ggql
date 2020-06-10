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
	"io"
)

type exeParser struct {
	parser
	exe *Executable
}

func parseExe(root *Root, reader io.Reader) (exe *Executable, err error) {
	exe = &Executable{Root: root, Ops: map[string]*Op{}}

	p := exeParser{parser: parser{root: root, reader: reader}, exe: exe}

	if err = p.skipBOM(); err != nil {
		return
	}
	var token string

	for {
		if p.eof || err != nil {
			break
		}
		if _, err = p.skipSpace(); err != nil {
			break
		}
		var op *Op
		token, err = p.readToken()
		if err == nil {
			switch token {
			case string(OpQuery), string(OpMutation), string(OpSubscription):
				op, err = p.readOp(OpType(token))
				if exe.Ops[op.Name] != nil {
					err = parseError(op.line, op.col, "duplicate '%s' operation", op.Name)
				} else {
					exe.Ops[op.Name] = op
				}
			case "fragment":
				var frag *Fragment
				if frag, err = p.readFragmentDef(); err == nil {
					if exe.Fragments == nil {
						exe.Fragments = map[string]*Fragment{frag.Name: frag}
					} else if existing := exe.Fragments[frag.Name]; existing != nil {
						if 0 < len(existing.Sels) {
							err = parseError(frag.line, frag.col, "duplicate '%s' fragment", frag.Name)
						} else {
							existing.Inline = frag.Inline
						}
					} else {
						exe.Fragments[frag.Name] = frag
					}
				}
			case "":
				if p.onDeck != '{' && p.onDeck != 0 {
					err = parseError(p.line, p.col, "expected an operation")
					continue
				}
				op = &Op{SelBase: SelBase{line: p.line, col: p.col}, Type: OpQuery}
				op.Sels, err = p.readSelectionSet()
				if len(op.Sels) == 0 {
					break
				}
				if exe.Ops[""] != nil {
					err = parseError(op.line, op.col, "duplicate un-named operation")
				} else {
					exe.Ops[""] = op
				}
			default:
				err = parseError(p.line, p.col-len(token), "'%s' is not a valid executable operation type", token)
			}
		}
	}
	return
}

func (p *exeParser) readOp(opType OpType) (op *Op, err error) {
	op = &Op{Type: opType, SelBase: SelBase{line: p.line, col: p.col}}

	if _, err = p.skipSpace(); err == nil {
		op.col = p.col
		op.Name, err = p.readToken()
	}
	if err == nil {
		op.Variables, err = p.readVarDefs()
	}
	if err == nil {
		op.Dirs, err = p.readDirUses()
	}
	if err == nil {
		op.Sels, err = p.readSelectionSet()
	}
	return
}

func (p *exeParser) readSelectionSet() (sels []Selection, err error) {
	var b byte
	b, err = p.skipSpace()
	if (b != '{' || b == 0) && err == nil {
		// No selection set
		return
	}
	_, _ = p.readByte() // re-read {
FOR:
	for {
		if err != nil {
			break
		}
		b, err = p.skipSpace()
		switch b {
		case 0:
			return nil, parseError(p.line, p.col, "selection set not terminated with a '}'")
		case '}':
			_, _ = p.readByte() // re-read }
			break FOR
		case '.':
			var sel Selection
			if sel, err = p.readFragment(); err == nil {
				sels = append(sels, sel)
			}
		default: // a Field
			if err == nil {
				var f *Field
				if f, err = p.readField(); err == nil {
					sels = append(sels, f)
				}
			}
		}
	}
	return
}

func (p *exeParser) readField() (f *Field, err error) {
	var b byte
	var token string

	token, err = p.readToken()
	if len(token) == 0 && err == nil {
		err = parseError(p.line, p.col, "a field name can not be blank")
	}
	f = &Field{SelBase: SelBase{line: p.line, col: p.col - len(token)}}
	if err == nil {
		b, err = p.skipSpace()
	}
	if b == ':' {
		_, _ = p.readByte() // re-read :
		f.Alias = token
		f.Name, err = p.readToken()
	} else {
		f.Name = token
	}
	if err == nil {
		f.Args, err = p.readArgValues()
	}
	if err == nil {
		f.Dirs, err = p.readDirUses()
	}
	if err == nil {
		f.Sels, err = p.readSelectionSet()
	}
	return
}

func (p *exeParser) readFragment() (sel Selection, err error) {
	var b byte
	for i := 3; 0 < i; i-- {
		if b, err = p.readByte(); err == nil && b != '.' {
			err = parseError(p.line, p.col, "fragments selections must start with '...'")
		}
		if err != nil {
			return
		}
	}
	var token string
	token, err = p.readToken()
	if err == nil {
		switch token {
		case "on":
			var t Type
			line := p.line
			col := p.col
			if t, err = p.readType(); err == nil {
				if _, ok := t.(*Ref); ok {
					err = parseError(line, col, "type %s not defined", t.Name())
				} else {
					sel, err = p.readInline(t)
				}
			}
		case "":
			sel, err = p.readInline(nil)
		default:
			sel, err = p.readFragRef(token)
		}
	}
	return
}

func (p *exeParser) readFragRef(token string) (fr *FragRef, err error) {
	fr = &FragRef{line: p.line, col: p.col}
	if frag := p.exe.Fragments[token]; frag != nil {
		fr.Fragment = frag
	} else {
		fr.Fragment = &Fragment{Name: token, Inline: Inline{SelBase: SelBase{line: p.line, col: p.col - len(token)}}}
		if p.exe.Fragments == nil {
			p.exe.Fragments = map[string]*Fragment{token: fr.Fragment}
		} else {
			p.exe.Fragments[token] = fr.Fragment
		}
	}
	fr.Dirs, err = p.readDirUses()

	return
}

func (p *exeParser) readInline(t Type) (in *Inline, err error) {
	in = &Inline{Condition: t, SelBase: SelBase{line: p.line, col: p.col}}

	if in.Dirs, err = p.readDirUses(); err == nil {
		in.Sels, err = p.readSelectionSet()
	}
	return
}

func (p *exeParser) readFragmentDef() (frag *Fragment, err error) {
	frag = &Fragment{}

	if _, err = p.skipSpace(); err == nil {
		frag.line = p.line
		frag.col = p.col
		frag.Name, err = p.readToken()
	}
	if err == nil {
		_, err = p.skipSpace()
	}
	if err == nil {
		var token string
		if token, err = p.readToken(); token != "on" {
			err = parseError(p.line, p.col-2, "missing fragment condition")
		}
	}
	if err == nil {
		frag.Condition, err = p.readType()
	}
	if err == nil {
		frag.Dirs, err = p.readDirUses()
	}
	if err == nil {
		frag.Sels, err = p.readSelectionSet()
	}
	return
}

func (p *exeParser) readVarDefs() (vds []*VarDef, err error) {
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b == '(' {
		_, _ = p.readByte() // re-read opening (
		// Read the arguments.
		var vd *VarDef
	FOR:
		for {
			if b, err = p.skipSpace(); err != nil {
				return nil, err
			}
			switch b {
			case 0:
				return nil, parseError(p.line, p.col, "variable definitions not terminated with a ')'")
			case ')':
				break FOR
			case '$':
				// expected for any variable definition
			default:
				return nil, parseError(p.line, p.col, "a variable definition must start with a '$'")
			}
			_, _ = p.readByte() // re-read opening $
			if vd, err = p.readVarDef(); err != nil {
				return nil, err
			}
			vds = append(vds, vd)
		}
		_, _ = p.readByte() // re-read closing )
	}
	return
}

func (p *exeParser) readVarDef() (vd *VarDef, err error) {
	vd = &VarDef{}
	if vd.Name, err = p.readToken(); err != nil {
		return
	}
	if len(vd.Name) == 0 {
		return nil, parseError(p.line, p.col, "variable name missing")
	}
	vd.line = p.line
	vd.col = p.col - len(vd.Name)
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b != ':' {
		return nil, parseError(p.line, p.col, "variable name not followed by a :")
	}
	_, _ = p.readByte() // re-read :
	if vd.Type, err = p.readType(); err != nil {
		return nil, err
	}
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b == '=' {
		_, _ = p.readByte() // re-read =
		if vd.Default, err = p.readValue(); err != nil {
			return nil, err
		}
	}
	if err == nil {
		vd.Dirs, err = p.readDirUses()
	}
	return
}
