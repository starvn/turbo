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
	"github.com/starvn/turbo/log"
	"io/ioutil"
	"net/http"
)

func DebugHandler(logger log.Logger) http.HandlerFunc {
	logPrefixSecondary := "[ENDPOINT /__debug/*]"
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug(logPrefixSecondary, "Method:", r.Method)
		logger.Debug(logPrefixSecondary, "URL:", r.RequestURI)
		logger.Debug(logPrefixSecondary, "Query:", r.URL.Query())
		logger.Debug(logPrefixSecondary, "Headers:", r.Header)
		body, _ := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		logger.Debug(logPrefixSecondary, "Body:", string(body))

		js, _ := json.Marshal(map[string]string{"message": "pong"})

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(js)
	}
}
