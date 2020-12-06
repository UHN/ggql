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

import "fmt"

type enumValueList struct {
	dict map[string]*EnumValue
	list []*EnumValue
}

func (el *enumValueList) Nth(i int) (n interface{}) {
	if 0 <= i && i < len(el.list) {
		n = el.list[i]
	}
	return
}

func (el *enumValueList) Len() int {
	return len(el.list)
}

func (el *enumValueList) add(evs ...*EnumValue) error {
	if el.dict == nil {
		el.dict = map[string]*EnumValue{}
	}
	for _, ev := range evs {
		name := string(ev.Value)
		if el.dict[name] != nil {
			return fmt.Errorf("%w enum value %s", ErrDuplicate, name)
		}
		el.dict[name] = ev
		el.dict[string(ev.Value)] = ev
		el.list = append(el.list, ev)
	}
	return nil
}

func (el *enumValueList) has(v Symbol) bool {
	if el.dict != nil {
		if _, ok := el.dict[string(v)]; ok {
			return true
		}
	}
	return false
}
