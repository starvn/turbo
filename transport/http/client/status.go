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

package client

import (
	"bytes"
	"context"
	"errors"
	"github.com/starvn/turbo/config"
	"io/ioutil"
	"net/http"
)

const Namespace = "github.com/starvn/turbo/transport/http/client"

var ErrInvalidStatusCode = errors.New("Invalid status code")

type HTTPStatusHandler func(context.Context, *http.Response) (*http.Response, error)

func GetHTTPStatusHandler(remote *config.Backend) HTTPStatusHandler {
	if e, ok := remote.ExtraConfig[Namespace]; ok {
		if m, ok := e.(map[string]interface{}); ok {
			if v, ok := m["return_error_details"]; ok {
				if b, ok := v.(string); ok && b != "" {
					return DetailedHTTPStatusHandler(b)
				}
			} else if v, ok := m["return_error_code"].(bool); ok && v {
				return ErrorHTTPStatusHandler
			}
		}
	}
	return DefaultHTTPStatusHandler
}

func DefaultHTTPStatusHandler(ctx context.Context, resp *http.Response) (*http.Response, error) {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, ErrInvalidStatusCode
	}

	return resp, nil
}

func ErrorHTTPStatusHandler(ctx context.Context, resp *http.Response) (*http.Response, error) {
	if _, err := DefaultHTTPStatusHandler(ctx, resp); err == nil {
		return resp, nil
	}
	return resp, newHTTPResponseError(resp)
}

func NoOpHTTPStatusHandler(_ context.Context, resp *http.Response) (*http.Response, error) {
	return resp, nil
}

func DetailedHTTPStatusHandler(name string) HTTPStatusHandler {
	return func(ctx context.Context, resp *http.Response) (*http.Response, error) {
		if _, err := DefaultHTTPStatusHandler(ctx, resp); err == nil {
			return resp, nil
		}

		return resp, NamedHTTPResponseError{
			HTTPResponseError: newHTTPResponseError(resp),
			name:              name,
		}
	}
}

func newHTTPResponseError(resp *http.Response) HTTPResponseError {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = []byte{}
	}
	_ = resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return HTTPResponseError{
		Code: resp.StatusCode,
		Msg:  string(body),
	}
}

type HTTPResponseError struct {
	Code int    `json:"http_status_code"`
	Msg  string `json:"http_body,omitempty"`
}

func (r HTTPResponseError) Error() string {
	return r.Msg
}

func (r HTTPResponseError) StatusCode() int {
	return r.Code
}

type NamedHTTPResponseError struct {
	HTTPResponseError
	name string
}

func (r NamedHTTPResponseError) Name() string {
	return r.name
}
