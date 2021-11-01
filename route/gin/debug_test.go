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

package gin

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/starvn/turbo/log"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDebugHandler(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	router := gin.New()
	router.GET("/_gin_endpoint/:param", DebugHandler(logger))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8088/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body, ioErr := ioutil.ReadAll(w.Result().Body)
	if ioErr != nil {
		t.Error("reading a response:", err.Error())
		return
	}
	_ = w.Result().Body.Close()

	expectedBody := "{\"message\":\"pong\"}"

	content := string(body)
	if w.Result().Header.Get("Cache-Control") != "" {
		t.Error("Cache-Control error:", w.Result().Header.Get("Cache-Control"))
	}
	if w.Result().Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Sonic") != "" {
		t.Error("X-Sonic error:", w.Result().Header.Get("X-Sonic"))
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
	if content != expectedBody {
		t.Error("Unexpected body:", content, "expected:", expectedBody)
	}
}
