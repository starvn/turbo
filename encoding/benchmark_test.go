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

package encoding

import (
	"io"
	"strings"
	"testing"
)

func BenchmarkDecoder(b *testing.B) {
	for _, dec := range []struct {
		name    string
		decoder func(io.Reader, *map[string]interface{}) error
	}{
		{
			name:    "json-collection",
			decoder: NewJSONDecoder(true),
		},
		{
			name:    "json-map",
			decoder: NewJSONDecoder(false),
		},
		{
			name:    "safe-json-collection",
			decoder: NewSafeJSONDecoder(true),
		},
		{
			name:    "safe-json-map",
			decoder: NewSafeJSONDecoder(true),
		},
	} {
		for _, tc := range []struct {
			name  string
			input string
		}{
			{
				name:  "collection",
				input: `["a","b","c"]`,
			},
			{
				name:  "map",
				input: `{"foo": "bar", "sonic": false, "turbo": 4.20}`,
			},
		} {
			b.Run(dec.name+"/"+tc.name, func(b *testing.B) {
				var result map[string]interface{}
				for i := 0; i < b.N; i++ {
					_ = dec.decoder(strings.NewReader(tc.input), &result)
				}
			})
		}
	}
}
