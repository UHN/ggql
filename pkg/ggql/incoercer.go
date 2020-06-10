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

// InCoercer interface for scalars defines the function that is used to coerce
// an input value into the scalar type.
type InCoercer interface {

	// CoerceIn coerces an input value into the expected input type if possible
	// otherwise an error is returned.
	CoerceIn(v interface{}) (interface{}, error)
}
