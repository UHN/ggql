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

import (
	"sort"
	"strings"
)

type typeList struct {
	dict map[string]Type
	list []Type
}

func newTypeList() *typeList {
	return &typeList{dict: map[string]Type{}}
}

func (tl *typeList) Nth(i int) (n interface{}) {
	if 0 <= i && i < len(tl.list) {
		n = tl.list[i]
	}
	return
}

func (tl *typeList) Len() int {
	return len(tl.list)
}

func (tl *typeList) add(ts ...Type) {
	for _, t := range ts {
		tl.dict[t.Name()] = t
		tl.list = append(tl.list, t)
	}
	sort.Slice(tl.list, func(i, j int) bool {
		ti := tl.list[i]
		tj := tl.list[j]
		if ti.Rank() == tj.Rank() {
			return 0 > strings.Compare(ti.Name(), tj.Name())
		}
		return ti.Rank() < (tj.Rank())
	})
}

func (tl *typeList) get(name string) (t Type) {
	if tl.dict != nil {
		t = tl.dict[name]
	}
	return
}

func (tl *typeList) dup() *typeList {
	d := typeList{
		list: make([]Type, len(tl.list)),
		dict: map[string]Type{},
	}
	copy(d.list, tl.list)
	if tl.dict != nil {
		for name, t := range tl.dict {
			d.dict[name] = t
		}
	}
	return &d
}
