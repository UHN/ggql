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
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrDuplicate indicates a type is already in the schema.
	ErrDuplicate = errors.New("duplicate")

	// ErrCoerce indicates a coercion failed.
	ErrCoerce = errors.New("can not coerce")

	// ErrValidation indicates a validation error.
	ErrValidation = errors.New("validation")

	// ErrParse indicates a parse error.
	ErrParse = errors.New("parse error")

	// ErrNotFound indicates an item was not found.
	ErrNotFound = errors.New("not found")

	// ErrTypeMismatch indicates there is a type mismatch.
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrResolve indicates an error during resolving occurred.
	ErrResolve = errors.New("resolve error")

	// ErrMeta indicates an error with a type or field registration.
	ErrMeta = errors.New("reflection error")
)

func newCoerceErr(val interface{}, typeName string) error {
	if IsNil(val) {
		return fmt.Errorf("%w null into a %s", ErrCoerce, typeName)
	}
	return fmt.Errorf("%w a %T into a %s", ErrCoerce, val, typeName)
}

// Errors is a slice of errors.
type Errors []error

// Error returns the instance as a string.
func (err Errors) Error() string {
	var b bytes.Buffer
	b.WriteString("Errors{\n")
	for _, e := range err {
		b.WriteString(fmt.Sprintf("  %s\n", e))
	}
	b.WriteString("}\n")

	return b.String()
}

func (err Errors) in(loc interface{}) {
	for _, e := range err {
		if ee, _ := e.(*Error); ee != nil {
			ee.in(loc)
		}
	}
}

// Error encapsulates the data for a GraphQL error. It is used to populate the
// "errors" element of a response if there are errors.
type Error struct {

	// Base error. Similar to using a %w in a fmt.Errorf call.
	Base error

	// Line on which the error occurs in the GraphQL portion of the
	// request. A zero value indicates the determination of the line could not
	// be made.
	Line int

	// Column on the line in the GraphQL portion of the request that the error
	// is associated with. A zero value indicates the determination of the
	// column could not be made.
	Column int

	// Path to the element in the response that the error applies to. If empty
	// or nil then it does not apply to the response data or that
	// determination could not be made.
	Path []interface{}
}

// Error returns a string representation of the error.
func (err *Error) Error() string {
	var b strings.Builder

	b.WriteString(err.Base.Error())
	if 0 < len(err.Path) {
		b.WriteString(" at ")
		for i, p := range err.Path {
			if 0 < i {
				b.WriteByte('.')
			}
			b.WriteString(fmt.Sprintf("%v", p))
		}
	}
	if 0 < err.Line || 0 < err.Column {
		b.WriteString(fmt.Sprintf(" from %d:%d", err.Line, err.Column))
	}
	return b.String()
}

func (err *Error) in(loc interface{}) {
	err.Path = append([]interface{}{loc}, err.Path...)
}

// parseError creates a parse Error with a line and column with the Abort flag
// set to true.
func parseError(line, col int, format string, args ...interface{}) error {
	return &Error{
		Base:   fmt.Errorf("%w: "+format, append([]interface{}{ErrParse}, args...)...),
		Line:   line,
		Column: col,
	}
}

// valError creates a validation Error with a line and column with the Abort
// flag set to true.
func valError(line, col int, format string, args ...interface{}) error {
	return &Error{
		Base:   fmt.Errorf("%w: "+format, append([]interface{}{ErrValidation}, args...)...),
		Line:   line,
		Column: col,
	}
}

func resError(line, col int, format string, args ...interface{}) error {
	return &Error{
		Base:   fmt.Errorf("%w: "+format, append([]interface{}{ErrResolve}, args...)...),
		Line:   line,
		Column: col,
	}
}

func resWarn(line, col int, format string, args ...interface{}) error {
	return &Error{
		Base:   fmt.Errorf("%w: "+format, append([]interface{}{ErrResolve}, args...)...),
		Line:   line,
		Column: col,
	}
}

func resWarnp(sel Selection, format string, args ...interface{}) error {
	pa := []interface{}{}
	if f, _ := sel.(*Field); f != nil {
		pa = append(pa, f.key())
	}
	err := Error{
		Base: fmt.Errorf("%w: "+format, append([]interface{}{ErrResolve}, args...)...),
		Path: pa,
	}
	if sel != nil {
		err.Line = sel.Line()
		err.Column = sel.Column()
	}
	return &err
}
