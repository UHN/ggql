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

type inputFieldList struct {
	dict map[string]*InputField
	list []*InputField
}

func (il *inputFieldList) Nth(i int) (n interface{}) {
	if 0 <= i && i < len(il.list) {
		n = il.list[i]
	}
	return
}

func (il *inputFieldList) Len() int {
	return len(il.list)
}

func (il *inputFieldList) add(fds ...*InputField) error {
	if il.dict == nil {
		il.dict = map[string]*InputField{}
	}
	for _, fd := range fds {
		name := fd.Name()
		if il.dict[name] != nil {
			return fmt.Errorf("%w input field %s", ErrDuplicate, name)
		}
		il.dict[name] = fd
		il.list = append(il.list, fd)
	}
	return nil
}

func (il *inputFieldList) get(name string) (i *InputField) {
	if il.dict != nil {
		i = il.dict[name]
	}
	return
}
