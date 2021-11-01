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
	"io"
	"net/url"
	"testing"
)

func TestLoadModifiers(t *testing.T) {
	total, err := LoadModifiers("./tests", ".so", RegisterModifier)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if total != 1 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 1", total)
	}

	modFactory, ok := GetRequestModifier("turbo-request-modifier-example")
	if !ok {
		t.Error("modifier factory not found in the register")
		t.Fail()
	}

	modifier := modFactory(map[string]interface{}{})

	input := requestWrapper{path: "/bar"}

	tmp, err := modifier(input)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}

	output, ok := tmp.(RequestWrapper)
	if !ok {
		t.Error("unexpected result type")
		t.Fail()
	}

	if res := output.Path(); res != "/bar/fooo" {
		t.Errorf("unexpected result path. have %s, want /bar/fooo", res)
	}
}

type RequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
}

type requestWrapper struct {
	method  string
	url     *url.URL
	query   url.Values
	path    string
	body    io.ReadCloser
	params  map[string]string
	headers map[string][]string
}

func (r requestWrapper) Method() string               { return r.method }
func (r requestWrapper) URL() *url.URL                { return r.url }
func (r requestWrapper) Query() url.Values            { return r.query }
func (r requestWrapper) Path() string                 { return r.path }
func (r requestWrapper) Body() io.ReadCloser          { return r.body }
func (r requestWrapper) Params() map[string]string    { return r.params }
func (r requestWrapper) Headers() map[string][]string { return r.headers }
