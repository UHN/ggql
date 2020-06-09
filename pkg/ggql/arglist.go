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

type argList struct {
	dict map[string]*Arg
	list []*Arg
}

func (al *argList) Nth(i int) (n interface{}) {
	if 0 <= i && i < len(al.list) {
		n = al.list[i]
	}
	return
}

func (al *argList) Len() int {
	return len(al.list)
}

func (al *argList) add(fds ...*Arg) error {
	if al.dict == nil {
		al.dict = map[string]*Arg{}
	}
	for _, fd := range fds {
		name := fd.Name()
		if al.dict[name] != nil {
			return fmt.Errorf("%w argument %s", ErrDuplicate, name)
		}
		al.dict[name] = fd
		al.list = append(al.list, fd)
	}
	return nil
}

func (al *argList) get(name string) (a *Arg) {
	if al.dict != nil {
		a = al.dict[name]
	}
	return
}
