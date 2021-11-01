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

// Package chi provides some basic implementations for building routers based on go-chi/chi
package chi

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/route"
	"github.com/starvn/turbo/route/mux"
	"github.com/starvn/turbo/transport/http/server"
	"net/http"
	"strings"
)

const ChiDefaultDebugPattern = "/__debug/"

const logPrefix = "[SERVICE: Chi]"

type RunServerFunc func(context.Context, config.ServiceConfig, http.Handler) error

type Config struct {
	Engine         chi.Router
	Middlewares    chi.Middlewares
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         log.Logger
	DebugPattern   string
	RunServer      RunServerFunc
}

func DefaultFactory(proxyFactory proxy.Factory, logger log.Logger) route.Factory {
	return NewFactory(
		Config{
			Engine:         chi.NewRouter(),
			Middlewares:    chi.Middlewares{middleware.Logger},
			HandlerFactory: NewEndpointHandler,
			ProxyFactory:   proxyFactory,
			Logger:         logger,
			DebugPattern:   ChiDefaultDebugPattern,
			RunServer:      server.RunServer,
		},
	)
}

func NewFactory(cfg Config) route.Factory {
	if cfg.DebugPattern == "" {
		cfg.DebugPattern = ChiDefaultDebugPattern
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
	return chiRouter{rf.cfg, ctx, rf.cfg.RunServer}
}

type chiRouter struct {
	cfg       Config
	ctx       context.Context
	RunServer RunServerFunc
}

func (r chiRouter) Run(cfg config.ServiceConfig) {
	r.cfg.Engine.Use(r.cfg.Middlewares...)
	if cfg.Debug {
		r.registerDebugEndpoints()
	}

	r.cfg.Engine.Get("/__health", mux.HealthHandler)

	server.InitHTTPDefaultTransport(cfg)

	r.registerSonicEndpoints(cfg.Endpoints)

	r.cfg.Engine.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
		http.NotFound(w, r)
	})

	if err := r.RunServer(r.ctx, cfg, r.cfg.Engine); err != nil {
		r.cfg.Logger.Error(logPrefix, err.Error())
	}

	r.cfg.Logger.Info(logPrefix, "Router execution ended")
}

func (r chiRouter) registerDebugEndpoints() {
	debugHandler := mux.DebugHandler(r.cfg.Logger)
	r.cfg.Engine.Get(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Post(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Put(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Patch(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Delete(r.cfg.DebugPattern, debugHandler)
}

func (r chiRouter) registerSonicEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error(logPrefix, "calling the ProxyFactory", err.Error())
			continue
		}

		r.registerSonicEndpoint(c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r chiRouter) registerSonicEndpoint(method string, endpoint *config.EndpointConfig, handler http.HandlerFunc, totBackends int) {
	method = strings.ToTitle(method)
	path := endpoint.Endpoint

	if method != http.MethodGet && totBackends > 1 {
		if !route.IsValidSequentialEndpoint(endpoint) {
			r.cfg.Logger.Error(logPrefix, method, "endpoints with sequential proxy enabled only allow a non-GET in the last backend! Ignoring", path)
			return
		}
	}

	switch method {
	case http.MethodGet:
		r.cfg.Engine.Get(path, handler)
	case http.MethodPost:
		r.cfg.Engine.Post(path, handler)
	case http.MethodPut:
		r.cfg.Engine.Put(path, handler)
	case http.MethodPatch:
		r.cfg.Engine.Patch(path, handler)
	case http.MethodDelete:
		r.cfg.Engine.Delete(path, handler)
	default:
		r.cfg.Logger.Error(logPrefix, "Unsupported method", method)
		return
	}
	r.cfg.Logger.Debug(logPrefix, "registering the endpoint", method, path)
}
