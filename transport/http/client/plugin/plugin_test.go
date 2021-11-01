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
	"github.com/starvn/turbo/transport/http/client"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestLoadWithLogger(t *testing.T) {
	buff := new(bytes.Buffer)
	l, _ := log.NewLogger("DEBUG", buff, "")
	total, err := LoadWithLogger("./tests", ".so", RegisterClient, l)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if total != 1 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 1", total)
	}

	hre := HTTPRequestExecutor(l, func(_ *config.Backend) client.HTTPRequestExecutor {
		t.Error("this factory should not been called")
		return nil
	})

	h := hre(&config.Backend{
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"name": "sonic-client-example",
			},
		},
	})

	req, _ := http.NewRequest("GET", "http://some.example.tld/path", nil)
	resp, err := h(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
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
