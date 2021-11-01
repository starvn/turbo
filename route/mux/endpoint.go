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
	"context"
	"fmt"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/core"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/transport/http/server"
	"net"
	"net/http"
	"net/textproto"
	"strings"
)

const requestParamsAsterisk string = "*"

type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

var EndpointHandler = CustomEndpointHandler(NewRequest)

func CustomEndpointHandler(rb RequestBuilder) HandlerFactory {
	return CustomEndpointHandlerWithHTTPError(rb, server.DefaultToHTTPError)
}

func CustomEndpointHandlerWithHTTPError(rb RequestBuilder, errF server.ToHTTPError) HandlerFactory {
	return func(configuration *config.EndpointConfig, prxy proxy.Proxy) http.HandlerFunc {
		cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds()))
		isCacheEnabled := configuration.CacheTTL.Seconds() != 0
		render := getRender(configuration)

		headersToSend := configuration.HeadersToPass
		if len(headersToSend) == 0 {
			headersToSend = server.HeadersToSend
		}
		method := strings.ToTitle(configuration.Method)

		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(core.SonicHeaderName, core.SonicHeaderValue)
			if r.Method != method {
				w.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
				http.Error(w, "", http.StatusMethodNotAllowed)
				return
			}

			requestCtx, cancel := context.WithTimeout(r.Context(), configuration.Timeout)

			response, err := prxy(requestCtx, rb(r, configuration.QueryString, headersToSend))

			select {
			case <-requestCtx.Done():
				if err == nil {
					err = server.ErrInternalError
				}
			default:
			}

			if response != nil && len(response.Data) > 0 {
				if response.IsComplete {
					w.Header().Set(server.CompleteResponseHeaderName, server.HeaderCompleteResponseValue)
					if isCacheEnabled {
						w.Header().Set("Cache-Control", cacheControlHeaderValue)
					}
				} else {
					w.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
				}

				for k, vs := range response.Metadata.Headers {
					for _, v := range vs {
						w.Header().Add(k, v)
					}
				}
			} else {
				w.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
				if err != nil {
					if t, ok := err.(responseError); ok {
						http.Error(w, err.Error(), t.StatusCode())
					} else {
						http.Error(w, err.Error(), errF(err))
					}
					cancel()
					return
				}
			}

			render(w, response)
			cancel()
		}
	}
}

type RequestBuilder func(r *http.Request, queryString, headersToSend []string) *proxy.Request

type ParamExtractor func(r *http.Request) map[string]string

func NoopParamExtractor(_ *http.Request) map[string]string { return map[string]string{} }

var NewRequest = NewRequestBuilder(NoopParamExtractor)

func NewRequestBuilder(paramExtractor ParamExtractor) RequestBuilder {
	return func(r *http.Request, queryString, headersToSend []string) *proxy.Request {
		params := paramExtractor(r)
		headers := make(map[string][]string, 3+len(headersToSend))

		for _, k := range headersToSend {
			if k == requestParamsAsterisk {
				headers = r.Header

				break
			}

			if h, ok := r.Header[textproto.CanonicalMIMEHeaderKey(k)]; ok {
				headers[k] = h
			}
		}

		headers["X-Forwarded-For"] = []string{clientIP(r)}
		headers["X-Forwarded-Host"] = []string{r.Host}
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = server.UserAgentHeaderValue
		} else {
			headers["X-Forwarded-Via"] = server.UserAgentHeaderValue
		}

		query := make(map[string][]string, len(queryString))
		queryValues := r.URL.Query()
		for i := range queryString {
			if queryString[i] == requestParamsAsterisk {
				query = queryValues

				break
			}

			if v, ok := queryValues[queryString[i]]; ok && len(v) > 0 {
				query[queryString[i]] = v
			}
		}

		return &proxy.Request{
			Method:  r.Method,
			Query:   query,
			Body:    r.Body,
			Params:  params,
			Headers: headers,
		}
	}
}

type responseError interface {
	error
	StatusCode() int
}

func clientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	}
	if clientIP != "" {
		return clientIP
	}

	if addr := r.Header.Get("X-Appengine-Remote-Addr"); addr != "" {
		return addr
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
