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
	"github.com/starvn/turbo/config"
	"testing"
)

func TestRegisterSubscriberFactory_ok(t *testing.T) {
	sf1 := func(*config.Backend) Subscriber {
		return SubscriberFunc(func() ([]string, error) { return []string{"one"}, nil })
	}
	sf2 := func(*config.Backend) Subscriber {
		return SubscriberFunc(func() ([]string, error) { return []string{"two", "three"}, nil })
	}
	if err := RegisterSubscriberFactory("name1", sf1); err != nil {
		t.Error(err)
	}
	if err := RegisterSubscriberFactory("name2", sf2); err != nil {
		t.Error(err)
	}

	if h, err := GetSubscriber(&config.Backend{SD: "name1"}).Hosts(); err != nil || len(h) != 1 {
		t.Error("error using the discovery name1")
	}

	if h, err := GetSubscriber(&config.Backend{SD: "name2"}).Hosts(); err != nil || len(h) != 2 {
		t.Error("error using the discovery name2")
	}

	if h, err := GetRegister().Get("name2")(&config.Backend{SD: "name2"}).Hosts(); err != nil || len(h) != 2 {
		t.Error("error using the discovery name2")
	}

	subscriberFactories = initRegister()
}

func TestRegisterSubscriberFactory_unknown(t *testing.T) {
	if h, err := GetSubscriber(&config.Backend{Host: []string{"name"}}).Hosts(); err != nil || len(h) != 1 {
		t.Error("error using the default discovery")
	}
}

func TestRegisterSubscriberFactory_errored(t *testing.T) {
	subscriberFactories.data.Register("errored", true)
	if h, err := GetSubscriber(&config.Backend{SD: "errored", Host: []string{"name"}}).Hosts(); err != nil || len(h) != 1 {
		t.Error("error using the default discovery")
	}
	subscriberFactories = initRegister()
}
