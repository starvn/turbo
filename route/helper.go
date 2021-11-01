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

package route

import (
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/proxy"
	"net/http"
)

func IsValidSequentialEndpoint(endpoint *config.EndpointConfig) bool {
	if endpoint.ExtraConfig[proxy.Namespace] == nil {
		return false
	}

	proxyCfg := endpoint.ExtraConfig[proxy.Namespace].(map[string]interface{})
	if proxyCfg["sequential"] == false {
		return false
	}

	for i, backend := range endpoint.Backend {
		if backend.Method != http.MethodGet && (i+1) != len(endpoint.Backend) {
			return false
		}
	}

	return true
}
