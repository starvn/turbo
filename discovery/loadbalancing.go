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
	"github.com/valyala/fastrand"
	"runtime"
	"sync/atomic"
)

type Balancer interface {
	Host() (string, error)
}

var ErrNoHosts = errors.New("no hosts available")

func NewBalancer(subscriber Subscriber) Balancer {
	if p := runtime.GOMAXPROCS(-1); p == 1 {
		return NewRoundRobinLB(subscriber)
	}
	return NewRandomLB(subscriber)
}

func NewRoundRobinLB(subscriber Subscriber) Balancer {
	s, ok := subscriber.(FixedSubscriber)
	start := uint64(0)
	if ok {
		if l := len(s); l == 1 {
			return nopBalancer(s[0])
		} else if l > 1 {
			start = uint64(fastrand.Uint32n(uint32(l)))
		}
	}
	return &roundRobinLB{
		balancer: balancer{subscriber: subscriber},
		counter:  start,
	}
}

type roundRobinLB struct {
	balancer
	counter uint64
}

func (r *roundRobinLB) Host() (string, error) {
	hosts, err := r.hosts()
	if err != nil {
		return "", err
	}
	offset := (atomic.AddUint64(&r.counter, 1) - 1) % uint64(len(hosts))
	return hosts[offset], nil
}

func NewRandomLB(subscriber Subscriber) Balancer {
	if s, ok := subscriber.(FixedSubscriber); ok && len(s) == 1 {
		return nopBalancer(s[0])
	}
	return &randomLB{
		balancer: balancer{subscriber: subscriber},
		rand:     fastrand.Uint32n,
	}
}

type randomLB struct {
	balancer
	rand func(uint32) uint32
}

func (r *randomLB) Host() (string, error) {
	hosts, err := r.hosts()
	if err != nil {
		return "", err
	}
	return hosts[int(r.rand(uint32(len(hosts))))], nil
}

type balancer struct {
	subscriber Subscriber
}

func (b *balancer) hosts() ([]string, error) {
	hs, err := b.subscriber.Hosts()
	if err != nil {
		return hs, err
	}
	if len(hs) <= 0 {
		return hs, ErrNoHosts
	}
	return hs, nil
}

type nopBalancer string

func (b nopBalancer) Host() (string, error) { return string(b), nil }
