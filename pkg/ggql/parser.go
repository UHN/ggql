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
	"bytes"
	"io"
	"strconv"
	"strings"
)

// There are two GraphQL parsers in this package. Both follow the spec defined
// at https://graphql.github.io/graphql-spec. One parser, the sdlParser parses
// schema SDL while the other, exeParser parses execution requests into a
// Executable. Both make use of the basic parser type in this package. The
// parsers are single pass parsers. The parser type includes lower level
// functions such as readByte, skipSpace, and readToken. The sdlParser and
// exeParser implement higher level functions that parse into the types
// associated with either the schema or execution request documents.
//
// The parser includes public functions for parsing values where a value is
// composed of core types plus []interface{} and map[string]interface{}. The
// parsing is liberal in that it will handle both JSON and SDL formats.
//
// The sdlParser and exeParser are initiated with the parseSDL and parseExe
// functions which loop until the end of the input is reached. Each element
// type has it's own function.
//

// Character maps are used to classify characters after they are read. This is
// an efficient way of checking whether a character belongs in a token or is a
// space character. The charMap and numMap strings are the classification maps
// laid out with labels to help locate the relevant classification for each
// ASCII value.

const (
	spaceChar = 'w' // white space in the char map, note a comma is considered white space
	tokenChar = 't' // token in the char map, alphanumeric and underscore
	numChar   = 'n' // number in the char map including ., +, - e, and E
	// puncChar  = 'p' // punctuation in the char map (not used in the code yet)

	//   0123456789abcdef0123456789abcdef
	charMap = "" +
		".........ww..w.................." + // 0x00
		"wp......pp..w.p.ttttttttttp..p.." + // 0x20
		"pttttttttttttttttttttttttttp.p.t" + // 0x40
		".ttttttttttttttttttttttttttppp.." + // 0x60
		"................................" + // 0x80
		"................................" + // 0xa0
		"................................" + // 0xc0
		"................................" // 0xe0

	//   0123456789abcdef0123456789abcdef
	numMap = "" +
		"................................" + // 0x00
		"...........n.nn.nnnnnnnnnn......" + // 0x20
		".....n.........................." + // 0x40
		".....n.........................." + // 0x60
		"................................" + // 0x80
		"................................" + // 0xa0
		"................................" + // 0xc0
		"................................" // 0xe0
)

type parser struct {
	root   *Root
	reader io.Reader
	line   int
	col    int
	onDeck byte
	eof    bool
}

// ParseValue parses a reader into a value where the input follows the SDL
// syntax. Basically a very relaxed JSON syntax. The parse does not validate
// against any specific type but generates a generic value on success.
func ParseValue(reader io.Reader) (interface{}, error) {
	return (&parser{reader: reader}).readValue()
}

// ParseValueString parses a string into a value where the input follows the
// SDL syntax. Basically a very relaxed JSON syntax. The parse does not
// validate against any specific type but generates a generic value on
// success.
func ParseValueString(s string) (interface{}, error) {
	return (&parser{reader: strings.NewReader(s)}).readValue()
}

func (p *parser) readByte() (b byte, err error) {
	if p.onDeck != 0 {
		b = p.onDeck
		p.onDeck = 0
		return
	}
	if p.eof {
		return 0, nil
	}
	ba := []byte{'0'}
	var n int
	if p.line == 0 {
		p.line = 1
		p.col = 1
	}
	// Break out of loop on error or success. If a 0, err is returned then try
	// again. Should not happen but go docs indicate it is possible. The
	// breaks instead of a return allow complete test coverage even if the
	// rare condition can not be replicated in unit tests.
	for {
		if n, err = p.reader.Read(ba); err != nil {
			break
		}
		if n != 0 {
			b = ba[0]
			if b == '\n' {
				p.line++
				p.col = 0
			}
			p.col++
			break
		}
	}
	if err == io.EOF {
		p.eof = true
		err = nil
		if 0 < n {
			// This occurs with HTTP POST requests. EOF with a character read.
			b = ba[0]
			p.col++
		}
	}
	return
}

func (p *parser) putBack(b byte) {
	p.onDeck = b
}

