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

// Location represents the GraphQL Location enum.
type Location string

// It would be more friendly to have constant names that matched the GraphQL
// enum values but golint complains about all caps and underscores. Basically
// naming conflict between languages or at least the people who defined the
// linter style rules. Instead each location constant name is preficed with
// 'Loc' and then camel case is used.
const (
	// LocSchema is the SCHEMA location.
	LocSchema = Location("SCHEMA")
	// LocScalar is the SCALAR location.
	LocScalar = Location("SCALAR")
	// LocObject is the OBJECT location.
	LocObject = Location("OBJECT")
	// LocFieldDefinition is the FIELDDEFINITION location.
	LocFieldDefinition = Location("FIELD_DEFINITION")
	// LocArgumentDefinition is the ARGUMENT_DEFINITION location.
	LocArgumentDefinition = Location("ARGUMENT_DEFINITION")
	// LocInterface is the INTERFACE location.
	LocInterface = Location("INTERFACE")
	// LocUnion is the UNION location.
	LocUnion = Location("UNION")
	// LocEnum is the ENUM location.
	LocEnum = Location("ENUM")
	// LocEnumValue is the ENUM_VALUE location.
	LocEnumValue = Location("ENUM_VALUE")
	// LocInputObject is the INPUT_OBJECT location.
	LocInputObject = Location("INPUT_OBJECT")
	// LocInputFieldDefinition is the INPUT_FIELD_DEFINITION location.
	LocInputFieldDefinition = Location("INPUT_FIELD_DEFINITION")

	// LocQuery is the QUERY location.
	LocQuery = Location("QUERY")
	// LocMutation is the MUTATION location.
	LocMutation = Location("MUTATION")
	// LocSubscription is the SUBSCRIPTION location.
	LocSubscription = Location("SUBSCRIPTION")
	// LocField is the FIELD location.
	LocField = Location("FIELD")
	// LocFragmentDefinition is the FRAGMENT_DEFINITION location.
	LocFragmentDefinition = Location("FRAGMENT_DEFINITION")
	// LocFragmentSpread is the FRAGMENT_SPREAD location.
	LocFragmentSpread = Location("FRAGMENT_SPREAD")
	// LocInlineFragment is the INLINE_FRAGMENT location.
	LocInlineFragment = Location("INLINE_FRAGMENT")
	// LocVariableDefinition is the VARIABLE_DEFINITION location.
	LocVariableDefinition = Location("VARIABLE_DEFINITION")
)

// IsLocation returns true if loc is a valid location.
func IsLocation(loc Location) bool {
	for _, l := range []Location{
		LocSchema,
		LocScalar,
		LocObject,
		LocFieldDefinition,
		LocArgumentDefinition,
		LocInterface,
		LocUnion,
		LocEnum,
		LocEnumValue,
		LocInputObject,
		LocInputFieldDefinition,
		LocQuery,
		LocMutation,
		LocSubscription,
		LocField,
		LocFragmentDefinition,
		LocFragmentSpread,
		LocInlineFragment,
		LocVariableDefinition,
	} {
		if l == loc {
			return true
		}
	}
	return false
}

// Locate the location for the type. An empty string is returned if it is not
// a locatable type.
func Locate(t interface{}) Location {
	switch tt := t.(type) {
	case *Object, *uuSchema:
		return LocObject
	case *Schema:
		return LocSchema
	case *FieldDef:
		return LocFieldDefinition
	case *ArgValue:
		return LocArgumentDefinition
	case *InputField, *Arg:
		return LocInputFieldDefinition
	case *Interface:
		return LocInterface
	case *Union:
		return LocUnion
	case *Enum:
		return LocEnum
	case *EnumValue:
		return LocEnumValue
	case *Input:
		return LocInputObject
	case *Scalar, *booleanScalar, *intScalar, *stringScalar, *floatScalar, *idScalar, *int64Scalar, *timeScalar:
		return LocScalar

	case *Field:
		return LocField
	case *Op:
		switch tt.Type {
		case OpQuery:
			return LocQuery
		case OpMutation:
			return LocMutation
		case OpSubscription:
			return LocSubscription
		}
	case *Fragment:
		return LocFragmentDefinition
	case *Inline:
		return LocInlineFragment
	case *FragRef:
		return LocFragmentSpread
	case *VarDef:
		return LocVariableDefinition
	}
	return ""
}

// IsInputType returns true if t is an input type.
func IsInputType(t interface{}) bool {
	switch tt := t.(type) {
	case *Input, *Enum:
		return true
	case *List:
		return IsInputType(tt.Base)
	case *NonNull:
		return IsInputType(tt.Base)
	default:
		if _, ok := t.(InCoercer); ok {
			return true
		}
	}
	return false
}

// IsOutputType returns true if t is an output type.
func IsOutputType(t interface{}) bool {
	switch tt := t.(type) {
	case *Object, *Interface, *Union, *Enum, *uuSchema:
		return true
	case *List:
		return IsOutputType(tt.Base)
	case *NonNull:
		return IsOutputType(tt.Base)
	default:
		if _, ok := t.(OutCoercer); ok {
			return true
		}
	}
	return false
}
