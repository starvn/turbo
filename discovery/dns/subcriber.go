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

// Package dns defines some implementations for a dns based service discovery
package dns

import (
	"fmt"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/discovery"
	"net"
	"sync"
	"time"
)

const Namespace = "dns"

func Register() error {
	return discovery.RegisterSubscriberFactory(Namespace, SubscriberFactory)
}

var (
	TTL           = 30 * time.Second
	DefaultLookup = net.LookupSRV
)

func SubscriberFactory(cfg *config.Backend) discovery.Subscriber {
	return New(cfg.Host[0])
}

func New(name string) discovery.Subscriber {
	return NewDetailed(name, DefaultLookup, TTL)
}

func NewDetailed(name string, lookup lookup, ttl time.Duration) discovery.Subscriber {
	s := subscriber{
		name:   name,
		cache:  &discovery.FixedSubscriber{},
		mutex:  &sync.Mutex{},
		ttl:    ttl,
		lookup: lookup,
	}
	s.update()
	go s.loop()
	return s
}

type lookup func(service, proto, name string) (cname string, addrs []*net.SRV, err error)

type subscriber struct {
	name   string
	cache  *discovery.FixedSubscriber
	mutex  *sync.Mutex
	ttl    time.Duration
	lookup lookup
}

func (s subscriber) Hosts() ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.cache.Hosts()
}

func (s subscriber) loop() {
	for {
		<-time.After(s.ttl)
		s.update()
	}
}

func (s subscriber) update() {
	instances, err := s.resolve()
	if err != nil {
		return
	}
	s.mutex.Lock()
	*(s.cache) = discovery.FixedSubscriber(instances)
	s.mutex.Unlock()
}

func (s subscriber) resolve() ([]string, error) {
	_, address, err := s.lookup("", "", s.name)
	if err != nil {
		return []string{}, err
	}
	instances := []string{}
	for _, addr := range address {
		instances = append(instances, fmt.Sprintf("http://%s", net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))))
		for i := 0; i < int(addr.Weight-1); i++ {
			instances = append(instances, fmt.Sprintf("http://%s", net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))))
		}
	}
	return instances, nil
}