func (p *parser) skipBOM() (err error) {
	var b byte
	// Check for a UTF-8 BOM. Any other BOM will fail later.
	if b, err = p.readByte(); err == nil && b != 0xEF {
		// No BOM
		p.putBack(b)
		return
	}
	for _, bom := range []byte{0xBB, 0xBF} {
		if err == nil {
			if b, err = p.readByte(); err == nil && b != bom {
				err = parseError(p.line, p.col, "invalid BOM")
			}
		}
	}
	return
}

// Skips white space as well as code comments that start with #.
func (p *parser) skipSpace() (b byte, err error) {
	for {
		if b, err = p.readByte(); err != nil || b == 0 {
			return
		}
		if charMap[b] == spaceChar {
			continue
		}
		if b == '#' {
			// Read to end of line.
			for {
				if b, err = p.readByte(); err != nil || b == 0 {
					return
				}
				if b == '\n' {
					break
				}
			}
		} else {
			p.putBack(b)
			break
		}
	}
	return
}

func (p *parser) readToken() (string, error) {
	b, err := p.skipSpace()
	if err != nil || b == 0 {
		return "", err
	}
	var buf bytes.Buffer
	for {
		if b, err = p.readByte(); err != nil || b == 0 {
			return buf.String(), err
		}
		if charMap[b] == tokenChar {
			_ = buf.WriteByte(b)
		} else {
			p.putBack(b)
			return buf.String(), nil
		}
	}
}

func (p *parser) readNumberToken() (string, error) {
	var buf bytes.Buffer
	for {
		b, err := p.readByte()
		if err != nil || b == 0 {
			return buf.String(), err
		}
		if numMap[b] == numChar {
			_ = buf.WriteByte(b)
		} else {
			p.putBack(b)
			return buf.String(), nil
		}
	}
}

func (p *parser) readType() (t Type, err error) {
	var b byte
	b, err = p.skipSpace()
	if err == nil {
		switch b {
		case 0:
			return
		case '[':
			_, _ = p.readByte() // re-read [
			if t, err = p.readType(); err != nil {
				return
			}
			b, err = p.skipSpace()
			switch {
			case err != nil:
				t = nil
			case b == ']':
				_, _ = p.readByte() // re-read ]
				t = &List{Base: t, line: p.line, col: p.col}
			default:
				err = parseError(p.line, p.col, "list not terminated with a ']'")
			}
		default:
			var token string
			token, err = p.readToken()
			if err == nil && 0 < len(token) {
				if t = p.root.GetType(token); t == nil {
					t = &Ref{Base: Base{N: token}}
				}
			}
		}
	}
	if err == nil && t != nil {
		b, err = p.skipSpace()
		if err == nil && b == '!' {
			t = &NonNull{Base: t, line: p.line, col: p.col}
			_, _ = p.readByte() // re-read !
		}
	}
	return
}

func (p *parser) readDesc() (desc string, err error) {
	if desc, err = p.readString(); err == nil {
		// Documentation strings with multiple lines should have the
		// indentation removed.
		var lines []string
		for _, line := range strings.Split(desc, "\n") {
			line = strings.TrimSpace(line)
			if 0 < len(line) {
				lines = append(lines, line)
			}
		}
		desc = strings.Join(lines, "\n")
	}
	return
}

