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
	"testing"
)

func TestIsValidSequentialEndpoint_ok(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "GET",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": true,
			},
		},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if !success {
		t.Error("Endpoint expected valid but receive invalid")
	}
}

func TestIsValidSequentialEndpoint_wrong_config_not_given(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "GET",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}

func TestIsValidSequentialEndpoint_wrong_config_set_false(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "GET",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": false,
			},
		}}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}

func TestIsValidSequentialEndpoint_wrong_order(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "PUT",
			},
			{
				Method: "GET",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": true,
			},
		},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}

func TestIsValidSequentialEndpoint_wrong_all_non_get(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "POST",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": true,
			},
		},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}
