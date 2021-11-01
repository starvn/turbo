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

// Package mux provides some basic implementations for building routers based on net/http mux
package mux

import (
	"context"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/route"
	"github.com/starvn/turbo/transport/http/server"
	"net/http"
	"strings"
)

const DefaultDebugPattern = "/__debug/"
const logPrefix = "[SERVICE: Mux]"

type RunServerFunc func(context.Context, config.ServiceConfig, http.Handler) error

type Config struct {
	Engine         Engine
	Middlewares    []HandlerMiddleware
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         log.Logger
	DebugPattern   string
	RunServer      RunServerFunc
}

type HandlerMiddleware interface {
	Handler(h http.Handler) http.Handler
}

func DefaultFactory(pf proxy.Factory, logger log.Logger) route.Factory {
	return factory{
		Config{
			Engine:         DefaultEngine(),
			Middlewares:    []HandlerMiddleware{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   pf,
			Logger:         logger,
			DebugPattern:   DefaultDebugPattern,
			RunServer:      server.RunServer,
		},
	}
}

func NewFactory(cfg Config) route.Factory {
	if cfg.DebugPattern == "" {
		cfg.DebugPattern = DefaultDebugPattern
	}
	return factory{cfg}
}

type factory struct {
	cfg Config
}

func (rf factory) New() route.Router {
	return rf.NewWithContext(context.Background())
}

func (rf factory) NewWithContext(ctx context.Context) route.Router {
	return httpRouter{rf.cfg, ctx, rf.cfg.RunServer}
}

type httpRouter struct {
	cfg       Config
	ctx       context.Context
	RunServer RunServerFunc
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (r httpRouter) Run(cfg config.ServiceConfig) {
	if cfg.Debug {
		debugHandler := DebugHandler(r.cfg.Logger)
		for _, method := range []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodConnect,
			http.MethodTrace,
		} {
			r.cfg.Engine.Handle(r.cfg.DebugPattern, method, debugHandler)
		}
	}
	r.cfg.Engine.Handle("/__health", "GET", http.HandlerFunc(HealthHandler))

	server.InitHTTPDefaultTransport(cfg)

	r.registerSonicEndpoints(cfg.Endpoints)

	if err := r.RunServer(r.ctx, cfg, r.handler()); err != nil {
		r.cfg.Logger.Error(logPrefix, err.Error())
	}

	r.cfg.Logger.Info(logPrefix, "Router execution ended")
}

func (r httpRouter) registerSonicEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error(logPrefix, "Calling the ProxyFactory", err.Error())
			continue
		}

		r.registerSonicEndpoint(c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r httpRouter) registerSonicEndpoint(method string, endpoint *config.EndpointConfig, handler http.HandlerFunc, totBackends int) {
	method = strings.ToTitle(method)
	path := endpoint.Endpoint
	if method != http.MethodGet && totBackends > 1 {
		if !route.IsValidSequentialEndpoint(endpoint) {
			r.cfg.Logger.Error(logPrefix, method, " endpoints with sequential proxy enabled only allow a non-GET in the last backend! Ignoring", path)
			return
		}
	}

	switch method {
	case http.MethodGet:
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodPatch:
	case http.MethodDelete:
	default:
		r.cfg.Logger.Error(logPrefix, "Unsupported method", method)
		return
	}
	r.cfg.Logger.Debug(logPrefix, "Registering the endpoint", method, path)
	r.cfg.Engine.Handle(path, method, handler)
}

func (r httpRouter) handler() http.Handler {
	var handler http.Handler = r.cfg.Engine
	for _, middleware := range r.cfg.Middlewares {
		r.cfg.Logger.Debug(logPrefix, "Adding the middleware", middleware)
		handler = middleware.Handler(handler)
	}
	return handler
}
