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

package gin

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

	"github.com/gin-gonic/gin"
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
		Port: 8072,
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/some",
				Method:   "GET",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/some",
				Method:   "post",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/some",
				Method:   "put",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/some",
				Method:   "PATCH",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/some",
				Method:   "DELETE",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
		},
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"auto_options": true,
			},
		},
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, endpoint := range serviceCfg.Endpoints {
		req, _ := http.NewRequest(strings.ToTitle(endpoint.Method), fmt.Sprintf("http://127.0.0.1:8072%s", endpoint.Endpoint), nil)
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
			t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
		}
		if resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderCompleteResponseValue {
			t.Error(server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
		}
		if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			t.Error("Content-Type error:", resp.Header.Get("Content-Type"))
		}
		if resp.Header.Get("X-Sonic") != "Version undefined" {
			t.Error("X-Sonic error:", resp.Header.Get("X-Sonic"))
		}
		if resp.StatusCode != http.StatusOK {
			t.Error("Unexpected status code:", resp.StatusCode)
		}
		if content != expectedBody {
			t.Error("Unexpected body:", content, "expected:", expectedBody)
		}
	}

	req, _ := http.NewRequest("OPTIONS", "http://127.0.0.1:8072/some", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Making the request:", err.Error())
		return
	}

	if allow := resp.Header.Get("Allow"); allow != "DELETE, GET, PATCH, POST, PUT" {
		t.Errorf("unexpected options response: '%s'", allow)
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

	r := DefaultFactory(noopProxyFactory(map[string]interface{}{"sonic": "turbo"}), logger).NewWithContext(ctx)

	serviceCfg := config.ServiceConfig{
		Debug: true,
		Port:  8073,
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
		req, _ := http.NewRequest(subject[0], fmt.Sprintf("http://127.0.0.1:8073/%s", subject[1]), nil)
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
		Port:  8074,
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
		req, _ := http.NewRequest(subject[0], fmt.Sprintf("http://127.0.0.1:8074/%s", subject[1]), nil)
		req.Header.Set("Content-Type", "application/json")
		checkResponseIs404(t, req)
	}
}

func TestRunServer_ko(t *testing.T) {
	buff := new(bytes.Buffer)
	logger, err := log.NewLogger("ERROR", buff, "")
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
			Engine:         gin.Default(),
			Middlewares:    []gin.HandlerFunc{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   pf,
			Logger:         logger,
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
	expectedBody := "404 page not found"
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
		t.Error(req.URL.String(), "Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Error(req.URL.String(), "Content-Type error:", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Sonic") != "" {
		t.Error(req.URL.String(), "X-Sonic error:", resp.Header.Get("X-Sonic"))
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error(req.URL.String(), "Unexpected status code:", resp.StatusCode)
	}
	if content != expectedBody {
		t.Error(req.URL.String(), "Unexpected body:", content, "expected:", expectedBody)
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
