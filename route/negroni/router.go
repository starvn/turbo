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

// Package negroni provides some basic implementations for building routes based on urfave/negroni
package negroni

import (
	gorilla "github.com/gorilla/mux"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/route"
	turbogorilla "github.com/starvn/turbo/route/gorilla"
	"github.com/starvn/turbo/route/mux"
	"github.com/urfave/negroni/v2"
	"net/http"
)

func DefaultFactory(pf proxy.Factory, logger log.Logger, middlewares []negroni.Handler) route.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger, middlewares))
}

func DefaultConfig(pf proxy.Factory, logger log.Logger, middlewares []negroni.Handler) mux.Config {
	return DefaultConfigWithRouter(pf, logger, NewGorillaRouter(), middlewares)
}

func DefaultConfigWithRouter(pf proxy.Factory, logger log.Logger, muxEngine *gorilla.Router, middlewares []negroni.Handler) mux.Config {
	cfg := turbogorilla.DefaultConfig(pf, logger)
	cfg.Engine = newNegroniEngine(muxEngine, middlewares...)
	return cfg
}

func NewGorillaRouter() *gorilla.Router {
	return gorilla.NewRouter()
}

func newNegroniEngine(muxEngine *gorilla.Router, middlewares ...negroni.Handler) negroniEngine {
	negroniRouter := negroni.Classic()
	for _, m := range middlewares {
		negroniRouter.Use(m)
	}

	negroniRouter.UseHandler(muxEngine)

	return negroniEngine{muxEngine, negroniRouter}
}

type negroniEngine struct {
	r *gorilla.Router
	n *negroni.Negroni
}

func (e negroniEngine) Handle(pattern, method string, handler http.Handler) {
	e.r.Handle(pattern, handler).Methods(method)
}

func (e negroniEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.n.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}
