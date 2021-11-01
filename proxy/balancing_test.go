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
	"errors"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/discovery/dns"
	"net"
	"net/url"
	"testing"
)

func TestNewLoadBalancedMiddleware_ok(t *testing.T) {
	want := "sonic:8080/turbo"
	lb := newLoadBalancedMiddleware(dummyBalancer("sonic:8080"))
	assertion := func(ctx context.Context, request *Request) (*Response, error) {
		if request.URL.String() != want {
			t.Errorf("The middleware did not update the request URL! want [%s], have [%s]\n", want, request.URL)
		}
		return nil, nil
	}
	if _, err := lb(assertion)(context.Background(), &Request{
		Path: "/turbo",
	}); err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
}

func TestNewLoadBalancedMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	lb := newLoadBalancedMiddleware(dummyBalancer("sonic"))
	lb(explosiveProxy(t), explosiveProxy(t))
}

func TestNewLoadBalancedMiddleware_explosiveBalancer(t *testing.T) {
	expected := errors.New("sonic")
	lb := newLoadBalancedMiddleware(explosiveBalancer{expected})
	if _, err := lb(explosiveProxy(t))(context.Background(), &Request{}); err != expected {
		t.Errorf("The middleware did not propagate the lb error\n")
	}
}

func TestNewRoundRobinLoadBalancedMiddleware(t *testing.T) {
	testLoadBalancedMw(t, NewRoundRobinLoadBalancedMiddleware(&config.Backend{
		Host: []string{"http://127.0.0.1:8080"},
	}))
}

func TestNewRandomLoadBalancedMiddleware(t *testing.T) {
	testLoadBalancedMw(t, NewRandomLoadBalancedMiddleware(&config.Backend{
		Host: []string{"http://127.0.0.1:8080"},
	}))
}

func testLoadBalancedMw(t *testing.T, lb Middleware) {
	for _, tc := range []struct {
		path     string
		query    url.Values
		expected string
	}{
		{
			path:     "/turbo",
			expected: "http://127.0.0.1:8080/turbo",
		},
		{
			path:     "/turbo?extra=true",
			expected: "http://127.0.0.1:8080/turbo?extra=true",
		},
		{
			path:     "/turbo?extra=true",
			query:    url.Values{"some": []string{"none"}},
			expected: "http://127.0.0.1:8080/turbo?extra=true&some=none",
		},
		{
			path:     "/turbo",
			query:    url.Values{"some": []string{"none"}},
			expected: "http://127.0.0.1:8080/turbo?some=none",
		},
	} {
		assertion := func(ctx context.Context, request *Request) (*Response, error) {
			if request.URL.String() != tc.expected {
				t.Errorf("The middleware did not update the request URL! want [%s], have [%s]\n", tc.expected, request.URL)
			}
			return nil, nil
		}
		if _, err := lb(assertion)(context.Background(), &Request{
			Path:  tc.path,
			Query: tc.query,
		}); err != nil {
			t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
		}
	}
}

func TestNewLoadBalancedMiddleware_parsingError(t *testing.T) {
	lb := NewRandomLoadBalancedMiddleware(&config.Backend{
		Host: []string{"127.0.0.1:8080"},
	})
	assertion := func(ctx context.Context, request *Request) (*Response, error) {
		t.Error("The middleware didn't block the request!")
		return nil, nil
	}
	if _, err := lb(assertion)(context.Background(), &Request{
		Path: "/turbo",
	}); err == nil {
		t.Error("The middleware didn't propagate the expected error")
	}
}

func TestNewRoundRobinLoadBalancedMiddleware_DNSSRV(t *testing.T) {
	defaultLookup := dns.DefaultLookup

	dns.DefaultLookup = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", []*net.SRV{
			{
				Port:   8080,
				Target: "127.0.0.1",
			},
		}, nil
	}
	testLoadBalancedMw(t, NewRoundRobinLoadBalancedMiddlewareWithSubscriber(dns.New("some.service.example.tld")))

	dns.DefaultLookup = defaultLookup
}

type dummyBalancer string

func (d dummyBalancer) Host() (string, error) { return string(d), nil }

type explosiveBalancer struct {
	Error error
}

func (e explosiveBalancer) Host() (string, error) { return "", e.Error }
