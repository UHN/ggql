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

// Subscriber is the interface for subscription implementations.
type Subscriber interface {

	// Send an event on channel associated with the subscription.
	Send(value interface{}) error

	// Match an eventID to the subscriber. Behavior is up to the
	// subscriber. An exact match could be used or a system that uses
	// wildcards can be used. A return of true indicates a match.
	Match(eventID string) bool

	// Unsubscribe should clean up any resources associated with a
	// subscription.
	Unsubscribe()
}
