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

package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestConfig_rejectInvalidVersion(t *testing.T) {
	subject := ServiceConfig{}
	err := subject.Init()
	if err == nil || strings.Index(err.Error(), "unsupported version: 0 (want: 1)") != 0 {
		t.Error("Error expected. Got", err.Error())
	}
}

func TestConfig_rejectInvalidEndpoints(t *testing.T) {
	samples := []string{
		"/__debug",
		"/__debug/",
		"/__debug/foo",
		"/__debug/foo/bar",
	}

	for _, e := range samples {
		subject := ServiceConfig{Version: TurboConfigVersion, Endpoints: []*EndpointConfig{{Endpoint: e, Method: "GET"}}}
		err := subject.Init()
		if err == nil || err.Error() != fmt.Sprintf("ignoring the 'GET %s' endpoint, since it is invalid!!!", e) {
			t.Errorf("Unexpected error processing '%s': %v", e, err)
		}
	}
}

func TestConfig_initBackendURLMappings_ok(t *testing.T) {
	samples := []string{
		"sonic/{turbo}",
		"/sonic/{turbo1}",
		"/sonic.local/",
		"sonic/{turbo_56}/{sonic-5t6}?a={foo}&b={foo}",
		"sonic/{turbo_56}{sonic-5t6}?a={foo}&b={foo}",
		"sonic/turbo{sonic-5t6}?a={foo}&b={foo}",
		"{resp0_x}/{turbo1}/{turbo_56}{sonic-5t6}?a={turbo}&b={foo}",
		"{resp0_x}/{turbo1}/{JWT.foo}",
	}

	expected := []string{
		"/sonic/{{.Turbo}}",
		"/sonic/{{.Turbo1}}",
		"/sonic.local/",
		"/sonic/{{.Turbo_56}}/{{.Sonic-5t6}}?a={{.Foo}}&b={{.Foo}}",
		"/sonic/{{.Turbo_56}}{{.Sonic-5t6}}?a={{.Foo}}&b={{.Foo}}",
		"/sonic/turbo{{.Sonic-5t6}}?a={{.Foo}}&b={{.Foo}}",
		"/{{.Resp0_x}}/{{.Turbo1}}/{{.Turbo_56}}{{.Sonic-5t6}}?a={{.Turbo}}&b={{.Foo}}",
		"/{{.Resp0_x}}/{{.Turbo1}}/{{.JWT.foo}}",
	}

	backend := Backend{}
	endpoint := EndpointConfig{Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

	inputSet := map[string]interface{}{
		"turbo":     nil,
		"turbo1":    nil,
		"turbo_56":  nil,
		"sonic-5t6": nil,
		"foo":       nil,
	}

	for i := range samples {
		backend.URLPattern = samples[i]
		if err := subject.initBackendURLMappings(0, 0, inputSet); err != nil {
			t.Error(err)
		}
		if backend.URLPattern != expected[i] {
			t.Errorf("want: %s, have: %s\n", expected[i], backend.URLPattern)
		}
	}
}

func TestConfig_initBackendURLMappings_tooManyOutput(t *testing.T) {
	backend := Backend{URLPattern: "sonic/{turbo_56}/{sonic-5t6}?a={foo}&b={foo}"}
	endpoint := EndpointConfig{
		Method:   "GET",
		Endpoint: "/some/{turbo}",
		Backend:  []*Backend{&backend},
	}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

	inputSet := map[string]interface{}{
		"turbo": nil,
	}

	expectedErrMsg := "input and output params do not match. endpoint: GET /some/{turbo}, backend: 0. input: [turbo], output: [foo sonic-5t6 turbo_56]"

	err := subject.initBackendURLMappings(0, 0, inputSet)
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestConfig_initBackendURLMappings_undefinedOutput(t *testing.T) {
	backend := Backend{URLPattern: "sonic/{turbo_56}/{sonic-5t6}?a={foo}&b={foo}"}
	endpoint := EndpointConfig{Endpoint: "/", Method: "GET", Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

	inputSet := map[string]interface{}{
		"turbo": nil,
		"sonic": nil,
		"foo":   nil,
	}

	expectedErrMsg := "undefined output param 'sonic-5t6'! endpoint: GET /, backend: 0. input: [foo sonic turbo], output: [foo sonic-5t6 turbo_56]"
	err := subject.initBackendURLMappings(0, 0, inputSet)
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("error expected. have: %v", err)
	}
}

func TestConfig_init(t *testing.T) {
	sonicBackend := Backend{
		URLPattern: "/__debug/sonic",
	}
	sonicEndpoint := EndpointConfig{
		Endpoint:       "/sonic",
		Method:         "post",
		Timeout:        1500 * time.Millisecond,
		CacheTTL:       6 * time.Hour,
		Backend:        []*Backend{&sonicBackend},
		OutputEncoding: "some_render",
	}

	githubBackend := Backend{
		URLPattern: "/",
		Host:       []string{"https://api.github.com"},
		AllowList:  []string{"authorizations_url", "code_search_url"},
	}
	githubEndpoint := EndpointConfig{
		Endpoint: "/github",
		Timeout:  1500 * time.Millisecond,
		CacheTTL: 6 * time.Hour,
		Backend:  []*Backend{&githubBackend},
	}

	userBackend := Backend{
		URLPattern: "/users/{user}",
		Host:       []string{"https://jsonplaceholder.typicode.com"},
		Mapping:    map[string]string{"email": "personal_email"},
	}
	rssBackend := Backend{
		URLPattern: "/users/{user}",
		Host:       []string{"https://jsonplaceholder.typicode.com"},
		Encoding:   "rss",
	}
	postBackend := Backend{
		URLPattern: "/posts/{user}",
		Host:       []string{"https://jsonplaceholder.typicode.com"},
		Group:      "posts",
		Encoding:   "xml",
	}
	userEndpoint := EndpointConfig{
		Endpoint: "/users/{user}",
		Backend:  []*Backend{&userBackend, &rssBackend, &postBackend},
	}

	subject := ServiceConfig{
		Version:   TurboConfigVersion,
		Timeout:   5 * time.Second,
		CacheTTL:  30 * time.Minute,
		Host:      []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{&sonicEndpoint, &githubEndpoint, &userEndpoint},
	}

	if err := subject.Init(); err != nil {
		t.Error("Error at the configuration init:", err.Error())
	}

	if len(sonicBackend.Host) != 1 || sonicBackend.Host[0] != subject.Host[0] {
		t.Error("Default hosts not applied to the sonic backend", sonicBackend.Host)
	}

	for level, method := range map[string]string{
		"userBackend":  userBackend.Method,
		"postBackend":  postBackend.Method,
		"userEndpoint": userEndpoint.Method,
	} {
		if method != "GET" {
			t.Errorf("Default method not applied at %s. Get: %s", level, method)
		}
	}

	if sonicBackend.Method != "post" {
		t.Error("unexpected sonicBackend")
	}

	if userBackend.Timeout != subject.Timeout {
		t.Error("default timeout not applied to the userBackend")
	}

	if userEndpoint.CacheTTL != subject.CacheTTL {
		t.Error("default CacheTTL not applied to the userEndpoint")
	}

	hash, err := subject.Hash()
	if err != nil {
		t.Error(err.Error())
	}

	if hash != "Nvw/S+RNSFKG9wOO5bFSSlyuT25j0wwYllEDqzNdBPA=" {
		t.Errorf("unexpected hash: %s", hash)
	}
}

func TestConfig_initKONoBackends(t *testing.T) {
	subject := ServiceConfig{
		Version: TurboConfigVersion,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/sonic",
				Method:   "POST",
				Backend:  []*Backend{},
			},
		},
	}

	if err := subject.Init(); err == nil ||
		err.Error() != "ignoring the 'POST /sonic' endpoint, since it has 0 backends defined!" {
		t.Error("Unexpected error at the configuration init!", err)
	}
}

