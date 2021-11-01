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
	"context"
	"github.com/gin-gonic/gin"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/encoding"
	"github.com/starvn/turbo/proxy"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestRender_Negotiated_ok(t *testing.T) {
	type A struct {
		B string
	}
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"content": A{B: "sonic"}},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: NEGOTIATE,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain", "application/x-yaml; charset=utf-8", "content:\n  b: sonic\n"},
		{"none", "", "application/json; charset=utf-8", `{"content":{"B":"sonic"}}`},
		{"json", "application/json", "application/json; charset=utf-8", `{"content":{"B":"sonic"}}`},
		{"xml", "application/xml", "application/xml; charset=utf-8", `<A><B>sonic</B></A>`},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioErr := ioutil.ReadAll(w.Result().Body)
		if ioErr != nil {
			t.Error("reading response body:", ioErr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Cache-Control") != "public, max-age=21600" {
			t.Error(testData[0], "Cache-Control error:", w.Result().Header.Get("Cache-Control"))
		}
		if w.Result().Header.Get("Content-Type") != testData[2] {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Sonic") != "Version undefined" {
			t.Error(testData[0], "X-Sonic error:", w.Result().Header.Get("X-Sonic"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != testData[3] {
			t.Error(testData[0], "Unexpected body:", content, "expected:", testData[3])
		}
	}
}

func TestRender_Negotiated_noData(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			Data: map[string]interface{}{},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: NEGOTIATE,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain", "application/x-yaml; charset=utf-8", "{}\n"},
		{"none", "", "application/json; charset=utf-8", "{}"},
		{"json", "application/json", "application/json; charset=utf-8", "{}"},
		{"xml", "application/xml", "application/xml; charset=utf-8", ""},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(w.Result().Body)

		body, ioErr := ioutil.ReadAll(w.Result().Body)
		if ioErr != nil {
			t.Error("reading response body:", ioErr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Content-Type") != testData[2] {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Sonic") != "Version undefined" {
			t.Error(testData[0], "X-Sonic error:", w.Result().Header.Get("X-Sonic"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != testData[3] {
			t.Error(testData[0], "Unexpected body:", content, "expected:", testData[3])
		}
	}
}

func TestRender_Negotiated_noResponse(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: NEGOTIATE,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain", "application/x-yaml; charset=utf-8", "{}\n"},
		{"none", "", "application/json; charset=utf-8", "{}"},
		{"json", "application/json", "application/json; charset=utf-8", "{}"},
		{"xml", "application/xml", "application/xml; charset=utf-8", ""},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(w.Result().Body)

		body, ioErr := ioutil.ReadAll(w.Result().Body)
		if ioErr != nil {
			t.Error("reading response body:", ioErr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Content-Type") != testData[2] {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Sonic") != "Version undefined" {
			t.Error(testData[0], "X-Sonic error:", w.Result().Header.Get("X-Sonic"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != testData[3] {
			t.Error(testData[0], "Unexpected body:", content, "expected:", testData[3])
		}
	}
}

func TestRender_unknown(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"sonic": "turbo"},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: "unknown",
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	expectedHeader := "application/json; charset=utf-8"
	expectedBody := `{"sonic":"turbo"}`

	for _, testData := range [][]string{
		{"plain", "text/plain"},
		{"none", ""},
		{"json", "application/json"},
		{"unknown", "unknown"},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(w.Result().Body)

		body, ioErr := ioutil.ReadAll(w.Result().Body)
		if ioErr != nil {
			t.Error("reading response body:", ioErr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Cache-Control") != "public, max-age=21600" {
			t.Error(testData[0], "Cache-Control error:", w.Result().Header.Get("Cache-Control"))
		}
		if w.Result().Header.Get("Content-Type") != expectedHeader {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Sonic") != "Version undefined" {
			t.Error(testData[0], "X-Sonic error:", w.Result().Header.Get("X-Sonic"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != expectedBody {
			t.Error(testData[0], "Unexpected body:", content, "expected:", expectedBody)
		}
	}
}

func TestRender_string(t *testing.T) {
	expectedContent := "sonic"
	expectedHeader := "text/plain; charset=utf-8"

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"content": expectedContent},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.STRING,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain"},
		{"none", ""},
		{"json", "application/json"},
		{"unknown", "unknown"},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(w.Result().Body)

		body, ioErr := ioutil.ReadAll(w.Result().Body)
		if ioErr != nil {
			t.Error("reading response body:", ioErr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Cache-Control") != "public, max-age=21600" {
			t.Error(testData[0], "Cache-Control error:", w.Result().Header.Get("Cache-Control"))
		}
		if w.Result().Header.Get("Content-Type") != expectedHeader {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Sonic") != "Version undefined" {
			t.Error(testData[0], "X-Sonic error:", w.Result().Header.Get("X-Sonic"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != expectedContent {
			t.Error(testData[0], "Unexpected body:", content, "expected:", expectedContent)
		}
	}
}

func TestRender_string_noData(t *testing.T) {
	expectedContent := ""
	expectedHeader := "text/plain; charset=utf-8"

	for k, p := range []proxy.Proxy{
		func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{
				IsComplete: false,
				Data:       map[string]interface{}{"content": 42},
			}, nil
		},
		func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{
				IsComplete: false,
				Data:       map[string]interface{}{},
			}, nil
		},
		func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return nil, nil
		},
	} {
		endpoint := &config.EndpointConfig{
			Timeout:        time.Second,
			CacheTTL:       6 * time.Hour,
			QueryString:    []string{"b"},
			OutputEncoding: encoding.STRING,
		}

		gin.SetMode(gin.TestMode)
		server := gin.New()
		server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(w.Result().Body)

		body, ioErr := ioutil.ReadAll(w.Result().Body)
		if ioErr != nil {
			t.Error("reading response body:", ioErr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Content-Type") != expectedHeader {
			t.Error(k, "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Sonic") != "Version undefined" {
			t.Error(k, "X-Sonic error:", w.Result().Header.Get("X-Sonic"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(k, "Unexpected status code:", w.Result().StatusCode)
		}
		if content != expectedContent {
			t.Error(k, "Unexpected body:", content, "expected:", expectedContent)
		}
	}
}

func TestRegisterRender(t *testing.T) {
	var total int
	expected := &proxy.Response{IsComplete: true, Data: map[string]interface{}{"a": "b"}}
	name := "test render"

	RegisterRender(name, func(_ *gin.Context, resp *proxy.Response) {
		*resp = *expected
		total++
	})

	subject := getRender(&config.EndpointConfig{OutputEncoding: name})

	var c *gin.Context
	resp := proxy.Response{}
	subject(c, &resp)

	if !reflect.DeepEqual(resp, *expected) {
		t.Error("unexpected response", resp)
	}

	if total != 1 {
		t.Error("the render was called an unexpected amount of times:", total)
	}
}

func TestRender_noop(t *testing.T) {
	expectedContent := "sonic"
	expectedHeader := "text/plain; charset=utf-8"
	expectedSetCookieValue := []string{"test1=test1", "test2=test2"}

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			Metadata: proxy.Metadata{
				StatusCode: 200,
				Headers: map[string][]string{
					"Content-Type": {expectedHeader},
					"Set-Cookie":   {"test1=test1", "test2=test2"},
				},
			},
			Io: bytes.NewBufferString(expectedContent),
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.NOOP,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(w.Result().Body)

	body, ioErr := ioutil.ReadAll(w.Result().Body)
	if ioErr != nil {
		t.Error("reading response body:", ioErr)
		return
	}

	content := string(body)
	if w.Result().Header.Get("Content-Type") != expectedHeader {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Sonic") != "Version undefined" {
		t.Error("X-Sonic error:", w.Result().Header.Get("X-Sonic"))
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
	if content != expectedContent {
		t.Error("Unexpected body:", content, "expected:", expectedContent)
	}
	gotCookie := w.Header()["Set-Cookie"]
	if !reflect.DeepEqual(gotCookie, expectedSetCookieValue) {
		t.Error("Unexpected Set-Cookie header:", gotCookie, "expected:", expectedSetCookieValue)
	}
}

func TestRender_noop_nilBody(t *testing.T) {
	expectedContent := ""
	expectedHeader := ""

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{IsComplete: true}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.NOOP,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(w.Result().Body)

	body, ioErr := ioutil.ReadAll(w.Result().Body)
	if ioErr != nil {
		t.Error("reading response body:", ioErr)
		return
	}

	content := string(body)
	if w.Result().Header.Get("Content-Type") != expectedHeader {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Sonic") != "Version undefined" {
		t.Error("X-Sonic error:", w.Result().Header.Get("X-Sonic"))
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
	if content != expectedContent {
		t.Error("Unexpected body:", content, "expected:", expectedContent)
	}
}

func TestRender_noop_nilResponse(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.NOOP,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Result().Header.Get("Content-Type") != "" {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Sonic") != "Version undefined" {
		t.Error("X-Sonic error:", w.Result().Header.Get("X-Sonic"))
	}
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
}
