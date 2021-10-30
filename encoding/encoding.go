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

// Package encoding provides Decoding implementations.
package encoding

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type Decoder func(io.Reader, *map[string]interface{}) error

type DecoderFactory func(bool) func(io.Reader, *map[string]interface{}) error

const NOOP = "no-op"

func NoOpDecoder(_ io.Reader, _ *map[string]interface{}) error { return nil }

func noOpDecoderFactory(_ bool) func(io.Reader, *map[string]interface{}) error { return NoOpDecoder }

const JSON = "json"

func NewJSONDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	if isCollection {
		return JSONCollectionDecoder
	}
	return JSONDecoder
}

func JSONDecoder(r io.Reader, v *map[string]interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}

func JSONCollectionDecoder(r io.Reader, v *map[string]interface{}) error {
	var collection []interface{}
	d := json.NewDecoder(r)
	d.UseNumber()
	if err := d.Decode(&collection); err != nil {
		return err
	}
	*(v) = map[string]interface{}{"collection": collection}
	return nil
}

const SAFE_JSON = "safejson"

func NewSafeJSONDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	return SafeJSONDecoder
}

func SafeJSONDecoder(r io.Reader, v *map[string]interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	var t interface{}
	if err := d.Decode(&t); err != nil {
		return err
	}
	switch tt := t.(type) {
	case map[string]interface{}:
		*v = tt
	case []interface{}:
		*v = map[string]interface{}{"collection": tt}
	default:
		*v = map[string]interface{}{"result": tt}
	}
	return nil
}

const STRING = "string"

func NewStringDecoder(_ bool) func(io.Reader, *map[string]interface{}) error {
	return StringDecoder
}

func StringDecoder(r io.Reader, v *map[string]interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	*(v) = map[string]interface{}{"content": string(data)}
	return nil
}
