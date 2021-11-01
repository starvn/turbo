//go:build integration || !race
// +build integration !race

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
	"fmt"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/proxy/plugin"
	"testing"
)

func TestNewPluginMiddleware(t *testing.T) {
	_, _ = plugin.LoadModifiers("./plugin/tests", ".so", plugin.RegisterModifier)

	validator := func(ctx context.Context, r *Request) (*Response, error) {
		if r.Path != "/bar/fooo/fooo" {
			return nil, fmt.Errorf("unexpected path %s", r.Path)
		}
		return nil, nil
	}

	bknd := NewBackendPluginMiddleware(
		&config.Backend{
			ExtraConfig: map[string]interface{}{
				plugin.Namespace: map[string]interface{}{
					"name": []interface{}{"turbo-request-modifier-example"},
				},
			},
		},
	)(validator)

	p := NewPluginMiddleware(
		&config.EndpointConfig{
			ExtraConfig: map[string]interface{}{
				plugin.Namespace: map[string]interface{}{
					"name": []interface{}{"turbo-request-modifier-example"},
				},
			},
		},
	)(bknd)

	resp, err := p(context.Background(), &Request{Path: "/bar"})
	if err != nil {
		t.Error(err.Error())
	}

	if resp != nil {
		t.Errorf("unexpected response: %v", resp)
	}
}
