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

// Package proxy provides proxy and proxy middleware interfaces and implementations
package proxy

import (
	"context"
	"errors"
	"github.com/starvn/turbo/config"
	"io"
)

const Namespace = "turbo/proxy"

type Metadata struct {
	Headers    map[string][]string
	StatusCode int
}

type Response struct {
	Data       map[string]interface{}
	IsComplete bool
	Metadata   Metadata
	Io         io.Reader
}

type readCloserWrapper struct {
	ctx context.Context
	rc  io.ReadCloser
}

func NewReadCloserWrapper(ctx context.Context, in io.ReadCloser) io.Reader {
	wrapper := readCloserWrapper{ctx, in}
	go wrapper.closeOnCancel()
	return wrapper
}

func (w readCloserWrapper) Read(b []byte) (int, error) {
	return w.rc.Read(b)
}

func (w readCloserWrapper) closeOnCancel() {
	<-w.ctx.Done()
	err := w.rc.Close()
	if err != nil {
		return
	}
}

var (
	ErrNoBackends       = errors.New("all endpoints must have at least one backend")
	ErrTooManyBackends  = errors.New("too many backends for this proxy")
	ErrTooManyProxies   = errors.New("too many proxies for this proxy middleware")
	ErrNotEnoughProxies = errors.New("not enough proxies for this endpoint")
)

type Proxy func(ctx context.Context, request *Request) (*Response, error)

type BackendFactory func(remote *config.Backend) Proxy

type Middleware func(next ...Proxy) Proxy

func EmptyMiddleware(next ...Proxy) Proxy {
	if len(next) > 1 {
		panic(ErrTooManyProxies)
	}
	return next[0]
}

func NoopProxy(_ context.Context, _ *Request) (*Response, error) { return nil, nil }
