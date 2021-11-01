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

package discovery

import (
	"fmt"
	"testing"
)

var balancerTestsCases = [][]string{
	{"a"},
	{"a", "b", "c"},
	{"a", "b", "c", "e", "f"},
}

func BenchmarkLB(b *testing.B) {
	for _, tc := range []struct {
		name string
		f    func([]string) Balancer
	}{
		{name: "round_robin", f: func(hs []string) Balancer { return NewRoundRobinLB(FixedSubscriber(hs)) }},
		{name: "random", f: func(hs []string) Balancer { return NewRandomLB(FixedSubscriber(hs)) }},
	} {
		for _, testCase := range balancerTestsCases {
			b.Run(fmt.Sprintf("%s/%d", tc.name, len(testCase)), func(b *testing.B) {
				balancer := tc.f(testCase)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _ = balancer.Host()
				}
			})
		}
	}
}

func BenchmarkLB_parallel(b *testing.B) {
	for _, tc := range []struct {
		name string
		f    func([]string) Balancer
	}{
		{name: "round_robin", f: func(hs []string) Balancer { return NewRoundRobinLB(FixedSubscriber(hs)) }},
		{name: "random", f: func(hs []string) Balancer { return NewRandomLB(FixedSubscriber(hs)) }},
	} {
		for _, testCase := range balancerTestsCases {
			b.Run(fmt.Sprintf("%s/%d", tc.name, len(testCase)), func(b *testing.B) {
				balancer := tc.f(testCase)
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_, _ = balancer.Host()
					}
				})
			})
		}
	}
}
