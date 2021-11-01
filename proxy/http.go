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
	"context"
	"fmt"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/encoding"
	"github.com/starvn/turbo/transport/http/client"
	"net/http"
	"strconv"
	"strings"
)

var httpProxy = CustomHTTPProxyFactory(client.NewHTTPClient)

func HTTPProxyFactory(client *http.Client) BackendFactory {
	return CustomHTTPProxyFactory(func(_ context.Context) *http.Client { return client })
}

func CustomHTTPProxyFactory(cf client.HTTPClientFactory) BackendFactory {
	return func(backend *config.Backend) Proxy {
		return NewHTTPProxy(backend, cf, backend.Decoder)
	}
}

func NewHTTPProxy(remote *config.Backend, cf client.HTTPClientFactory, decode encoding.Decoder) Proxy {
	return NewHTTPProxyWithHTTPExecutor(remote, client.DefaultHTTPRequestExecutor(cf), decode)
}

func NewHTTPProxyWithHTTPExecutor(remote *config.Backend, re client.HTTPRequestExecutor, dec encoding.Decoder) Proxy {
	if remote.Encoding == encoding.NOOP {
		return NewHTTPProxyDetailed(remote, re, client.NoOpHTTPStatusHandler, NoOpHTTPResponseParser)
	}

	ef := NewEntityFormatter(remote)
	rp := DefaultHTTPResponseParserFactory(HTTPResponseParserConfig{dec, ef})
	return NewHTTPProxyDetailed(remote, re, client.GetHTTPStatusHandler(remote), rp)
}

func NewHTTPProxyDetailed(remote *config.Backend, re client.HTTPRequestExecutor, ch client.HTTPStatusHandler, rp HTTPResponseParser) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		requestToBackend, err := http.NewRequest(strings.ToTitle(request.Method), request.URL.String(), request.Body)
		if err != nil {
			return nil, err
		}
		requestToBackend.Header = make(map[string][]string, len(request.Headers))
		for k, vs := range request.Headers {
			tmp := make([]string, len(vs))
			copy(tmp, vs)
			requestToBackend.Header[k] = tmp
		}
		if request.Body != nil {
			if v, ok := request.Headers["Content-Length"]; ok && len(v) == 1 && v[0] != "chunked" {
				if size, err := strconv.Atoi(v[0]); err == nil {
					requestToBackend.ContentLength = int64(size)
				}
			}
		}

		resp, err := re(ctx, requestToBackend)
		if requestToBackend.Body != nil {
			err := requestToBackend.Body.Close()
			if err != nil {
				return nil, err
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if err != nil {
			return nil, err
		}

		resp, err = ch(ctx, resp)
		if err != nil {
			if t, ok := err.(responseError); ok {
				return &Response{
					Data: map[string]interface{}{
						fmt.Sprintf("error_%s", t.Name()): t,
					},
					Metadata: Metadata{StatusCode: t.StatusCode()},
				}, nil
			}
			return nil, err
		}

		return rp(ctx, resp)
	}
}

func NewRequestBuilderMiddleware(remote *config.Backend) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			r := request.Clone()
			r.GeneratePath(remote.URLPattern)
			r.Method = remote.Method
			return next[0](ctx, &r)
		}
	}
}

type responseError interface {
	Error() string
	Name() string
	StatusCode() int
}