func (p *parser) readString() (string, error) {
	var buf bytes.Buffer

	// Read the next byte. It should be a ". If not then there is no string to
	// read.
	b, err := p.readByte()
	if err != nil || b == 0 {
		return "", err
	}
	if b != '"' {
		p.putBack(b)
		return "", nil
	}
	noTerm := func() (string, error) {
		return "", parseError(p.line, p.col, "string not terminated")
	}
	// Determine if it is a simple string terminated by single quotes or a
	// multiline string surrounded by triple quotes.
	if b, err = p.readByte(); err != nil {
		return "", err
	}
	switch b {
	case '"':
		if b, err = p.readByte(); err != nil {
			return "", err
		}
		if b != '"' { // assume an empty string for now
			p.putBack(b)
			return "", nil
		}
		for {
			if b, err = p.readByte(); err != nil {
				return "", err
			}
			switch b {
			case '"':
				if b, err = p.readByte(); err != nil {
					return "", err
				}
				if b == '"' {
					if b, err = p.readByte(); err != nil {
						return "", err
					}
					if b == '"' {
						return buf.String(), nil
					}
					buf.WriteByte('"')
					buf.WriteByte('"')
					buf.WriteByte(b)
				} else {
					buf.WriteByte('"')
					buf.WriteByte(b)
				}
			case '\\':
				var r rune
				if r, err = p.readEscaped(); err != nil {
					return "", err
				}
				buf.WriteRune(r)
			case 0:
				return noTerm()
			default:
				buf.WriteByte(b)
			}
		}
	case 0:
		return noTerm()
	default:
		p.putBack(b)
		for {
			if b, err = p.readByte(); err != nil {
				return "", err
			}
			switch b {
			case '"':
				return buf.String(), nil
			case '\\':
				var r rune
				if r, err = p.readEscaped(); err != nil {
					return "", err
				}
				buf.WriteRune(r)
			case 0:
				return noTerm()
			default:
				buf.WriteByte(b)
			}
		}
	}
}

func (p *parser) readEscaped() (rune, error) {
	badEscape := func() (rune, error) {
		return 0, parseError(p.line, p.col, "invalid escaped unicode character")
	}
	b, err := p.readByte()
	if err != nil {
		return 0, err
	}
	switch b {
	case 0:
		return badEscape()
	case '"':
		return rune('"'), nil
	case '\\':
		return rune('\\'), nil
	case '/':
		return rune('/'), nil
	case 'b':
		return rune('\b'), nil
	case 'f':
		return rune('\f'), nil
	case 'n':
		return rune('\n'), nil
	case 'r':
		return rune('\r'), nil
	case 't':
		return rune('\t'), nil
	case 'u':
		var u uint32
		for i := 0; i < 4; i++ {
			if b, err = p.readByte(); err != nil {
				return 0, err
			}
			switch b {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				u = (u << 4) + uint32(b-'0')
			case 'a', 'b', 'c', 'd', 'e', 'f':
				u = (u << 4) + uint32(b-'a'+10)
			case 'A', 'B', 'C', 'D', 'E', 'F':
				u = (u << 4) + uint32(b-'A'+10)
			default:
				return badEscape()
			}
		}
		return rune(u), nil
	default:
		return badEscape()
	}
}

func (p *parser) readValue() (v interface{}, err error) {
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return
	}
	var token string
	switch b {
	case 0:
		return nil, nil
	case '"':
		if token, err = p.readString(); err != nil {
			return
		}
		v = token
	case '$':
		_, _ = p.readByte() // re-read $
		if token, err = p.readToken(); err != nil {
			return
		}
		v = Var(token)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if token, err = p.readNumberToken(); err != nil {
			return
		}
		switch p.onDeck {
		case 0, ' ', '\t', '\n', '\r', '\f', ',', '}', ']', '{', '[', ')':
			// okay
		default:
			return nil, parseError(p.line, p.col, "value '%s' followed by a non-numeric character %s",
				token, []byte{p.onDeck})
		}
		if i, e := strconv.ParseInt(token, 10, 64); e == nil {
			v = i
		} else if f, e2 := strconv.ParseFloat(token, 64); e2 == nil {
			v = f
		} else {
			return nil, parseError(p.line, p.col, "value '%s' not a number", token)
		}
	case '[':
		_, _ = p.readByte() // re-read [
		list := []interface{}{}
		for {
			if b, err = p.skipSpace(); err != nil {
				return
			}
			switch b {
			case 0:
				return nil, parseError(p.line, p.col, "list value not terminated")
			case ']':
				_, _ = p.readByte() // re-read ]
				v = list
				return
			}
			if v, err = p.readValue(); err != nil {
				return
			}
			list = append(list, v)
		}
	case '{':
		_, _ = p.readByte() // re-read {
		obj := map[string]interface{}{}
		for {
			if b, err = p.skipSpace(); err != nil {
				return
			}
			switch b {
			case 0:
				return nil, parseError(p.line, p.col, "object not terminated")
			case '}':
				_, _ = p.readByte() // re-read }
				v = obj
				return
			}
			// The keys can be either a token or a string. (with or without quotes)
			if b == '"' {
				if token, err = p.readString(); err != nil {
					return
				}
			} else if token, err = p.readToken(); err != nil {
				return
			}
			if b, err = p.skipSpace(); err != nil {
				return
			}
			if b != ':' {
				return nil, parseError(p.line, p.col, "object key must be followed by a ':'")
			}
			_, _ = p.readByte() // re-read :
			if v, err = p.readValue(); err != nil {
				return
			}
			obj[token] = v
		}
	default:
		line := p.line
		col := p.col
		onDeck := p.onDeck
		if token, err = p.readToken(); err != nil {
			return
		}
		if line == p.line && col == p.col && onDeck == p.onDeck {
			return nil, parseError(p.line, p.col, "invalid value")
		}
		switch token {
		case trueStr:
			v = true
		case falseStr:
			v = false
		case nullStr, "":
			v = nil
		default:
			v = Symbol(token)
		}
	}
	return
}

