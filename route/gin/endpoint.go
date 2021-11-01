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
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/core"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/transport/http/server"
	"net/textproto"
	"strings"
)

const requestParamsAsterisk string = "*"

type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) gin.HandlerFunc

var EndpointHandler = CustomErrorEndpointHandler(log.NoOp, server.DefaultToHTTPError)

func CustomErrorEndpointHandler(logger log.Logger, errF server.ToHTTPError) HandlerFactory {
	return func(configuration *config.EndpointConfig, prxy proxy.Proxy) gin.HandlerFunc {
		cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds()))
		isCacheEnabled := configuration.CacheTTL.Seconds() != 0
		requestGenerator := NewRequest(configuration.HeadersToPass)
		render := getRender(configuration)
		logPrefix := "[ENDPOINT: " + configuration.Endpoint + "]"

		return func(c *gin.Context) {
			requestCtx, cancel := context.WithTimeout(c, configuration.Timeout)

			c.Header(core.SonicHeaderName, core.SonicHeaderValue)

			response, err := prxy(requestCtx, requestGenerator(c, configuration.QueryString))

			select {
			case <-requestCtx.Done():
				if err == nil {
					err = server.ErrInternalError
				}
			default:
			}

			complete := server.HeaderIncompleteResponseValue

			if response != nil && len(response.Data) > 0 {
				if response.IsComplete {
					complete = server.HeaderCompleteResponseValue
					if isCacheEnabled {
						c.Header("Cache-Control", cacheControlHeaderValue)
					}
				}

				for k, vs := range response.Metadata.Headers {
					for _, v := range vs {
						c.Writer.Header().Add(k, v)
					}
				}
			}

			c.Header(server.CompleteResponseHeaderName, complete)

			if err != nil {
				if t, ok := err.(multiError); ok {
					for i, errN := range t.Errors() {
						logger.Error(fmt.Sprintf("%s Error #%d: %s", logPrefix, i, errN.Error()))
					}
				} else {
					logger.Error(logPrefix, err.Error())
				}

				if response == nil {
					if t, ok := err.(responseError); ok {
						c.Status(t.StatusCode())
					} else {
						c.Status(errF(err))
					}
					if returnErrorMsg {
						_, _ = c.Writer.WriteString(err.Error())
					}
					cancel()
					return
				}
			}

			render(c, response)
			cancel()
		}
	}
}

func NewRequest(headersToSend []string) func(*gin.Context, []string) *proxy.Request {
	if len(headersToSend) == 0 {
		headersToSend = server.HeadersToSend
	}

	return func(c *gin.Context, queryString []string) *proxy.Request {
		params := make(map[string]string, len(c.Params))
		for _, param := range c.Params {
			params[strings.Title(param.Key[:1])+param.Key[1:]] = param.Value
		}

		headers := make(map[string][]string, 3+len(headersToSend))

		for _, k := range headersToSend {
			if k == requestParamsAsterisk {
				headers = c.Request.Header

				break
			}

			if h, ok := c.Request.Header[textproto.CanonicalMIMEHeaderKey(k)]; ok {
				headers[k] = h
			}
		}

		headers["X-Forwarded-For"] = []string{c.ClientIP()}
		headers["X-Forwarded-Host"] = []string{c.Request.Host}
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = server.UserAgentHeaderValue
		} else {
			headers["X-Forwarded-Via"] = server.UserAgentHeaderValue
		}

		query := make(map[string][]string, len(queryString))
		queryValues := c.Request.URL.Query()
		for i := range queryString {
			if queryString[i] == requestParamsAsterisk {
				query = c.Request.URL.Query()

				break
			}

			if v, ok := queryValues[queryString[i]]; ok && len(v) > 0 {
				query[queryString[i]] = v
			}
		}

		return &proxy.Request{
			Method:  c.Request.Method,
			Query:   query,
			Body:    c.Request.Body,
			Params:  params,
			Headers: headers,
		}
	}
}

type responseError interface {
	error
	StatusCode() int
}

type multiError interface {
	error
	Errors() []error
}
