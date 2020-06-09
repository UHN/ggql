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
	"testing"

	"github.com/uhn/ggql/pkg/ggql"
)

func TestRef(t *testing.T) {
	ref := &ggql.Ref{Base: ggql.Base{N: "Dummy"}}
	// Never add a ref directly. The type is only public so that code coverage
	// on unit test can be closer to 100%.

	checkEqual(t, "", ref.SDL(false), "Ref SDL() mismatch")
	checkEqual(t, "Dummy", ref.String(), "Ref String() mismatch")
	checkEqual(t, 0, ref.Rank(), "Ref Rank() mismatch")

	checkNil(t, ref.Write(nil, false), "check Write returns nil")
}
