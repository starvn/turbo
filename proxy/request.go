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
	"io"
	"io/ioutil"
	"net/url"
)

type Request struct {
	Method  string
	URL     *url.URL
	Query   url.Values
	Path    string
	Body    io.ReadCloser
	Params  map[string]string
	Headers map[string][]string
}

func (r *Request) GeneratePath(URLPattern string) {
	if len(r.Params) == 0 {
		r.Path = URLPattern
		return
	}
	buff := []byte(URLPattern)
	for k, v := range r.Params {
		var key []byte
		key = append(key, "{{."...)
		key = append(key, k...)
		key = append(key, "}}"...)
		buff = bytes.Replace(buff, key, []byte(v), -1)
	}
	r.Path = string(buff)
}

func (r *Request) Clone() Request {
	return Request{
		Method:  r.Method,
		URL:     r.URL,
		Query:   r.Query,
		Path:    r.Path,
		Body:    r.Body,
		Params:  r.Params,
		Headers: r.Headers,
	}
}

func CloneRequest(r *Request) *Request {
	clone := r.Clone()
	clone.Headers = CloneRequestHeaders(r.Headers)
	clone.Params = CloneRequestParams(r.Params)
	if r.Body == nil {
		return &clone
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	_ = r.Body.Close()

	r.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	clone.Body = ioutil.NopCloser(buf)

	return &clone
}

func CloneRequestHeaders(headers map[string][]string) map[string][]string {
	m := make(map[string][]string, len(headers))
	for k, vs := range headers {
		tmp := make([]string, len(vs))
		copy(tmp, vs)
		m[k] = tmp
	}
	return m
}

func CloneRequestParams(params map[string]string) map[string]string {
	m := make(map[string]string, len(params))
	for k, v := range params {
		m[k] = v
	}
	return m
}
