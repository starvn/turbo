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
	"github.com/starvn/turbo/register"
)

func NewRegister() *Register {
	return &Register{
		responseCombiners,
	}
}

type Register struct {
	*combinerRegister
}

type combinerRegister struct {
	data     *register.Untyped
	fallback ResponseCombiner
}

func newCombinerRegister(data map[string]ResponseCombiner, fallback ResponseCombiner) *combinerRegister {
	r := register.NewUntyped()
	for k, v := range data {
		r.Register(k, v)
	}
	return &combinerRegister{r, fallback}
}

func (r *combinerRegister) GetResponseCombiner(name string) (ResponseCombiner, bool) {
	v, ok := r.data.Get(name)
	if !ok {
		return r.fallback, ok
	}
	if rc, ok := v.(ResponseCombiner); ok {
		return rc, ok
	}
	return r.fallback, ok
}

func (r *combinerRegister) SetResponseCombiner(name string, rc ResponseCombiner) {
	r.data.Register(name, rc)
}
