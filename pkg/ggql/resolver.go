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

// Resolver is the interface for resolving fields on an object.
type Resolver interface {

	// Resolve a field on an object. The field argument is the name of the
	// field to resolve. The args parameter includes the values associated
	// with the arguments provided by the caller. The function should return
	// the field's object or an error. A return of nil is also possible.
	Resolve(field *Field, args map[string]interface{}) (interface{}, error)
}
