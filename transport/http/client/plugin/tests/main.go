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

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/starvn/turbo/log"
	"html"
	"net/http"
)

var ClientRegisterer = registerer("sonic-client-example")

type registerer string

var logger log.Logger = nil

func (r registerer) RegisterLogger(v interface{}) {
	l, ok := v.(log.Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(ClientRegisterer, "client plugin loaded!!!")
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	if logger == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, _ = fmt.Fprintf(w, "Hello, %q", html.EscapeString(req.URL.Path))
		}), nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello, %q", html.EscapeString(req.URL.Path))
		logger.Debug("request:", html.EscapeString(req.URL.Path))
	}), nil
}

func init() {
	fmt.Println(ClientRegisterer, "client plugin loaded!!!")
}

func main() {}
