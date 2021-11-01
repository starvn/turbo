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
)

const (
	shadowKey = "shadow"
)

type shadowFactory struct {
	f Factory
}

func (s shadowFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	if len(cfg.Backend) == 0 {
		err = ErrNoBackends
		return
	}

	var shadow []*config.Backend
	var regular []*config.Backend

	for _, b := range cfg.Backend {
		if isShadowBackend(b) {
			shadow = append(shadow, b)
			continue
		}
		regular = append(regular, b)
	}

	cfg.Backend = regular

	p, err = s.f.New(cfg)

	if len(shadow) > 0 {
		cfg.Backend = shadow
		pShadow, _ := s.f.New(cfg)
		p = ShadowMiddleware(p, pShadow)
	}

	return
}

func NewShadowFactory(f Factory) Factory {
	return shadowFactory{f}
}

func ShadowMiddleware(next ...Proxy) Proxy {
	switch len(next) {
	case 0:
		panic(ErrNotEnoughProxies)
	case 1:
		return next[0]
	case 2:
		return NewShadowProxy(next[0], next[1])
	default:
		panic(ErrTooManyProxies)
	}
}

func NewShadowProxy(p1, p2 Proxy) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		go func() {
			_, _ = p2(newContextWrapper(ctx), CloneRequest(request))
		}()
		return p1(ctx, request)
	}
}

func isShadowBackend(c *config.Backend) bool {
	if v, ok := c.ExtraConfig[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if v, ok := e[shadowKey]; ok {
				c, ok := v.(bool)
				return ok && c
			}
		}
	}
	return false
}

type contextWrapper struct {
	context.Context
	data context.Context
}

func (c contextWrapper) Value(key interface{}) interface{} {
	return c.data.Value(key)
}

func newContextWrapper(data context.Context) contextWrapper {
	return contextWrapper{
		Context: context.Background(),
		data:    data,
	}
}
