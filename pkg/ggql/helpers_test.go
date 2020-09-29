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
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

const (
	red           = "\x1b[31m"
	green         = "\x1b[32m"
	normal        = "\x1b[m"
	timeScalarSDL = `
"""
Time can be coerced to and from a string that is formatted according to the
RFC3339 specification with nanoseconds.
"""
scalar Time
`
)

type call struct {
	fun  string
	file string
	line int
}

func (c call) String() string {
	return fmt.Sprintf("%s @ %s:%d", c.fun, c.file, c.line)
}

func checkNil(t *testing.T, val interface{}, format string, args ...interface{}) {
	if !ggql.IsNil(val) {
		if !reflect.ValueOf(val).IsNil() {
			t.Fatalf("\nexpect: nil, actual: %s\n%s"+format,
				append([]interface{}{val, stack()}, args...)...)
		}
	}
}

func checkNotNil(t *testing.T, val interface{}, format string, args ...interface{}) {
	if ggql.IsNil(val) {
		t.Fatalf("\nexpect: not nil, actual: nil\n%s"+format, append([]interface{}{stack()}, args...)...)
	}
}

func checkEqual(t *testing.T, expect, actual interface{}, format string, args ...interface{}) {
	eq := false
	switch ta := actual.(type) {
	case bool:
		if tx, ok := expect.(bool); ok {
			eq = tx == ta
		}

	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64:
		eq = asInt(t, actual, format, args...) == asInt(t, expect, format, args...)
	case float32, float64:
		eq = asFloat(t, actual, format, args...) == asFloat(t, expect, format, args...)
	case ggql.Symbol:
		if tx, ok := expect.(ggql.Symbol); ok {
			eq = tx == ta
		}
	case string:
		if tx, ok := expect.(string); ok {
			eq = tx == ta
			if !eq {
				tx, ta = colorizeStrings(tx, ta)
				expect = tx
				actual = ta
			}
		}
	}
	if !eq {
		t.Fatalf("\nexpect: %v\nactual: %v\n%s"+format, append([]interface{}{expect, actual, stack()}, args...)...)
	}
}

func asInt(t *testing.T, v interface{}, format string, args ...interface{}) (i int64) {
	switch tv := v.(type) {
	case int:
		i = int64(tv)
	case int8:
		i = int64(tv)
	case int16:
		i = int64(tv)
	case int32:
		i = int64(tv)
	case int64:
		i = tv
	case uint:
		i = int64(tv)
	case uint8:
		i = int64(tv)
	case uint16:
		i = int64(tv)
	case uint32:
		i = int64(tv)
	case uint64:
		i = int64(tv)
	default:
		t.Fatalf("expected int not %T "+format, append([]interface{}{v, stack()}, args...)...)
	}
	return
}

func asFloat(t *testing.T, v interface{}, format string, args ...interface{}) (f float64) {
	switch tv := v.(type) {
	case float32:
		f = float64(tv)
	case float64:
		f = tv
	default:
		t.Fatalf("expected float not %T "+format, append([]interface{}{v, stack()}, args...)...)
	}
	return
}

func stack() string {
	var b bytes.Buffer

	pc := make([]uintptr, 40)
	cnt := runtime.Callers(2, pc) - 2
	stack := make([]call, cnt)

	var fun *runtime.Func
	var c *call

	for i := 0; i < cnt; i++ {
		c = &stack[i]
		fun = runtime.FuncForPC(pc[i])
		c.file, c.line = fun.FileLine(pc[i])
		c.fun = fun.Name()
		b.WriteString(c.String())
		b.WriteByte('\n')
	}
	return b.String()
}

func colorizeStrings(expect, actual string) (string, string) {
	max := len(expect)
	if len(actual) < max {
		max = len(actual)
	}

	for i := 0; i < max; i++ {
		if expect[i] != actual[i] {
			expect = expect[:i] + green + expect[i:] + normal
			actual = actual[:i] + red + actual[i:] + normal
			return expect, actual
		}
	}
	if max < len(actual) {
		actual = actual[:max] + red + actual[max:] + normal
	} else {
		expect = expect[:max] + green + expect[max:] + normal
	}
	return expect, actual
}

type failWriter struct {
	max int
}

func (w *failWriter) Write(p []byte) (n int, err error) {
	w.max -= len(p)
	if w.max < 0 {
		return 0, fmt.Errorf("forced a fail")
	}
	return len(p), nil
}

type failReader struct {
	max     int
	pos     int
	content []byte
}

func (r *failReader) Read(p []byte) (n int, err error) {
	start := r.pos
	r.pos += len(p)
	if r.max < r.pos {
		return 0, fmt.Errorf("forced a fail")
	}
	return copy(p, r.content[start:]), nil
}
