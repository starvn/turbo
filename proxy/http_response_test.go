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
	"compress/gzip"
	"context"
	"github.com/starvn/turbo/encoding"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNopHTTPResponseParser(t *testing.T) {
	w := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("header1", "value1")
		_, _ = w.Write([]byte("some nice, interesting and long content"))
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	handler(w, req)
	result, err := NoOpHTTPResponseParser(context.Background(), w.Result())
	if !result.IsComplete {
		t.Error("unexpected result")
	}
	if len(result.Data) != 0 {
		t.Error("unexpected result")
	}
	if result.Metadata.StatusCode != http.StatusOK {
		t.Error("unexpected result")
	}
	headers := result.Metadata.Headers
	if h, ok := headers["Header1"]; !ok || h[0] != "value1" {
		t.Error("unexpected result:", result.Metadata.Headers)
	}
	body, err := ioutil.ReadAll(result.Io)
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
	if string(body) != "some nice, interesting and long content" {
		t.Error("unexpected result")
	}
}

func TestDefaultHTTPResponseParser_gzipped(t *testing.T) {
	w := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		gzipWriter, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer func(gzipWriter *gzip.Writer) {
			_ = gzipWriter.Close()
		}(gzipWriter)

		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = gzipWriter.Write([]byte(`{"msg":"some nice, interesting and long content"}`))
		_ = gzipWriter.Flush()
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	req.Header.Add("Accept-Encoding", "gzip")
	handler(w, req)

	result, err := DefaultHTTPResponseParserFactory(HTTPResponseParserConfig{
		Decoder:         encoding.JSONDecoder,
		EntityFormatter: DefaultHTTPResponseParserConfig.EntityFormatter,
	})(context.Background(), w.Result())

	if err != nil {
		t.Error(err)
	}

	if !result.IsComplete {
		t.Error("unexpected result")
	}
	if len(result.Data) != 1 {
		t.Error("unexpected result")
	}
	if m, ok := result.Data["msg"]; !ok || m != "some nice, interesting and long content" {
		t.Error("unexpected result")
	}
}

func TestDefaultHTTPResponseParser_plain(t *testing.T) {
	w := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"msg":"some nice, interesting and long content"}`))
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	handler(w, req)

	result, err := DefaultHTTPResponseParserFactory(HTTPResponseParserConfig{
		Decoder:         encoding.JSONDecoder,
		EntityFormatter: DefaultHTTPResponseParserConfig.EntityFormatter,
	})(context.Background(), w.Result())

	if err != nil {
		t.Error(err)
	}

	if !result.IsComplete {
		t.Error("unexpected result")
	}
	if len(result.Data) != 1 {
		t.Error("unexpected result")
	}
	if m, ok := result.Data["msg"]; !ok || m != "some nice, interesting and long content" {
		t.Error("unexpected result")
	}
}
