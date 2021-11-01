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
	"testing"
	"time"
)

func BenchmarkNewMergeDataMiddleware(b *testing.B) {
	backend := config.Backend{}
	backends := make([]*config.Backend, 10)
	for i := range backends {
		backends[i] = &backend
	}

	proxies := []Proxy{
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
	}

	for _, totalParts := range []int{2, 3, 4, 5, 6, 7, 8, 9, 10} {
		b.Run(fmt.Sprintf("with %d parts", totalParts), func(b *testing.B) {
			endpoint := config.EndpointConfig{
				Backend: backends[:totalParts],
				Timeout: time.Duration(100) * time.Millisecond,
			}
			proxy := NewMergeDataMiddleware(&endpoint)(proxies[:totalParts]...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				proxy(context.Background(), &Request{Params: map[string]string{}})
			}
		})
	}
}

func BenchmarkNewMergeDataMiddleware_sequential(b *testing.B) {
	backends := make([]*config.Backend, 10)
	pattern := "/some"
	keys := []string{}
	for i := range backends {
		b := &config.Backend{
			URLKeys:    make([]string, 4*i),
			URLPattern: pattern,
		}
		copy(b.URLKeys, keys)
		backends[i] = b
		for _, t := range []string{"int", "float", "bool", "string"} {
			next := fmt.Sprintf("Resp%d_%s", i, t)
			pattern += "/{{." + next + "}}"
			keys = append(keys, next)
		}
	}

	proxies := []Proxy{
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"int": 1, "float": 3.14, "bool": true, "string": "wwwww"}, IsComplete: true}),
	}

	for _, totalParts := range []int{2, 3, 4, 5, 6, 7, 8, 9, 10} {
		b.Run(fmt.Sprintf("with %d parts", totalParts), func(b *testing.B) {
			endpoint := config.EndpointConfig{
				Backend: backends[:totalParts],
				Timeout: time.Duration(100) * time.Millisecond,
				ExtraConfig: config.ExtraConfig{
					Namespace: map[string]interface{}{
						isSequentialKey: true,
					},
				},
			}
			proxy := NewMergeDataMiddleware(&endpoint)(proxies[:totalParts]...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = proxy(context.Background(), &Request{Params: map[string]string{}})
			}
		})
	}
}
