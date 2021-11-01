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
	"github.com/starvn/turbo/log"
	"strings"
	"time"
)

func NewLoggingMiddleware(logger log.Logger, name string) Middleware {
	logPrefix := "[" + strings.ToUpper(name) + "]"
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			begin := time.Now()
			logger.Info(logPrefix, "Calling backend")
			logger.Debug(logPrefix, "Request", request)

			result, err := next[0](ctx, request)

			logger.Info(logPrefix, "Call to backend took", time.Since(begin).String())
			if err != nil {
				logger.Warning(logPrefix, "Call to backend failed:", err.Error())
				return result, err
			}
			if result == nil {
				logger.Warning(logPrefix, "Call to backend returned a null response")
			}

			return result, err
		}
	}
}
