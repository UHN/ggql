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
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestArgValueNil(t *testing.T) {
	av := ggql.ArgValue{Arg: "a", Value: nil}
	var b bytes.Buffer

	_ = av.Write(&b)
	checkEqual(t, "a: null", b.String(), "Arg null write")
}

func TestArgValueInt(t *testing.T) {
	av := ggql.ArgValue{Arg: "a"}
	for _, v := range []interface{}{
		int16(3),
		int32(3),
		int64(3),
		int(3),
	} {
		var b bytes.Buffer

		av.Value = v
		_ = av.Write(&b)
		checkEqual(t, "a: 3", b.String(), "Arg %T write", v)
	}
}

func TestArgValueFloat(t *testing.T) {
	av := ggql.ArgValue{Arg: "a"}
	for _, v := range []interface{}{
		float32(3.1),
		float64(3.1),
	} {
		var b bytes.Buffer

		av.Value = v
		_ = av.Write(&b)
		checkEqual(t, "a: 3.1", b.String(), "Arg %T write", v)
	}
}

func TestArgValueBoolean(t *testing.T) {
	av := ggql.ArgValue{Arg: "a", Value: true}
	var b bytes.Buffer

	_ = av.Write(&b)
	checkEqual(t, "a: true", b.String(), "Arg Boolean write")

	av.Value = false
	b.Reset()
	_ = av.Write(&b)
	checkEqual(t, "a: false", b.String(), "Arg Boolean write")
}

func TestArgValueString(t *testing.T) {
	av := ggql.ArgValue{Arg: "a", Value: "str"}
	var b bytes.Buffer

	_ = av.Write(&b)
	checkEqual(t, `a: "str"`, b.String(), "Arg String write")
}

func TestArgValueList(t *testing.T) {
	av := ggql.ArgValue{Arg: "a", Value: []interface{}{"one", "two"}}
	var b bytes.Buffer

	_ = av.Write(&b)
	checkEqual(t, `a: ["one", "two"]`, b.String(), "Arg List write")
}

func TestArgValueObject(t *testing.T) {
	av := ggql.ArgValue{Arg: "a", Value: map[string]interface{}{"one": "1"}}
	var b bytes.Buffer

	_ = av.Write(&b)
	checkEqual(t, `a: {one: "1"}`, b.String(), "Arg InputObject write")

	av.Value = map[string]interface{}{"one": "1", "two": "2"}
	ggql.Sort = true
	b.Reset()
	_ = av.Write(&b)
	checkEqual(t, `a: {one: "1", two: "2"}`, b.String(), "Arg InputObject write")
}

func TestArgValueOther(t *testing.T) {
	av := ggql.ArgValue{Arg: "a", Value: int8(3)}
	var b bytes.Buffer

	_ = av.Write(&b)
	checkEqual(t, `a: "3"`, b.String(), "Arg InputObject write")
}
