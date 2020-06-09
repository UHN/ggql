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
	"fmt"
	"strings"
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

// Sub a Subscriber test implementation.
type Sub struct {
	id  string
	log *strings.Builder
}

func newSub(id string, log *strings.Builder) *Sub {
	sub := Sub{id: id, log: log}

	return &sub
}

func (sub *Sub) Send(value interface{}) error {
	if sub.log == nil {
		return fmt.Errorf("fail here")
	}
	ggql.WriteJSONValue(sub.log, value, 2)

	return nil
}

func (sub *Sub) Match(eventID string) bool {
	return sub.id == eventID || len(sub.id) == 0
}

func (sub *Sub) Unsubscribe() {
}

func (sub *Subscription) Resolve(field *ggql.Field, args map[string]interface{}) (interface{}, error) {
	switch field.Name {
	case "like":
		topic, _ := args["topic"].(string)

		return ggql.NewSubscription(newSub(topic, sub.log), field, args), nil
	case "bad":
		return nil, nil
	case "fail":
		topic, _ := args["topic"].(string)

		return ggql.NewSubscription(newSub(topic, nil), field, args), nil
	}
	return nil, fmt.Errorf("type Subscription does not have field %s", field)
}

func TestResolveSubscription(t *testing.T) {
	var log strings.Builder
	root := setupTestSongs(t, &log)

	var b strings.Builder

	src := `subscription tinEar {like{name,likes}}`
	expect := `{
  "data": null
}
`
	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)
	checkEqual(t, expect, b.String(), "result mismatch for %s", src)

	sep28 := &Date{Year: 2018, Month: 11, Day: 2}
	cnt, err := root.AddEvent("anything", &Song{Name: "Down In The Basement", Duration: 216, Release: sep28})
	checkNil(t, err, "AddEvent returned an error. %s", err)
	checkEqual(t, cnt, 1, "should be a single match")

	checkEqual(t, `{
  "likes": 0,
  "name": "Down In The Basement"
}
`, log.String(), "result mismatch for subscription")

	root.Unsubscribe("")
}

func TestResolveSubscriptionNone(t *testing.T) {
	var log strings.Builder
	root := setupTestSongs(t, &log)

	err := root.ParseString(`extend type Subscription {bad: Song}`)
	checkNil(t, err, "extend should not fail. %s", err)

	var b strings.Builder

	src := `subscription tinEar {bad{name}}`
	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, `{
  "data": null,
  "errors": [
    {
      "locations": [
        {
          "column": 15,
          "line": 1
        }
      ],
      "message": "resolve error: no subscriptions resolved"
    }
  ]
}
`, b.String(), "result mismatch for subscription")
}

func TestResolveSubscriptionError(t *testing.T) {
	var log strings.Builder
	root := setupTestSongs(t, &log)

	err := root.ParseString(`extend type Subscription {fail: Song}`)
	checkNil(t, err, "extend should not fail. %s", err)

	var b strings.Builder

	src := `subscription tinEar {fail{name}}`
	result := root.ResolveString(src, "", nil)
	_ = ggql.WriteJSONValue(&b, result, 2)

	checkEqual(t, `{
  "data": null
}
`, b.String(), "result mismatch for subscription")

	sep28 := &Date{Year: 2018, Month: 11, Day: 2}
	_, err = root.AddEvent("anything", &Song{Name: "Down In The Basement", Duration: 216, Release: sep28})
	checkNotNil(t, err, "expected a failure")
}
