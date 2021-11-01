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

package proxy

import (
	"bytes"
	"context"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/discovery"
	"github.com/starvn/turbo/log"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFactoryFunc(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	factory := FactoryFunc(func(cfg *config.EndpointConfig) (Proxy, error) { return DefaultFactory(logger).New(cfg) })

	if _, err := factory.New(&config.EndpointConfig{}); err != ErrNoBackends {
		t.Errorf("Expecting ErrNoBackends. Got: %v\n", err)
	}
}

func TestDefaultFactoryWithSubscriber(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	factory := DefaultFactoryWithSubscriber(logger, discovery.FixedSubscriberFactory)

	if _, err := factory.New(&config.EndpointConfig{}); err != ErrNoBackends {
		t.Errorf("Expecting ErrNoBackends. Got: %v\n", err)
	}
}

func TestDefaultFactory_noBackends(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	factory := DefaultFactory(logger)

	if _, err := factory.New(&config.EndpointConfig{}); err != ErrNoBackends {
		t.Errorf("Expecting ErrNoBackends. Got: %v\n", err)
	}
}

func TestNewDefaultFactory_ok(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := log.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	expectedResponse := Response{
		IsComplete: true,
		Data:       map[string]interface{}{"foo": "bar"},
	}
	expectedMethod := "SOME"
	expectedHost := "http://example.com/"
	expectedPath := "/foo"
	expectedURL := expectedHost + strings.TrimLeft(expectedPath, "/")

	URL, err := url.Parse(expectedHost)
	if err != nil {
		t.Errorf("building the sample url: %s\n", err.Error())
	}

	request := Request{
		Method: expectedMethod,
		Path:   expectedPath,
		URL:    URL,
		Body:   newDummyReadCloser(""),
	}

	assertion := func(ctx context.Context, request *Request) (*Response, error) {
		if request.URL.String() != expectedURL {
			t.Errorf("The middlewares did not update the request URL! want [%s], have [%s]\n", expectedURL, request.URL)
		}
		return &expectedResponse, nil
	}
	factory := NewDefaultFactory(func(_ *config.Backend) Proxy { return assertion }, logger)

	backend := config.Backend{
		URLPattern: expectedPath,
		Method:     expectedMethod,
	}
	endpointSingle := config.EndpointConfig{
		Backend: []*config.Backend{&backend},
	}
	endpointMulti := config.EndpointConfig{
		Backend:         []*config.Backend{&backend, &backend},
		ConcurrentCalls: 3,
	}
	serviceConfig := config.ServiceConfig{
		Version:   config.TurboConfigVersion,
		Endpoints: []*config.EndpointConfig{&endpointSingle, &endpointMulti},
		Timeout:   100 * time.Millisecond,
		Host:      []string{expectedHost},
	}
	if err := serviceConfig.Init(); err != nil {
		t.Errorf("Error during the config init: %s\n", err.Error())
	}

	proxyMulti, err := factory.New(&endpointMulti)
	if err != nil {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}

	response, err := proxyMulti(context.Background(), &request)
	if err != nil {
		t.Errorf("The proxy middleware propagated an unexpected error: %s\n", err.Error())
	}
	if !response.IsComplete || len(response.Data) != len(expectedResponse.Data) {
		t.Errorf("The proxy middleware propagated an unexpected error: %v\n", response)
	}

	proxySingle, err := factory.New(&endpointSingle)
	if err != nil {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}

	response, err = proxySingle(context.Background(), &request)
	if err != nil {
		t.Errorf("The proxy middleware propagated an unexpected error: %s\n", err.Error())
	}
	if !response.IsComplete || len(response.Data) != len(expectedResponse.Data) {
		t.Errorf("The proxy middleware propagated an unexpected error: %v\n", response)
	}
}
