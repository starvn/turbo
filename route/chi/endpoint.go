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
	"github.com/go-chi/chi/v5"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/route/mux"
	"net/http"
	"strings"
)

type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

func NewEndpointHandler(cfg *config.EndpointConfig, proxy proxy.Proxy) http.HandlerFunc {
	hf := mux.CustomEndpointHandler(
		mux.NewRequestBuilder(func(r *http.Request) map[string]string {
			return extractParamsFromEndpoint(r)
		}),
	)
	return hf(cfg, proxy)
}

func extractParamsFromEndpoint(r *http.Request) map[string]string {
	ctx := r.Context()
	rctx := chi.RouteContext(ctx)

	params := map[string]string{}
	if len(rctx.URLParams.Keys) > 0 {
		for _, param := range rctx.URLParams.Keys {
			params[strings.Title(param[:1])+param[1:]] = chi.URLParam(r, param)
		}
	}
	return params
}
