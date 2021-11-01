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
	"errors"
	"fmt"
	"github.com/starvn/turbo/config"
	"math"
	"testing"
)

func TestRoundRobinLB(t *testing.T) {
	for _, endpoints := range balancerTestsCases {
		t.Run(fmt.Sprintf("%d hosts", len(endpoints)), func(t *testing.T) {
			var (
				n          = len(endpoints)
				counts     = make(map[string]int, n)
				iterations = 100000 * n
				want       = iterations / n
			)

			for _, e := range endpoints {
				counts[e] = 0
			}

			subscriber := FixedSubscriber(endpoints)
			balancer := NewRoundRobinLB(subscriber)

			if b, ok := balancer.(*roundRobinLB); ok {
				b.counter = 0
			}

			for i := 0; i < iterations; i++ {
				endpoint, err := balancer.Host()
				if err != nil {
					t.Fail()
				}
				expected := i % n
				if v := endpoints[expected]; v != endpoint {
					t.Errorf("%d: want %s, have %s", i, endpoints[expected], endpoint)
				}
				counts[endpoint]++
			}

			for i, have := range counts {
				if have != want {
					t.Errorf("%s: want %d, have %d", i, want, have)
				}
			}
		})
	}
}

func TestRoundRobinLB_noEndpoints(t *testing.T) {
	subscriber := FixedSubscriber{}
	balancer := NewRoundRobinLB(subscriber)
	_, err := balancer.Host()
	if want, have := ErrNoHosts, err; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

func TestRandomLB(t *testing.T) {
	var (
		endpoints  = []string{"a", "b", "c", "d", "e", "f", "g"}
		n          = len(endpoints)
		counts     = make(map[string]int, n)
		iterations = 1000000
		want       = iterations / n
		tolerance  = want / 100 // 1%
	)

	for _, e := range endpoints {
		counts[e] = 0
	}

	subscriber := FixedSubscriber(endpoints)
	balancer := NewRandomLB(subscriber)

	for i := 0; i < iterations; i++ {
		endpoint, err := balancer.Host()
		if err != nil {
			t.Fail()
		}
		counts[endpoint]++
	}

	for i, have := range counts {
		delta := int(math.Abs(float64(want - have)))
		if delta > tolerance {
			t.Errorf("%s: want %d, have %d, delta %d > %d tolerance", i, want, have, delta, tolerance)
		}
	}
}

func TestRandomLB_single(t *testing.T) {
	endpoints := []string{"a"}
	iterations := 1000000
	subscriber := FixedSubscriber(endpoints)
	balancer := NewRandomLB(subscriber)

	for i := 0; i < iterations; i++ {
		endpoint, err := balancer.Host()
		if err != nil {
			t.Fail()
		}
		if endpoint != endpoints[0] {
			t.Errorf("unexpected host %s", endpoint)
		}
	}
}

func TestRandomLB_noEndpoints(t *testing.T) {
	subscriber := FixedSubscriberFactory(&config.Backend{})
	balancer := NewRandomLB(subscriber)
	_, err := balancer.Host()
	if want, have := ErrNoHosts, err; want != have {
		t.Errorf("want %v, have %v", want, have)
	}
}

type erroredSubscriber string

func (s erroredSubscriber) Hosts() ([]string, error) { return []string{}, errors.New(string(s)) }

func TestRoundRobinLB_erroredSubscriber(t *testing.T) {
	want := "sonic"
	balancer := NewRoundRobinLB(erroredSubscriber(want))
	host, have := balancer.Host()
	if host != "" || want != have.Error() {
		t.Errorf("want %s, have %s", want, have.Error())
	}
}

func TestRandomLB_erroredSubscriber(t *testing.T) {
	want := "turbo"
	balancer := NewRandomLB(erroredSubscriber(want))
	host, have := balancer.Host()
	if host != "" || want != have.Error() {
		t.Errorf("want %s, have %s", want, have.Error())
	}
}
