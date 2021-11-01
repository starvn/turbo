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

// Package httptreemux provides some basic implementations for building routers based on dimfeld/httptreemux
package httptreemux

import (
	"github.com/dimfeld/httptreemux/v5"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/route"
	"github.com/starvn/turbo/route/mux"
	"github.com/starvn/turbo/transport/http/server"
	"net/http"
	"strings"
)

func DefaultFactory(pf proxy.Factory, logger log.Logger) route.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger))
}

func DefaultConfig(pf proxy.Factory, logger log.Logger) mux.Config {
	return mux.Config{
		Engine:         NewEngine(httptreemux.NewContextMux()),
		Middlewares:    []mux.HandlerMiddleware{},
		HandlerFactory: mux.CustomEndpointHandler(mux.NewRequestBuilder(ParamsExtractor)),
		ProxyFactory:   pf,
		Logger:         logger,
		DebugPattern:   "/__debug/{params}",
		RunServer:      server.RunServer,
	}
}

func ParamsExtractor(r *http.Request) map[string]string {
	params := map[string]string{}
	for key, value := range httptreemux.ContextParams(r.Context()) {
		params[strings.Title(key)] = value
	}
	return params
}

func NewEngine(m *httptreemux.ContextMux) Engine {
	return Engine{m}
}

type Engine struct {
	r *httptreemux.ContextMux
}

func (g Engine) Handle(pattern, method string, handler http.Handler) {
	g.r.Handle(method, pattern, handler.(http.HandlerFunc))
}

func (g Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.r.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}
