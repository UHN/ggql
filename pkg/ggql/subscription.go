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

// Subscription encapsulates the information about a subscription.
type Subscription struct {
	sub   Subscriber
	field *Field
	args  map[string]interface{}
}

// NewSubscription creates a new subscription. It should be called in a
// Resolver.Resolve() call.
func NewSubscription(sub Subscriber, field *Field, args map[string]interface{}) *Subscription {
	return &Subscription{
		sub:   sub,
		field: field,
		args:  args,
	}
}

func (sub *Subscription) prep(root *Root) {
	sub.field.ConType = root.getFieldType(sub.field.ConType, sub.field.Name)
}
