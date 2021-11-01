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

package mux

import (
	"encoding/json"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/encoding"
	"github.com/starvn/turbo/proxy"
	"io"
	"net/http"
	"sync"
)

type Render func(http.ResponseWriter, *proxy.Response)

const NEGOTIATE = "negotiate"

var (
	mutex          = &sync.RWMutex{}
	renderRegister = map[string]Render{
		encoding.STRING:   stringRender,
		encoding.JSON:     jsonRender,
		encoding.NOOP:     noopRender,
		"json-collection": jsonCollectionRender,
	}
)

func RegisterRender(name string, r Render) {
	mutex.Lock()
	renderRegister[name] = r
	mutex.Unlock()
}

func getRender(cfg *config.EndpointConfig) Render {
	fallback := jsonRender
	if len(cfg.Backend) == 1 {
		fallback = getWithFallback(cfg.Backend[0].Encoding, fallback)
	}

	if cfg.OutputEncoding == "" {
		return fallback
	}

	return getWithFallback(cfg.OutputEncoding, fallback)
}

func getWithFallback(key string, fallback Render) Render {
	mutex.RLock()
	r, ok := renderRegister[key]
	mutex.RUnlock()
	if !ok {
		return fallback
	}
	return r
}

var (
	emptyResponse   = []byte("{}")
	emptyCollection = []byte("[]")
)

func jsonRender(w http.ResponseWriter, response *proxy.Response) {
	w.Header().Set("Content-Type", "application/json")
	if response == nil {
		_, _ = w.Write(emptyResponse)
		return
	}

	js, err := json.Marshal(response.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(js)
}

func jsonCollectionRender(w http.ResponseWriter, response *proxy.Response) {
	w.Header().Set("Content-Type", "application/json")
	if response == nil {
		_, _ = w.Write(emptyCollection)
		return
	}
	col, ok := response.Data["collection"]
	if !ok {
		_, _ = w.Write(emptyCollection)
		return
	}

	js, err := json.Marshal(col)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(js)
}

func stringRender(w http.ResponseWriter, response *proxy.Response) {
	w.Header().Set("Content-Type", "text/plain")
	if response == nil {
		_, _ = w.Write([]byte{})
		return
	}
	d, ok := response.Data["content"]
	if !ok {
		_, _ = w.Write([]byte{})
		return
	}
	msg, ok := d.(string)
	if !ok {
		_, _ = w.Write([]byte{})
		return
	}
	_, _ = w.Write([]byte(msg))
}

func noopRender(w http.ResponseWriter, response *proxy.Response) {
	if response == nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	for k, vs := range response.Metadata.Headers {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	if response.Metadata.StatusCode != 0 {
		w.WriteHeader(response.Metadata.StatusCode)
	}

	if response.Io == nil {
		return
	}
	_, _ = io.Copy(w, response.Io)
}