func TestConfig_initKOMultipleBackendsForNoopEncoder(t *testing.T) {
	subject := ServiceConfig{
		Version: TurboConfigVersion,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint:       "/sonic",
				Method:         "post",
				OutputEncoding: "no-op",
				Backend: []*Backend{
					{
						Encoding: "no-op",
					},
					{
						Encoding: "no-op",
					},
				},
			},
		},
	}

	if err := subject.Init(); err != errInvalidNoOpEncoding {
		t.Error("Expecting an error at the configuration init!", err)
	}
}

func TestConfig_initKOInvalidHost(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The init process did not panic with an invalid host!")
		}
	}()
	subject := ServiceConfig{
		Version: TurboConfigVersion,
		Host:    []string{"http://127.0.0.1:8080http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/sonic",
				Method:   "post",
				Backend:  []*Backend{},
			},
		},
	}

	err := subject.Init()
	if err != nil {
		return
	}
}

func TestConfig_initKOInvalidDebugPattern(t *testing.T) {
	dp := debugPattern

	debugPattern = "a(b"
	subject := ServiceConfig{
		Version: TurboConfigVersion,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/__debug/sonic",
				Method:   "GET",
				Backend:  []*Backend{},
			},
		},
	}

	if err := subject.Init(); err == nil ||
		err.Error() != "ignoring the 'GET /__debug/sonic' endpoint due to a parsing error: error parsing regexp: missing closing ): `a(b`" {
		t.Error("Expecting an error at the configuration init!", err)
	}

	debugPattern = dp
}
