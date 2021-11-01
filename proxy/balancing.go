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
	"github.com/starvn/turbo/discovery"
	"net/url"
	"strings"
)

func NewLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewLoadBalancedMiddlewareWithSubscriber(discovery.GetSubscriber(remote))
}

func NewLoadBalancedMiddlewareWithSubscriber(subscriber discovery.Subscriber) Middleware {
	return newLoadBalancedMiddleware(discovery.NewBalancer(subscriber))
}

func NewRoundRobinLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewRoundRobinLoadBalancedMiddlewareWithSubscriber(discovery.GetSubscriber(remote))
}

func NewRandomLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewRandomLoadBalancedMiddlewareWithSubscriber(discovery.GetSubscriber(remote))
}

func NewRoundRobinLoadBalancedMiddlewareWithSubscriber(subscriber discovery.Subscriber) Middleware {
	return newLoadBalancedMiddleware(discovery.NewRoundRobinLB(subscriber))
}

func NewRandomLoadBalancedMiddlewareWithSubscriber(subscriber discovery.Subscriber) Middleware {
	return newLoadBalancedMiddleware(discovery.NewRandomLB(subscriber))
}

func newLoadBalancedMiddleware(lb discovery.Balancer) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			host, err := lb.Host()
			if err != nil {
				return nil, err
			}
			r := request.Clone()

			var b strings.Builder
			b.WriteString(host)
			b.WriteString(r.Path)
			r.URL, err = url.Parse(b.String())
			if err != nil {
				return nil, err
			}
			if len(r.Query) > 0 {
				if len(r.URL.RawQuery) > 0 {
					r.URL.RawQuery += "&" + r.Query.Encode()
				} else {
					r.URL.RawQuery += r.Query.Encode()
				}
			}

			return next[0](ctx, &r)
		}
	}
}
