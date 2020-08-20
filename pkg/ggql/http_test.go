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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/uhn/ggql/pkg/ggql"
)

// When parsing from a HTTP request the last character read will be provided
// along with an EOF. This test make sure the last character is not
// ignored. Tried io.Pipe() and that does not have the same behavior so
// resorted to HTTP client server.
func TestParseHTTP(t *testing.T) {
	ggql.Sort = true
	port := 11954
	var log strings.Builder
	root := setupTestSongs(t, &log)

	http.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		defer func() { _ = req.Body.Close() }()

		result := root.ResolveReader(req.Body, "", nil)

		_ = ggql.WriteJSONValue(w, result, -1)
	})
	go func() { _ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil) }()

	u := fmt.Sprintf("http://localhost:%d/graphql", port)

	cx, cf := context.WithTimeout(context.Background(), 2*time.Second)
	defer cf()
	req, err := http.NewRequestWithContext(cx, "POST", u, bytes.NewBufferString(`{artist(name: "Fazerdaze"){name}}`))
	checkNil(t, err, "POST request creation failed. %s", err)
	req.Header.Add("Content-Type", "application/graphql")
	var res *http.Response
	res, err = (&http.Client{}).Do(req)
	checkNil(t, err, "POST failed. %s", err)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	checkEqual(t, `{"data":{"artist":{"name":"Fazerdaze"}}}`, string(body), "parsed vs given")
}
