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
	"strconv"
	"testing"
)

const veryLargeString = "abcdefghijklmopqrstuvwxyzabcdefghijklmopqrstuvwxyzabcdefghijklmopqrstuvwxyzabcdefghijklmopqrstuvwxyz"

func BenchmarkNewLoadBalancedMiddleware(b *testing.B) {
	for _, tc := range []int{3, 5, 9, 13, 17, 21, 25, 50, 100} {
		b.Run(strconv.Itoa(tc), func(b *testing.B) {
			proxy := newLoadBalancedMiddleware(dummyBalancer(veryLargeString[:tc]))(dummyProxy(&Response{}))
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = proxy(context.Background(), &Request{
					Path: veryLargeString[:tc],
				})
			}
		})
	}
}

func BenchmarkNewLoadBalancedMiddleware_parallel3(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:3])
}

func BenchmarkNewLoadBalancedMiddleware_parallel5(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:5])
}

func BenchmarkNewLoadBalancedMiddleware_parallel9(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:9])
}

func BenchmarkNewLoadBalancedMiddleware_parallel13(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:13])
}

func BenchmarkNewLoadBalancedMiddleware_parallel17(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:17])
}

func BenchmarkNewLoadBalancedMiddleware_parallel21(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:21])
}

func BenchmarkNewLoadBalancedMiddleware_parallel25(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:25])
}

func BenchmarkNewLoadBalancedMiddleware_parallel50(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:50])
}

func BenchmarkNewLoadBalancedMiddleware_parallel100(b *testing.B) {
	benchmarkNewLoadBalancedMiddleware_parallel(b, veryLargeString[:100])
}

func benchmarkNewLoadBalancedMiddleware_parallel(b *testing.B, subject string) {
	b.RunParallel(func(pb *testing.PB) {
		proxy := newLoadBalancedMiddleware(dummyBalancer(subject))(dummyProxy(&Response{}))
		for pb.Next() {
			_, _ = proxy(context.Background(), &Request{
				Path: subject,
			})
		}
	})
}
