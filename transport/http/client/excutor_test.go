/*
 * Copyright (c) 2021 Huy Duc Dao
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHTTPRequestExecutor(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	re := DefaultHTTPRequestExecutor(NewHTTPClient)

	req, _ := http.NewRequest("GET", ts.URL, ioutil.NopCloser(&bytes.Buffer{}))

	resp, err := re(context.Background(), req)

	if err != nil {
		t.Error("unexpected error:", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", resp.StatusCode)
	}
}
