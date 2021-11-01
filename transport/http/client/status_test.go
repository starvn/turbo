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
	"github.com/starvn/turbo/config"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestDetailedHTTPStatusHandler(t *testing.T) {
	expectedErrName := "some"
	cfg := &config.Backend{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				"return_error_details": expectedErrName,
			},
		},
	}
	sh := GetHTTPStatusHandler(cfg)

	for _, code := range []int{http.StatusOK, http.StatusCreated} {
		resp := &http.Response{
			StatusCode: code,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", code, r)
			return
		}

		if err != nil {
			t.Errorf("#%d unexpected error: %s", code, err.Error())
			return
		}
	}

	for i, code := range statusCodes {
		msg := http.StatusText(code)

		resp := &http.Response{
			StatusCode: code,
			Body:       ioutil.NopCloser(bytes.NewBufferString(msg)),
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", i, r)
			return
		}

		e, ok := err.(NamedHTTPResponseError)
		if !ok {
			t.Errorf("#%d unexpected error type %T: %s", i, err, err.Error())
			return
		}

		if e.StatusCode() != code {
			t.Errorf("#%d unexpected status code: %d", i, e.Code)
			return
		}

		if e.Error() != msg {
			t.Errorf("#%d unexpected message: %s", i, e.Msg)
			return
		}

		if e.Name() != expectedErrName {
			t.Errorf("#%d unexpected error name: %s", i, e.name)
			return
		}
	}
}

func TestDefaultHTTPStatusHandler(t *testing.T) {
	sh := GetHTTPStatusHandler(&config.Backend{})

	for _, code := range []int{http.StatusOK, http.StatusCreated} {
		resp := &http.Response{
			StatusCode: code,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", code, r)
			return
		}

		if err != nil {
			t.Errorf("#%d unexpected error: %s", code, err.Error())
			return
		}
	}

	for _, code := range statusCodes {
		msg := http.StatusText(code)

		resp := &http.Response{
			StatusCode: code,
			Body:       ioutil.NopCloser(bytes.NewBufferString(msg)),
		}

		r, err := sh(context.Background(), resp)

		if r != nil {
			t.Errorf("#%d unexpected response: %v", code, r)
			return
		}

		if err != ErrInvalidStatusCode {
			t.Errorf("#%d unexpected error: %v", code, err)
			return
		}
	}
}

var statusCodes = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusPaymentRequired,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusNotAcceptable,
	http.StatusProxyAuthRequired,
	http.StatusRequestTimeout,
	http.StatusConflict,
	http.StatusGone,
	http.StatusLengthRequired,
	http.StatusPreconditionFailed,
	http.StatusRequestEntityTooLarge,
	http.StatusRequestURITooLong,
	http.StatusUnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable,
	http.StatusExpectationFailed,
	http.StatusTeapot,
	http.StatusMisdirectedRequest,
	http.StatusUnprocessableEntity,
	http.StatusLocked,
	http.StatusFailedDependency,
	http.StatusUpgradeRequired,
	http.StatusPreconditionRequired,
	http.StatusTooManyRequests,
	http.StatusRequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons,
	http.StatusInternalServerError,
	http.StatusNotImplemented,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates,
	http.StatusInsufficientStorage,
	http.StatusLoopDetected,
	http.StatusNotExtended,
	http.StatusNetworkAuthenticationRequired,
}
