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
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/proxy/plugin"
	"io"
	"net/url"
)

func NewPluginMiddleware(endpoint *config.EndpointConfig) Middleware {
	cfg, ok := endpoint.ExtraConfig[plugin.Namespace].(map[string]interface{})

	if !ok {
		return EmptyMiddleware
	}

	return newPluginMiddleware(cfg)
}

func NewBackendPluginMiddleware(remote *config.Backend) Middleware {
	cfg, ok := remote.ExtraConfig[plugin.Namespace].(map[string]interface{})

	if !ok {
		return EmptyMiddleware
	}

	return newPluginMiddleware(cfg)
}

func newPluginMiddleware(cfg map[string]interface{}) Middleware {
	plugins, ok := cfg["name"].([]interface{})
	if !ok {
		return EmptyMiddleware
	}

	var reqModifiers []func(interface{}) (interface{}, error)
	var respModifiers []func(interface{}) (interface{}, error)

	for _, p := range plugins {
		name, ok := p.(string)
		if !ok {
			continue
		}

		if mf, ok := plugin.GetRequestModifier(name); ok {
			reqModifiers = append(reqModifiers, mf(cfg))
			continue
		}

		if mf, ok := plugin.GetResponseModifier(name); ok {
			respModifiers = append(respModifiers, mf(cfg))
		}
	}

	totReqModifiers, totRespModifiers := len(reqModifiers), len(respModifiers)
	if totReqModifiers == totRespModifiers && totRespModifiers == 0 {
		return EmptyMiddleware
	}

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}

		if totReqModifiers == 0 {
			return func(ctx context.Context, r *Request) (*Response, error) {
				resp, err := next[0](ctx, r)
				if err != nil {
					return resp, err
				}

				return executeResponseModifiers(respModifiers, resp)
			}
		}

		if totRespModifiers == 0 {
			return func(ctx context.Context, r *Request) (*Response, error) {
				var err error
				r, err = executeRequestModifiers(reqModifiers, r)
				if err != nil {
					return nil, err
				}

				return next[0](ctx, r)
			}
		}

		return func(ctx context.Context, r *Request) (*Response, error) {
			var err error
			r, err = executeRequestModifiers(reqModifiers, r)
			if err != nil {
				return nil, err
			}

			resp, err := next[0](ctx, r)
			if err != nil {
				return resp, err
			}

			return executeResponseModifiers(respModifiers, resp)
		}
	}
}

func executeRequestModifiers(reqModifiers []func(interface{}) (interface{}, error), r *Request) (*Request, error) {
	var tmp RequestWrapper
	tmp = requestWrapper{
		method:  r.Method,
		url:     r.URL,
		query:   r.Query,
		path:    r.Path,
		body:    r.Body,
		params:  r.Params,
		headers: r.Headers,
	}

	for _, f := range reqModifiers {
		res, err := f(tmp)
		if err != nil {
			return nil, err
		}
		t, ok := res.(RequestWrapper)
		if !ok {
			continue
		}
		tmp = t
	}

	r.Method = tmp.Method()
	r.URL = tmp.URL()
	r.Query = tmp.Query()
	r.Path = tmp.Path()
	r.Body = tmp.Body()
	r.Params = tmp.Params()
	r.Headers = tmp.Headers()

	return r, nil
}

func executeResponseModifiers(respModifiers []func(interface{}) (interface{}, error), r *Response) (*Response, error) {
	var tmp ResponseWrapper
	tmp = responseWrapper{
		data:       r.Data,
		isComplete: r.IsComplete,
		metadata: metadataWrapper{
			headers:    r.Metadata.Headers,
			statusCode: r.Metadata.StatusCode,
		},
		io: r.Io,
	}

	for _, f := range respModifiers {
		res, err := f(tmp)
		if err != nil {
			return nil, err
		}
		t, ok := res.(ResponseWrapper)
		if !ok {
			continue
		}
		tmp = t
	}

	r.Data = tmp.Data()
	r.IsComplete = tmp.IsComplete()
	r.Io = tmp.Io()
	r.Metadata = Metadata{}
	if m := tmp.Metadata(); m != nil {
		r.Metadata.Headers = m.Headers()
		r.Metadata.StatusCode = m.StatusCode()
	}
	return r, nil
}

type RequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
}

type ResponseMetadataWrapper interface {
	Headers() map[string][]string
	StatusCode() int
}

type ResponseWrapper interface {
	Data() map[string]interface{}
	Io() io.Reader
	IsComplete() bool
	Metadata() ResponseMetadataWrapper
}

type requestWrapper struct {
	method  string
	url     *url.URL
	query   url.Values
	path    string
	body    io.ReadCloser
	params  map[string]string
	headers map[string][]string
}

func (r requestWrapper) Method() string               { return r.method }
func (r requestWrapper) URL() *url.URL                { return r.url }
func (r requestWrapper) Query() url.Values            { return r.query }
func (r requestWrapper) Path() string                 { return r.path }
func (r requestWrapper) Body() io.ReadCloser          { return r.body }
func (r requestWrapper) Params() map[string]string    { return r.params }
func (r requestWrapper) Headers() map[string][]string { return r.headers }

type metadataWrapper struct {
	headers    map[string][]string
	statusCode int
}

func (m metadataWrapper) Headers() map[string][]string { return m.headers }
func (m metadataWrapper) StatusCode() int              { return m.statusCode }

type responseWrapper struct {
	data       map[string]interface{}
	isComplete bool
	metadata   metadataWrapper
	io         io.Reader
}

func (r responseWrapper) Data() map[string]interface{}      { return r.data }
func (r responseWrapper) IsComplete() bool                  { return r.isComplete }
func (r responseWrapper) Metadata() ResponseMetadataWrapper { return r.metadata }
func (r responseWrapper) Io() io.Reader                     { return r.io }
