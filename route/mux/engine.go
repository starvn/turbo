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

package mux

import (
	"github.com/starvn/turbo/transport/http/server"
	"net/http"
	"sync"
)

type Engine interface {
	http.Handler
	Handle(pattern, method string, handler http.Handler)
}

type engine struct {
	handler *http.ServeMux
	dict    map[string]map[string]http.HandlerFunc
}

func NewHTTPErrorInterceptor(w http.ResponseWriter) *HTTPErrorInterceptor {
	return &HTTPErrorInterceptor{w, new(sync.Once)}
}

type HTTPErrorInterceptor struct {
	http.ResponseWriter
	once *sync.Once
}

func (i *HTTPErrorInterceptor) WriteHeader(code int) {
	i.once.Do(func() {
		if code != http.StatusOK {
			i.ResponseWriter.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
		}
	})
	i.ResponseWriter.WriteHeader(code)
}

func DefaultEngine() *engine {
	return &engine{
		handler: http.NewServeMux(),
		dict:    map[string]map[string]http.HandlerFunc{},
	}
}

func (e *engine) Handle(pattern, method string, handler http.Handler) {
	if _, ok := e.dict[pattern]; !ok {
		e.dict[pattern] = map[string]http.HandlerFunc{}
		e.handler.Handle(pattern, e.registrableHandler(pattern))
	}
	e.dict[pattern][method] = handler.ServeHTTP
}

func (e *engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(NewHTTPErrorInterceptor(w), r)
}

func (e *engine) registrableHandler(pattern string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if handler, ok := e.dict[pattern][req.Method]; ok {
			handler(rw, req)
			return
		}

		rw.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
		http.Error(rw, "", http.StatusMethodNotAllowed)
	})
}
