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
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestLocate(t *testing.T) {
	for loc, v := range map[ggql.Location]interface{}{
		ggql.LocObject:               &ggql.Object{},
		ggql.LocSchema:               &ggql.Schema{},
		ggql.LocFieldDefinition:      &ggql.FieldDef{},
		ggql.LocArgumentDefinition:   &ggql.ArgValue{},
		ggql.LocInputFieldDefinition: &ggql.Arg{},
		ggql.LocInterface:            &ggql.Interface{},
		ggql.LocUnion:                &ggql.Union{},
		ggql.LocEnum:                 &ggql.Enum{},
		ggql.LocEnumValue:            &ggql.EnumValue{},
		ggql.LocInputObject:          &ggql.Input{},
		ggql.LocScalar:               &ggql.Scalar{},
		"":                           "not locatable",
	} {
		checkEqual(t, string(loc), string(ggql.Locate(v)),
			"locate %T should return %s not '%s'", v, loc, ggql.Locate(v))
	}
}

func TestIsInputType(t *testing.T) {
	checkEqual(t, true, ggql.IsInputType(&ggql.Input{}), "Input should be an input type")
	checkEqual(t, true, ggql.IsInputType(&ggql.List{Base: &ggql.Input{}}), "[Input] should be an input type")
	checkEqual(t, true, ggql.IsInputType(&ggql.NonNull{Base: &ggql.Input{}}), "Input! should be an input type")
	checkEqual(t, true, ggql.IsInputType(&ggql.NonNull{Base: &ggql.List{Base: &ggql.NonNull{Base: &ggql.Input{}}}}), "[Input!]! should be an input type")
	checkEqual(t, false, ggql.IsInputType(&ggql.Object{}), "Object should not be an input type")
}

func TestIsoutputType(t *testing.T) {
	checkEqual(t, true, ggql.IsOutputType(&ggql.Object{}), "Output should be an output type")
	checkEqual(t, true, ggql.IsOutputType(&ggql.List{Base: &ggql.Object{}}), "[Output] should be an output type")
	checkEqual(t, true, ggql.IsOutputType(&ggql.NonNull{Base: &ggql.Object{}}), "Output! should be an output type")
	checkEqual(t, true, ggql.IsOutputType(&ggql.NonNull{Base: &ggql.List{Base: &ggql.NonNull{Base: &ggql.Object{}}}}), "[Output!]! should be an output type")
	checkEqual(t, false, ggql.IsOutputType(&ggql.Input{}), "Input should not be an output type")
}
