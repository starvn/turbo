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
	"github.com/starvn/turbo/transport/http/client"
	"net/http"
	"net/http/httptest"
)

const Namespace = "github.com/starvn/turbo/transport/http/client/executor"

func HTTPRequestExecutor(
	logger log.Logger,
	next func(*config.Backend) client.HTTPRequestExecutor,
) func(*config.Backend) client.HTTPRequestExecutor {
	return func(cfg *config.Backend) client.HTTPRequestExecutor {
		logPrefix := "[BACKEND: " + cfg.URLPattern + "]"
		v, ok := cfg.ExtraConfig[Namespace]
		if !ok {
			return next(cfg)
		}
		extra, ok := v.(map[string]interface{})
		if !ok {
			logger.Debug(logPrefix, "["+Namespace+"]", "Wrong extra config type for backend")
			return next(cfg)
		}

		r, ok := clientRegister.Get(Namespace)
		if !ok {
			logger.Debug(logPrefix, "No plugins registered for the module")
			return next(cfg)
		}

		name, ok := extra["name"].(string)
		if !ok {
			logger.Debug(logPrefix, "No name defined in the extra config for", cfg.URLPattern)
			return next(cfg)
		}

		rawHf, ok := r.Get(name)
		if !ok {
			logger.Debug(logPrefix, "No plugin registered as", name)
			return next(cfg)
		}

		hf, ok := rawHf.(func(context.Context, map[string]interface{}) (http.Handler, error))
		if !ok {
			logger.Warning(logPrefix, "Wrong plugin handler type:", name)
			return next(cfg)
		}

		handler, err := hf(context.Background(), extra)
		if err != nil {
			logger.Warning(logPrefix, "Error getting the plugin handler:", err.Error())
			return next(cfg)
		}

		logger.Debug(logPrefix, "Injecting plugin", name)
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req.WithContext(ctx))
			return w.Result(), nil
		}
	}
}