func (p *parser) readDirUses() (dus []*DirectiveUse, err error) {
	var du *DirectiveUse
	for {
		if du, err = p.readDirUse(); du == nil || err != nil {
			break
		}
		dus = append(dus, du)
	}
	return
}

// https://graphql.github.io/graphql-spec/June2018/#Directive
func (p *parser) readDirUse() (du *DirectiveUse, err error) {
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b != '@' {
		return nil, nil
	}
	// Skip @. No need to check errors since the next byte was already read.
	_, _ = p.readByte()
	du = &DirectiveUse{line: p.line, col: p.col - 1}
	if du.Directive, err = p.readType(); err != nil {
		return nil, err
	}
	if du.Directive == nil {
		return nil, parseError(p.line, p.col, "directive missing")
	}
	if p.onDeck == '(' {
		_, _ = p.readByte() // re-read opening (
		// Read the arguments.
		var av *ArgValue
		du.Args = map[string]*ArgValue{}
		for {
			if b, err = p.skipSpace(); err != nil {
				return nil, err
			}
			if b == ')' {
				break
			}
			if b == 0 {
				return nil, parseError(p.line, p.col, "arguments not terminated with a ')'")
			}
			if av, err = p.readArgValue(); err != nil {
				return nil, err
			}
			du.Args[av.Arg] = av
		}
		_, _ = p.readByte() // re-read closing )
	}
	// Default arg values.
	if dir, _ := du.Directive.(*Directive); dir != nil && 0 < dir.args.Len() {
		if du.Args == nil {
			du.Args = map[string]*ArgValue{}
		}
		for _, a := range dir.args.list {
			if av := du.Args[a.N]; av == nil {
				du.Args[a.N] = &ArgValue{Arg: a.N, Value: a.Default, line: p.line, col: p.col - len(a.N) + 1}
			}
		}
	}
	return
}

func (p *parser) readArgValues() (avs []*ArgValue, err error) {
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b == '(' {
		_, _ = p.readByte() // re-read opening (
		// Read the arguments.
		var av *ArgValue
		for {
			if b, err = p.skipSpace(); err != nil {
				return nil, err
			}
			if b == ')' {
				break
			}
			if b == 0 {
				return nil, parseError(p.line, p.col, "arguments not terminated with a ')'")
			}
			if av, err = p.readArgValue(); err != nil {
				return nil, err
			}
			avs = append(avs, av)
		}
		_, _ = p.readByte() // re-read closing )
	}
	return
}

func (p *parser) readArgValue() (av *ArgValue, err error) {
	av = &ArgValue{}
	if av.Arg, err = p.readToken(); err != nil {
		return
	}
	av.line = p.line
	av.col = p.col - len(av.Arg) - 1
	if len(av.Arg) == 0 {
		return nil, parseError(p.line, p.col, "argument name missing")
	}
	var b byte
	if b, err = p.skipSpace(); err != nil {
		return nil, err
	}
	if b != ':' {
		return nil, parseError(p.line, p.col, "argument name not followed by a :")
	}
	_, _ = p.readByte() // re-read :
	av.Value, err = p.readValue()

	return
}
