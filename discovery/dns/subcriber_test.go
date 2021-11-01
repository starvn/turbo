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

package dns

import (
	"errors"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/discovery"
	"net"
	"testing"
	"time"
)

func TestSubscriber_New(t *testing.T) {
	if err := Register(); err != nil {
		t.Error("registering the dns module:", err.Error())
	}
	srvSet := []*net.SRV{
		{
			Port:   80,
			Target: "127.0.0.1",
			Weight: 1,
		},
		{
			Port:   81,
			Target: "127.0.0.1",
			Weight: 2,
		},
	}
	DefaultLookup = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", srvSet, nil
	}

	s := discovery.GetSubscriber(&config.Backend{Host: []string{"some.example.tld"}, SD: Namespace})
	hosts, err := s.Hosts()
	if err != nil {
		t.Error("Getting the hosts:", err.Error())
	}
	if len(hosts) != 3 {
		t.Error("Wrong number of hosts:", len(hosts))
	}
	if hosts[0] != "http://127.0.0.1:80" {
		t.Error("Wrong host #0 (expected http://127.0.0.1:80):", hosts[0])
	}
	if hosts[1] != "http://127.0.0.1:81" {
		t.Error("Wrong host #1 (expected http://127.0.0.1:81):", hosts[1])
	}
	if hosts[2] != "http://127.0.0.1:81" {
		t.Error("Wrong host #2 (expected http://127.0.0.1:81):", hosts[2])
	}
}

func TestSubscriber_LoockupError(t *testing.T) {
	errToReturn := errors.New("Some random error")
	defaultLookup := func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", []*net.SRV{}, errToReturn
	}
	ttl := 1 * time.Millisecond
	s := NewDetailed("some.example.tld", defaultLookup, ttl)
	hosts, err := s.Hosts()
	if err != nil {
		t.Error("Unexpected error!", err)
	}
	if len(hosts) != 0 {
		t.Error("Wrong number of hosts:", len(hosts))
	}
}
