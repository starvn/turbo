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

// Package register offers tools for creating and managing registers
package register

import "sync"

func New() *Namespaced {
	return &Namespaced{data: NewUntyped()}
}

type Namespaced struct {
	data *Untyped
}

func (n *Namespaced) Get(namespace string) (*Untyped, bool) {
	v, ok := n.data.Get(namespace)
	if !ok {
		return nil, ok
	}
	register, ok := v.(*Untyped)
	return register, ok
}

func (n *Namespaced) Register(namespace, name string, v interface{}) {
	if register, ok := n.Get(namespace); ok {
		register.Register(name, v)
		return
	}

	register := NewUntyped()
	register.Register(name, v)
	n.data.Register(namespace, register)
}

func (n *Namespaced) AddNamespace(namespace string) {
	if _, ok := n.Get(namespace); ok {
		return
	}
	n.data.Register(namespace, NewUntyped())
}

func NewUntyped() *Untyped {
	return &Untyped{
		data:  map[string]interface{}{},
		mutex: &sync.RWMutex{},
	}
}

type Untyped struct {
	data  map[string]interface{}
	mutex *sync.RWMutex
}

func (u *Untyped) Register(name string, v interface{}) {
	u.mutex.Lock()
	u.data[name] = v
	u.mutex.Unlock()
}

func (u *Untyped) Get(name string) (interface{}, bool) {
	u.mutex.RLock()
	v, ok := u.data[name]
	u.mutex.RUnlock()
	return v, ok
}

func (u *Untyped) Clone() map[string]interface{} {
	u.mutex.RLock()
	res := make(map[string]interface{}, len(u.data))
	for k, v := range u.data {
		res[k] = v
	}
	u.mutex.RUnlock()
	return res
}
