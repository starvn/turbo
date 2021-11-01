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
	"bytes"
	"context"
	"fmt"
	"github.com/starvn/turbo/log"
	"strings"
	"testing"
)

func TestNewLoggingMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := log.NewLogger("INFO", buff, "pref")
	mw := NewLoggingMiddleware(logger, "sonic")
	mw(explosiveProxy(t), explosiveProxy(t))
}

func TestNewLoggingMiddleware_ok(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := log.NewLogger("DEBUG", buff, "pref")
	resp := &Response{IsComplete: true}
	mw := NewLoggingMiddleware(logger, "sonic")
	p := mw(dummyProxy(resp))
	r, err := p(context.Background(), &Request{})
	if r != resp {
		t.Error("The proxy didn't return the expected response")
		return
	}
	if err != nil {
		t.Errorf("The proxy returned an unexpected error: %s", err.Error())
		return
	}
	logMsg := buff.String()
	if strings.Count(logMsg, "pref") != 3 {
		t.Error("The logs don't have the injected prefix")
	}
	if strings.Count(logMsg, "INFO") != 2 {
		t.Error("The logs don't have the expected INFO messages")
	}
	if strings.Count(logMsg, "DEBU") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if !strings.Contains(logMsg, "[SONIC] Calling backend") {
		t.Error("The logs didn't mark the start of the execution")
	}
	if !strings.Contains(logMsg, "[SONIC] Call to backend took") {
		t.Error("The logs didn't mark the end of the execution")
	}
}

func TestNewLoggingMiddleware_erroredResponse(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := log.NewLogger("DEBUG", buff, "pref")
	resp := &Response{IsComplete: true}
	mw := NewLoggingMiddleware(logger, "sonic")
	expectedError := fmt.Errorf("NO-body expects the %s Inquisition!", "Spanish")
	p := mw(func(_ context.Context, _ *Request) (*Response, error) {
		return resp, expectedError
	})
	r, err := p(context.Background(), &Request{})
	if r != resp {
		t.Error("The proxy didn't return the expected response")
		return
	}
	if err != expectedError {
		t.Errorf("The proxy didn't return the expected error: %s", err.Error())
		return
	}
	logMsg := buff.String()
	if strings.Count(logMsg, "pref") != 4 {
		t.Error("The logs don't have the injected prefix")
	}
	if strings.Count(logMsg, "INFO") != 2 {
		t.Error("The logs don't have the expected INFO messages")
	}
	if strings.Count(logMsg, "DEBU") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if strings.Count(logMsg, "WARN") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if !strings.Contains(logMsg, "[SONIC] Call to backend failed: NO-body expects the Spanish Inquisition!") {
		t.Error("The logs didn't mark the fail of the execution")
	}
	if !strings.Contains(logMsg, "[SONIC] Calling backend") {
		t.Error("The logs didn't mark the start of the execution")
	}
	if !strings.Contains(logMsg, "[SONIC] Call to backend took") {
		t.Error("The logs didn't mark the end of the execution")
	}
}

func TestNewLoggingMiddleware_nullResponse(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := log.NewLogger("DEBUG", buff, "pref")
	mw := NewLoggingMiddleware(logger, "sonic")
	p := mw(dummyProxy(nil))
	r, err := p(context.Background(), &Request{})
	if r != nil {
		t.Error("The proxy didn't return the expected response")
		return
	}
	if err != nil {
		t.Errorf("The proxy returned an unexpected error: %s", err.Error())
		return
	}
	logMsg := buff.String()
	if strings.Count(logMsg, "pref") != 4 {
		t.Error("The logs don't have the injected prefix")
	}
	if strings.Count(logMsg, "INFO") != 2 {
		t.Error("The logs don't have the expected INFO messages")
	}
	if strings.Count(logMsg, "DEBUG") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if strings.Count(logMsg, "WARN") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if !strings.Contains(logMsg, "[SONIC] Call to backend returned a null response") {
		t.Error("The logs didn't mark the fail of the execution")
	}
	if !strings.Contains(logMsg, "[SONIC] Calling backend") {
		t.Error("The logs didn't mark the start of the execution")
	}
	if !strings.Contains(logMsg, "[SONIC] Call to backend took") {
		t.Error("The logs didn't mark the end of the execution")
	}
}
