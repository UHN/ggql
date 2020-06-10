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
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"
	"unicode/utf8"
)

func valueString(v interface{}) string {
	var b bytes.Buffer
	_ = WriteSDLValue(&b, v)
	return b.String()
}

// WriteSDLValue writes a value in SDL format. The option indent determines
// how loose or tight the output looks. The default of 0 is on one line with
// spaces after comma separators. A negative value attempts to make the output
// as tight as possible while also avoiding spaces. Values above zero indent
// by that number of spaces.
func WriteSDLValue(w io.Writer, v interface{}, indents ...int) (err error) {
	indent := 0
	if 0 < len(indents) {
		indent = indents[0]
	}
	return writeValue(w, v, true, 0, indent)
}

// WriteJSONValue writes a value in JSON format. The option indent determines
// how loose or tight the output looks. The default of 0 is on one line with
// spaces after comma separators. A negative value attempts to make the output
// as tight as possible by avoiding spaces. Values above zero indent by that
// number of spaces.
func WriteJSONValue(w io.Writer, v interface{}, indents ...int) (err error) {
	indent := 0
	if 0 < len(indents) {
		indent = indents[0]
	}
	return writeValue(w, v, false, 0, indent)
}

func writeValue(w io.Writer, v interface{}, sdl bool, depth, indent int) (err error) {
	if v == nil {
		_, err = w.Write([]byte(nullStr))
	} else {
		switch tv := v.(type) {
		case string:
			err = writeString(w, tv, true)
		case Symbol:
			err = writeString(w, string(tv), !sdl)
		case Var:
			err = writeString(w, "$"+string(tv), !sdl)
		case int16:
			_, err = w.Write([]byte(strconv.FormatInt(int64(tv), 10)))
		case int32:
			_, err = w.Write([]byte(strconv.FormatInt(int64(tv), 10)))
		case int64:
			_, err = w.Write([]byte(strconv.FormatInt(tv, 10)))
		case int:
			_, err = w.Write([]byte(strconv.FormatInt(int64(tv), 10)))
		case float64:
			_, err = w.Write([]byte(strconv.FormatFloat(tv, 'g', -1, 64)))
		case float32:
			_, err = w.Write([]byte(strconv.FormatFloat(float64(tv), 'g', -1, 32)))
		case bool:
			if tv {
				_, err = w.Write([]byte(trueStr))
			} else {
				_, err = w.Write([]byte(falseStr))
			}
		case map[string]interface{}:
			err = writeMap(w, tv, sdl, depth, indent)
		case []interface{}:
			d2 := depth + 1
			var i2 []byte
			if 0 < indent {
				i2 = append([]byte{'\n'}, bytes.Repeat([]byte{' '}, d2*indent)...)
			}
			_, err = w.Write([]byte{'['})
			noSep := true
			for _, v2 := range tv {
				if !noSep {
					if sep := elementSep(sdl, indent, v2); 0 < len(sep) {
						if err == nil {
							_, err = w.Write(sep)
						}
					}
				}
				if 0 < indent {
					_, err = w.Write(i2)
				}
				if err == nil {
					err = writeValue(w, v2, sdl, d2, indent)
				}
				noSep = indent < 0 && sdl && isCollection(v2)
			}
			if err == nil {
				if 0 < indent {
					_, err = w.Write([]byte{'\n'})
					if err == nil {
						_, err = w.Write(bytes.Repeat([]byte{' '}, depth*indent))
					}
				}
			}
			if err == nil {
				_, err = w.Write([]byte{']'})
			}
			if err == nil && 0 < indent && depth == 0 {
				_, err = w.Write([]byte{'\n'})
			}
		case time.Time:
			_, err = w.Write([]byte{'"'})
			if err == nil {
				_, err = w.Write([]byte(tv.Format(time.RFC3339Nano)))
			}
			if err == nil {
				_, err = w.Write([]byte{'"'})
			}
		default:
			_, err = w.Write([]byte(fmt.Sprintf(`"%v"`, v)))
		}
	}
	return
}

