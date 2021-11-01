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

package plugin

import (
	"context"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/log"
	"net/http"
)

const Namespace = "github.com/starvn/turbo/transport/http/server/plugin"
const logPrefix = "[PLUGIN: Server]"

type RunServer func(context.Context, config.ServiceConfig, http.Handler) error

func New(logger log.Logger, next RunServer) RunServer {
	return func(ctx context.Context, cfg config.ServiceConfig, handler http.Handler) error {
		v, ok := cfg.ExtraConfig[Namespace]

		if !ok {
			return next(ctx, cfg, handler)
		}
		extra, ok := v.(map[string]interface{})
		if !ok {
			logger.Debug(logPrefix, "Wrong extra_config type")
			return next(ctx, cfg, handler)
		}

		r, ok := serverRegister.Get(Namespace)
		if !ok {
			logger.Debug(logPrefix, "No plugins registered for the module")
			return next(ctx, cfg, handler)
		}

		name, nameOk := extra["name"].(string)
		fifoRaw, fifoOk := extra["name"].([]interface{})
		if !nameOk && !fifoOk {
			logger.Debug(logPrefix, "No plugins required in the extra config")
			return next(ctx, cfg, handler)
		}
		fifo := []string{}

		if !fifoOk {
			fifo = []string{name}
		} else {
			for _, x := range fifoRaw {
				if v, ok := x.(string); ok {
					fifo = append(fifo, v)
				}
			}
		}

		for _, name := range fifo {
			rawHf, ok := r.Get(name)
			if !ok {
				logger.Debug(logPrefix, "No plugin resgistered as", name)
				return next(ctx, cfg, handler)
			}

			hf, ok := rawHf.(func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error))
			if !ok {
				logger.Warning(logPrefix, "Wrong plugin handler type:", name)
				return next(ctx, cfg, handler)
			}

			handlerWrapper, err := hf(context.Background(), extra, handler)
			if err != nil {
				logger.Warning(logPrefix, "Error getting the plugin handler:", err.Error())
				return next(ctx, cfg, handler)
			}

			logger.Debug(logPrefix, "Injecting plugin", name)
			handler = handlerWrapper
		}
		return next(ctx, cfg, handler)
	}
}
