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

package mux

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEngine(t *testing.T) {
	e := DefaultEngine()

	for _, method := range []string{"PUT", "POST", "DELETE"} {
		e.Handle("/", method, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "hi there!", http.StatusTeapot)
		}))
	}

	for _, tc := range []struct {
		method string
		status int
	}{
		{status: http.StatusTeapot, method: "PUT"},
		{status: http.StatusTeapot, method: "POST"},
		{status: http.StatusTeapot, method: "DELETE"},
		{status: http.StatusMethodNotAllowed, method: "GET"},
	} {
		req, _ := http.NewRequest(tc.method, "http://127.0.0.1:8081/_mux_endpoint?b=1&c[]=x&c[]=y&d=1&d=2&a=42", ioutil.NopCloser(&bytes.Buffer{}))

		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)

		if sc := w.Result().StatusCode; tc.status != sc {
			t.Error("unexpected status code:", sc)
		}
	}
}