func writeMap(w io.Writer, m map[string]interface{}, sdl bool, depth, indent int) (err error) {
	d2 := depth + 1
	space := 0 <= indent
	var i2 []byte
	if 0 < indent {
		i2 = append([]byte{'\n'}, bytes.Repeat([]byte{' '}, d2*indent)...)
	}
	_, err = w.Write([]byte{'{'})
	noSep := true
	wv := func(key string, vv interface{}) {
		if err == nil && (!sdl || indent <= 0) && !noSep {
			_, err = w.Write([]byte{','})
			if err == nil && indent == 0 {
				_, err = w.Write([]byte{' '})
			}
		}
		if 0 < indent {
			_, err = w.Write(i2)
		}
		if err == nil && !sdl {
			_, err = w.Write([]byte{'"'})
		}
		if err == nil {
			_, err = w.Write([]byte(key))
		}
		if err == nil && !sdl {
			_, err = w.Write([]byte{'"'})
		}
		if err == nil {
			_, err = w.Write([]byte{':'})
		}
		if err == nil && space {
			_, err = w.Write([]byte{' '})
		}
		if err == nil {
			err = writeValue(w, vv, sdl, d2, indent)
		}
		noSep = indent < 0 && sdl && isCollection(vv)
	}
	if Sort {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			wv(k, m[k])
		}
	} else {
		for k, v2 := range m {
			wv(k, v2)
		}
	}
	if err == nil {
		if 0 < indent {
			_, err = w.Write([]byte{'\n'})
			if err == nil {
				_, err = w.Write(bytes.Repeat([]byte{' '}, depth*indent))
			}
		}
	}
	if err == nil {
		_, err = w.Write([]byte{'}'})
	}
	if err == nil && 0 < indent && depth == 0 {
		_, err = w.Write([]byte{'\n'})
	}
	return
}

func isCollection(v interface{}) bool {
	switch v.(type) {
	case []interface{}, map[string]interface{}:
		return true
	default:
		return false
	}
}

func elementSep(sdl bool, indent int, v interface{}) []byte {
	if sdl {
		switch {
		case indent == 0:
			return []byte{',', ' '}
		case 0 < indent:
			return []byte{}
		default: // indent < 0
			if isCollection(v) {
				return []byte{}
			}
			return []byte{','}
		}
	} else { // JSON
		if indent == 0 {
			return []byte{',', ' '}
		}
		return []byte{','}
	}
}

func writeString(w io.Writer, s string, withQuotes bool) (err error) {
	const hexChars = "0123456789abcdef"

	if withQuotes {
		_, err = w.Write([]byte{'"'})
	}
	if err == nil {
		for _, r := range s {
			switch r {
			case '\b':
				_, err = w.Write([]byte{'\\', 'b'})
			case '\f':
				_, err = w.Write([]byte{'\\', 'f'})
			case '\n':
				_, err = w.Write([]byte{'\\', 'n'})
			case '\r':
				_, err = w.Write([]byte{'\\', 'r'})
			case '\t':
				_, err = w.Write([]byte{'\\', 't'})
			case '\\', '"':
				_, err = w.Write([]byte{'\\', byte(r)})
			default:
				if r < 0x80 {
					if r < ' ' {
						// Convert the rune to hex.
						_, err = w.Write([]byte{'\\', 'u', hexChars[r>>12], hexChars[(r>>8)&0x0f], hexChars[(r>>4)&0x0f], hexChars[r&0x0f]})
					} else {
						_, err = w.Write([]byte{byte(r)})
					}
				} else {
					buf := make([]byte, 8)
					n := utf8.EncodeRune(buf, r)
					buf = buf[:n]
					_, err = w.Write(buf)
				}
			}
			if err != nil {
				break
			}
		}
	}
	if err == nil && withQuotes {
		_, err = w.Write([]byte{'"'})
	}
	return
}
