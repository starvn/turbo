//go:build integration || !race
// +build integration !race

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

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/log"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadWithLogger(t *testing.T) {
	buff := new(bytes.Buffer)
	l, _ := log.NewLogger("DEBUG", buff, "")
	total, err := LoadWithLogger("./tests", ".so", RegisterHandler, l)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if total != 1 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 1", total)
	}

	var handler http.Handler

	hre := New(l, func(_ context.Context, _ config.ServiceConfig, h http.Handler) error {
		handler = h
		return nil
	})

	if err := hre(
		context.Background(),
		config.ServiceConfig{
			ExtraConfig: map[string]interface{}{
				Namespace: map[string]interface{}{
					"name": "sonic-server-example",
				},
			},
		},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("this handler should not been called")
		}),
	); err != nil {
		t.Error(err)
		return
	}

	req, _ := http.NewRequest("GET", "http://some.example.tld/path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	_ = resp.Body.Close()

	if string(b) != "Hello, \"/path\"" {
		t.Errorf("unexpected response body: %s", string(b))
	}

	fmt.Println(buff.String())
}
