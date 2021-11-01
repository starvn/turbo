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
	"github.com/starvn/turbo/register"
)

func RegisterSubscriberFactory(name string, sf SubscriberFactory) error {
	return subscriberFactories.Register(name, sf)
}

func GetSubscriber(cfg *config.Backend) Subscriber {
	return subscriberFactories.Get(cfg.SD)(cfg)
}

func GetRegister() *Register {
	return subscriberFactories
}

type untypedRegister interface {
	Register(name string, v interface{})
	Get(name string) (interface{}, bool)
}

type Register struct {
	data untypedRegister
}

func initRegister() *Register {
	return &Register{register.NewUntyped()}
}

func (r *Register) Register(name string, sf SubscriberFactory) error {
	r.data.Register(name, sf)
	return nil
}

func (r *Register) Get(name string) SubscriberFactory {
	tmp, ok := r.data.Get(name)
	if !ok {
		return FixedSubscriberFactory
	}
	sf, ok := tmp.(SubscriberFactory)
	if !ok {
		return FixedSubscriberFactory
	}
	return sf
}

var subscriberFactories = initRegister()
