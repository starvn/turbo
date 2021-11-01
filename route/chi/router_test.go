//go:build !race
// +build !race

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

package chi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/transport/http/server"
)

func TestDefaultFactory_ok(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(5 * time.Millisecond)
	}()

	r := DefaultFactory(noopProxyFactory(map[string]interface{}{"sonic": "turbo"}), logger).NewWithContext(ctx)
	expectedBody := "{\"sonic\":\"turbo\"}"

	serviceCfg := config.ServiceConfig{
		Port: 8062,
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/get",
				Method:   "GET",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/get",
				Method:   "POST",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/post",
				Method:   "Post",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/put",
				Method:   "put",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/patch",
				Method:   "PATCH",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/delete",
				Method:   "DELETE",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
		},
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, endpoint := range serviceCfg.Endpoints {
		req, _ := http.NewRequest(strings.ToTitle(endpoint.Method), fmt.Sprintf("http://127.0.0.1:8062%s", endpoint.Endpoint), nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error("Making the request:", err.Error())
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		body, ioErr := ioutil.ReadAll(resp.Body)
		if ioErr != nil {
			t.Error("Reading the response:", ioErr.Error())
			return
		}
		content := string(body)
		if resp.Header.Get("Cache-Control") != "" {
			t.Error(endpoint.Endpoint, "Cache-Control error:", resp.Header.Get("Cache-Control"))
		}
		if resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderCompleteResponseValue {
			t.Error(server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
		}
		if resp.Header.Get("Content-Type") != "application/json" {
			t.Error(endpoint.Endpoint, "Content-Type error:", resp.Header.Get("Content-Type"))
		}
		if resp.Header.Get("X-Sonic") != "Version undefined" {
			t.Error(endpoint.Endpoint, "X-Sonic error:", resp.Header.Get("X-Sonic"))
		}
		if resp.StatusCode != http.StatusOK {
			t.Error(endpoint.Endpoint, "Unexpected status code:", resp.StatusCode)
		}
		if content != expectedBody {
			t.Error(endpoint.Endpoint, "Unexpected body:", content, "expected:", expectedBody)
		}
	}
}

func TestDefaultFactory_ko(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(5 * time.Millisecond)
	}()

	r := NewFactory(Config{
		Engine:         chi.NewRouter(),
		Middlewares:    chi.Middlewares{},
		HandlerFactory: NewEndpointHandler,
		ProxyFactory:   noopProxyFactory(map[string]interface{}{"sonic": "turbo"}),
		Logger:         logger,
		RunServer:      server.RunServer,
	}).NewWithContext(ctx)

	serviceCfg := config.ServiceConfig{
		Debug: true,
		Port:  8063,
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/ignored",
				Method:   "GETTT",
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/empty",
				Method:   "GETTT",
				Backend:  []*config.Backend{},
			},
			{
				Endpoint: "/also-ignored",
				Method:   "PUT",
				Backend: []*config.Backend{
					{},
					{},
				},
			},
		},
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, subject := range [][]string{
		{"GET", "ignored"},
		{"GET", "empty"},
		{"PUT", "also-ignored"},
	} {
		req, _ := http.NewRequest(subject[0], fmt.Sprintf("http://127.0.0.1:8063/%s", subject[1]), nil)
		req.Header.Set("Content-Type", "application/json")
		checkResponseIs404(t, req)
	}
}

func TestDefaultFactory_proxyFactoryCrash(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(5 * time.Millisecond)
	}()

	r := DefaultFactory(erroredProxyFactory{fmt.Errorf("%s", "crash!!!")}, logger).NewWithContext(ctx)

	serviceCfg := config.ServiceConfig{
		Debug: true,
		Port:  8064,
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/ignored",
				Method:   "GET",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
		},
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, subject := range [][]string{{"GET", "ignored"}, {"PUT", "also-ignored"}} {
		req, _ := http.NewRequest(subject[0], fmt.Sprintf("http://127.0.0.1:8064/%s", subject[1]), nil)
		req.Header.Set("Content-Type", "application/json")
		checkResponseIs404(t, req)
	}
}

func TestRunServer_ko(t *testing.T) {
	buff := new(bytes.Buffer)
	logger, err := log.NewLogger("DEBUG", buff, "")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	errorMsg := "runServer error"
	runServerFunc := func(_ context.Context, _ config.ServiceConfig, _ http.Handler) error {
		return errors.New(errorMsg)
	}

	pf := noopProxyFactory(map[string]interface{}{"sonic": "turbo"})
	r := NewFactory(
		Config{
			Engine:         chi.NewRouter(),
			Middlewares:    chi.Middlewares{},
			HandlerFactory: NewEndpointHandler,
			ProxyFactory:   pf,
			Logger:         logger,
			DebugPattern:   ChiDefaultDebugPattern,
			RunServer:      runServerFunc,
		},
	).New()

	serviceCfg := config.ServiceConfig{}
	r.Run(serviceCfg)
	re := regexp.MustCompile(errorMsg)
	if !re.MatchString(string(buff.Bytes())) {
		t.Errorf("the logger doesn't contain the expected msg: %s", buff.Bytes())
	}
}

func checkResponseIs404(t *testing.T, req *http.Request) {
	expectedBody := "404 page not found\n"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Making the request:", err.Error())
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, ioErr := ioutil.ReadAll(resp.Body)
	if ioErr != nil {
		t.Error("Reading the response:", ioErr.Error())
		return
	}
	content := string(body)

	if resp.Header.Get("Cache-Control") != "" {
		t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderIncompleteResponseValue {
		t.Error(req.URL.String(), server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
	}
	if resp.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Error("Content-Type error:", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Sonic") != "" {
		t.Error("X-Sonic error:", resp.Header.Get("X-Sonic"))
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Unexpected status code:", resp.StatusCode)
	}
	if content != expectedBody {
		t.Error("Unexpected body:", content, "expected:", expectedBody)
	}
}

type noopProxyFactory map[string]interface{}

func (n noopProxyFactory) New(_ *config.EndpointConfig) (proxy.Proxy, error) {
	return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       n,
		}, nil
	}, nil
}

type erroredProxyFactory struct {
	Error error
}

func (e erroredProxyFactory) New(_ *config.EndpointConfig) (proxy.Proxy, error) {
	return proxy.NoopProxy, e.Error
}
