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

type fieldList struct {
	dict map[string]*FieldDef
	list []*FieldDef
}

func (fl *fieldList) Nth(i int) (n interface{}) {
	if 0 <= i && i < len(fl.list) {
		n = fl.list[i]
	}
	return
}

func (fl *fieldList) Len() int {
	return len(fl.list)
}

func (fl *fieldList) add(fds ...*FieldDef) error {
	if fl.dict == nil {
		fl.dict = map[string]*FieldDef{}
	}
	for _, fd := range fds {
		name := fd.Name()
		if fl.dict[name] != nil {
			return fmt.Errorf("%w field %s", ErrDuplicate, name)
		}
		fl.dict[name] = fd
		fl.list = append(fl.list, fd)
	}
	return nil
}

func (fl *fieldList) get(name string) (f *FieldDef) {
	if fl.dict != nil {
		f = fl.dict[name]
	}
	return
}
