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
	"time"
)

func NewConcurrentMiddleware(remote *config.Backend) Middleware {
	if remote.ConcurrentCalls == 1 {
		panic(ErrTooManyProxies)
	}
	serviceTimeout := time.Duration(75*remote.Timeout.Nanoseconds()/100) * time.Nanosecond

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}

		return func(ctx context.Context, request *Request) (*Response, error) {
			localCtx, cancel := context.WithTimeout(ctx, serviceTimeout)

			results := make(chan *Response, remote.ConcurrentCalls)
			failed := make(chan error, remote.ConcurrentCalls)

			for i := 0; i < remote.ConcurrentCalls; i++ {
				go processConcurrentCall(localCtx, next[0], request, results, failed)
			}

			var response *Response
			var err error

			for i := 0; i < remote.ConcurrentCalls; i++ {
				select {
				case response = <-results:
					if response != nil && response.IsComplete {
						cancel()
						return response, nil
					}
				case err = <-failed:
				case <-ctx.Done():
				}
			}
			cancel()
			return response, err
		}
	}
}

var errNullResult = errors.New("invalid response")

func processConcurrentCall(ctx context.Context, next Proxy, request *Request, out chan<- *Response, failed chan<- error) {
	localCtx, cancel := context.WithCancel(ctx)

	result, err := next(localCtx, request)
	if err != nil {
		failed <- err
		cancel()
		return
	}
	if result == nil {
		failed <- errNullResult
		cancel()
		return
	}
	select {
	case out <- result:
	case <-ctx.Done():
		failed <- ctx.Err()
	}
	cancel()
}
