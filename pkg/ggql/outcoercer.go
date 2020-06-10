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

// OutCoercer interface for scalars defines the function that is used to coerce
// a scalar value into an output type that can be encoded as JSON.
type OutCoercer interface {

	// CoerceOut coerces a result value into a value suitable for output.
	CoerceOut(v interface{}) (interface{}, error)
}
