/*
 * Copyright (c) 2021 Huy Duc Dao
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/discovery"
	"github.com/starvn/turbo/log"
)

type Factory interface {
	New(cfg *config.EndpointConfig) (Proxy, error)
}

type FactoryFunc func(*config.EndpointConfig) (Proxy, error)

func (f FactoryFunc) New(cfg *config.EndpointConfig) (Proxy, error) { return f(cfg) }

func DefaultFactory(logger log.Logger) Factory {
	return NewDefaultFactory(httpProxy, logger)
}

func DefaultFactoryWithSubscriber(logger log.Logger, sF discovery.SubscriberFactory) Factory {
	return NewDefaultFactoryWithSubscriber(httpProxy, logger, sF)
}

func NewDefaultFactory(backendFactory BackendFactory, logger log.Logger) Factory {
	return NewDefaultFactoryWithSubscriber(backendFactory, logger, discovery.GetSubscriber)
}

func NewDefaultFactoryWithSubscriber(backendFactory BackendFactory, logger log.Logger, sF discovery.SubscriberFactory) Factory {
	return defaultFactory{backendFactory, logger, sF}
}

type defaultFactory struct {
	backendFactory    BackendFactory
	logger            log.Logger
	subscriberFactory discovery.SubscriberFactory
}

func (pf defaultFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	switch len(cfg.Backend) {
	case 0:
		err = ErrNoBackends
	case 1:
		p, err = pf.newSingle(cfg)
	default:
		p, err = pf.newMulti(cfg)
	}
	if err != nil {
		return
	}

	p = NewPluginMiddleware(cfg)(p)
	p = NewStaticMiddleware(cfg)(p)
	return
}

func (pf defaultFactory) newMulti(cfg *config.EndpointConfig) (p Proxy, err error) {
	backendProxy := make([]Proxy, len(cfg.Backend))
	for i, backend := range cfg.Backend {
		backendProxy[i] = pf.newStack(backend)
	}
	p = NewMergeDataMiddleware(cfg)(backendProxy...)
	p = NewFlatmapMiddleware(cfg)(p)
	return
}

func (pf defaultFactory) newSingle(cfg *config.EndpointConfig) (Proxy, error) {
	return pf.newStack(cfg.Backend[0]), nil
}

func (pf defaultFactory) newStack(backend *config.Backend) (p Proxy) {
	p = pf.backendFactory(backend)
	p = NewBackendPluginMiddleware(backend)(p)
	p = NewGraphQLMiddleware(backend)(p)
	p = NewLoadBalancedMiddlewareWithSubscriber(pf.subscriberFactory(backend))(p)
	if backend.ConcurrentCalls > 1 {
		p = NewConcurrentMiddleware(backend)(p)
	}
	p = NewRequestBuilderMiddleware(backend)(p)
	return
}
