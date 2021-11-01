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

// Package gin provides some basic implementations for building routers based on gin-gonic/gin
package gin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/core"
	"github.com/starvn/turbo/log"
	"github.com/starvn/turbo/proxy"
	"github.com/starvn/turbo/route"
	"github.com/starvn/turbo/transport/http/server"
	"net/http"
	"sort"
	"strings"
	"sync"
)

const logPrefix = "[SERVICE: Gin]"

type RunServerFunc func(context.Context, config.ServiceConfig, http.Handler) error

type Config struct {
	Engine         *gin.Engine
	Middlewares    []gin.HandlerFunc
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         log.Logger
	RunServer      RunServerFunc
}

func DefaultFactory(proxyFactory proxy.Factory, logger log.Logger) route.Factory {
	return NewFactory(
		Config{
			Engine:         gin.Default(),
			Middlewares:    []gin.HandlerFunc{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   proxyFactory,
			Logger:         logger,
			RunServer:      server.RunServer,
		},
	)
}

func NewFactory(cfg Config) route.Factory {
	return factory{cfg}
}

type factory struct {
	cfg Config
}

func (rf factory) New() route.Router {
	return rf.NewWithContext(context.Background())
}

func (rf factory) NewWithContext(ctx context.Context) route.Router {
	return ginRouter{
		cfg:        rf.cfg,
		ctx:        ctx,
		runServerF: rf.cfg.RunServer,
		mu:         new(sync.Mutex),
		urlCatalog: map[string][]string{},
	}
}

type ginRouter struct {
	cfg        Config
	ctx        context.Context
	runServerF RunServerFunc
	mu         *sync.Mutex
	urlCatalog map[string][]string
}

func (r ginRouter) Run(cfg config.ServiceConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	server.InitHTTPDefaultTransport(cfg)

	r.registerEndpointsAndMiddlewares(cfg)

	// TODO: remove this ugly hack once the https://github.com/gin-gonic/gin/pull/2692 and
	// https://github.com/gin-gonic/gin/issues/2862 are completely fixed
	go r.cfg.Engine.Run("XXXX")

	if err := r.runServerF(r.ctx, cfg, r.cfg.Engine); err != nil && err != http.ErrServerClosed {
		r.cfg.Logger.Error(logPrefix, err.Error())
	}

	r.cfg.Logger.Info(logPrefix, "Router execution ended")
}

func (r ginRouter) registerEndpointsAndMiddlewares(cfg config.ServiceConfig) {
	if cfg.Debug {
		r.cfg.Engine.Any("/__debug/*param", DebugHandler(r.cfg.Logger))
	}

	r.cfg.Engine.GET("/__health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	endpointGroup := r.cfg.Engine.Group("/")
	endpointGroup.Use(r.cfg.Middlewares...)

	r.registerSonicEndpoints(endpointGroup, cfg.Endpoints)

	if opts, ok := cfg.ExtraConfig[Namespace].(map[string]interface{}); ok {
		if v, ok := opts["auto_options"].(bool); ok && v {
			r.cfg.Logger.Debug(logPrefix, "Enabling the auto options endpoints")
			r.registerOptionEndpoints(endpointGroup)
		}
	}
}

func (r ginRouter) registerSonicEndpoints(rg *gin.RouterGroup, endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error(logPrefix, "Calling the ProxyFactory", err.Error())
			continue
		}

		r.registerSonicEndpoint(rg, c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r ginRouter) registerSonicEndpoint(rg *gin.RouterGroup, method string, e *config.EndpointConfig, h gin.HandlerFunc, total int) {
	method = strings.ToTitle(method)
	path := e.Endpoint
	if method != http.MethodGet && total > 1 {
		if !route.IsValidSequentialEndpoint(e) {
			r.cfg.Logger.Error(logPrefix, method, "endpoints with sequential proxy enabled only allow a non-GET in the last backend! Ignoring", path)
			return
		}
	}

	switch method {
	case http.MethodGet:
		rg.GET(path, h)
	case http.MethodPost:
		rg.POST(path, h)
	case http.MethodPut:
		rg.PUT(path, h)
	case http.MethodPatch:
		rg.PATCH(path, h)
	case http.MethodDelete:
		rg.DELETE(path, h)
	default:
		r.cfg.Logger.Error(logPrefix, "Unsupported method", method)
	}

	methods, ok := r.urlCatalog[path]
	if !ok {
		r.urlCatalog[path] = []string{method}
		return
	}
	r.urlCatalog[path] = append(methods, method)
}

func (r ginRouter) registerOptionEndpoints(rg *gin.RouterGroup) {
	for path, methods := range r.urlCatalog {
		sort.Strings(methods)
		allowed := strings.Join(methods, ", ")

		rg.OPTIONS(path, func(c *gin.Context) {
			c.Header("Allow", allowed)
			c.Header(core.SonicHeaderName, core.SonicHeaderValue)
		})
	}
}
