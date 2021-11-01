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
	"github.com/starvn/turbo/transport/http/client/graphql"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

func NewGraphQLMiddleware(remote *config.Backend) Middleware {
	opt, err := graphql.GetOptions(remote.ExtraConfig)
	if err != nil {
		return EmptyMiddleware
	}

	extractor := graphql.New(*opt)
	var generateBodyFn func(*Request) ([]byte, error)
	var generateQueryFn func(*Request) (url.Values, error)

	switch opt.Type {
	case graphql.OperationMutation:
		generateBodyFn = func(req *Request) ([]byte, error) {
			if req.Body == nil {
				return extractor.BodyFromBody(strings.NewReader(""))
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(req.Body)
			return extractor.BodyFromBody(req.Body)
		}
		generateQueryFn = func(req *Request) (url.Values, error) {
			if req.Body == nil {
				return extractor.QueryFromBody(strings.NewReader(""))
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(req.Body)
			return extractor.QueryFromBody(req.Body)
		}

	case graphql.OperationQuery:
		generateBodyFn = func(req *Request) ([]byte, error) {
			return extractor.BodyFromParams(req.Params)
		}
		generateQueryFn = func(req *Request) (url.Values, error) {
			return extractor.QueryFromParams(req.Params)
		}

	default:
		return EmptyMiddleware
	}

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}

		if opt.Method == graphql.MethodGet {
			return func(ctx context.Context, req *Request) (*Response, error) {
				q, err := generateQueryFn(req)
				if err != nil {
					return nil, err
				}

				req.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
				req.Method = string(opt.Method)
				req.Headers["Content-Length"] = []string{"0"}
				if req.Query != nil {
					for k, vs := range q {
						for _, v := range vs {
							req.Query.Add(k, v)
						}
					}
				} else {
					req.Query = q
				}

				return next[0](ctx, req)
			}
		}

		return func(ctx context.Context, req *Request) (*Response, error) {
			b, err := generateBodyFn(req)
			if err != nil {
				return nil, err
			}

			req.Body = ioutil.NopCloser(bytes.NewReader(b))
			req.Method = string(opt.Method)
			req.Headers["Content-Length"] = []string{strconv.Itoa(len(string(b)))}

			return next[0](ctx, req)
		}
	}
}
