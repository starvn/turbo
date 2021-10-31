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

package encoding

import (
	"github.com/starvn/turbo/register"
	"io"
)

func GetRegister() *DecoderRegister {
	return decoders
}

type untypedRegister interface {
	Register(name string, v interface{})
	Get(name string) (interface{}, bool)
	Clone() map[string]interface{}
}

type DecoderRegister struct {
	data untypedRegister
}

func (r *DecoderRegister) Register(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error {
	r.data.Register(name, dec)
	return nil
}

func (r *DecoderRegister) Get(name string) func(bool) func(io.Reader, *map[string]interface{}) error {
	for _, n := range []string{name, JSON} {
		if v, ok := r.data.Get(n); ok {
			if dec, ok := v.(func(bool) func(io.Reader, *map[string]interface{}) error); ok {
				return dec
			}
		}
	}
	return NewJSONDecoder
}

var (
	decoders        = initDecoderRegister()
	defaultDecoders = map[string]func(bool) func(io.Reader, *map[string]interface{}) error{
		JSON:      NewJSONDecoder,
		SAFE_JSON: NewSafeJSONDecoder,
		STRING:    NewStringDecoder,
		NOOP:      noOpDecoderFactory,
	}
)

func initDecoderRegister() *DecoderRegister {
	r := &DecoderRegister{data: register.NewUntyped()}
	for k, v := range defaultDecoders {
		r.Register(k, v)
	}
	return r
}
