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
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"turbo/register"
)

func TestNoOpDecoder(t *testing.T) {
	decoders = initDecoderRegister()
	defer func() { decoders = initDecoderRegister() }()

	d := decoders.Get(NOOP)(false)

	errorMsg := erroredReader("this error should never been sent")
	var result map[string]interface{}
	if err := d(errorMsg, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if result != nil {
		t.Error("Unexpected value:", result)
	}
}

func TestRegister(t *testing.T) {
	decoders = initDecoderRegister()
	defer func() { decoders = initDecoderRegister() }()

	original := GetRegister()

	if len(original.data.Clone()) != 4 {
		t.Error("Unexpected number of registered factories:", len(original.data.Clone()))
	}

	decoders = &DecoderRegister{data: register.NewUntyped()}
	err := decoders.Register("some", NewJSONDecoder)
	if err != nil {
		return
	}

	if len(decoders.data.Clone()) != 1 {
		t.Error("Unexpected number of registered factories:", len(decoders.data.Clone()))
	}
}

func TestGet(t *testing.T) {
	decoders = initDecoderRegister()
	defer func() { decoders = initDecoderRegister() }()

	if len(decoders.data.Clone()) != 4 {
		t.Error("Unexpected number of registered factories:", len(decoders.data.Clone()))
	}

	checkDecoder(t, JSON)
	checkDecoder(t, "some")

	decoders = &DecoderRegister{data: register.NewUntyped()}
	err := decoders.Register("some", NewJSONDecoder)
	if err != nil {
		return
	}

	if len(decoders.data.Clone()) != 1 {
		t.Error("Unexpected number of registered factories:", len(decoders.data.Clone()))
	}

	checkDecoder(t, JSON)
	checkDecoder(t, "some")
}

func TestRegister_complete_ok(t *testing.T) {
	decoders = initDecoderRegister()
	defer func() { decoders = initDecoderRegister() }()

	expectedMsg := "a custom message to decode"
	expectedResponse := map[string]interface{}{"a": 42}

	if err := decoders.Register("custom", func(_ bool) func(io.Reader, *map[string]interface{}) error {
		return func(r io.Reader, v *map[string]interface{}) error {
			d, err := ioutil.ReadAll(r)
			if err != nil {
				t.Error(err)
				return err
			}
			if expectedMsg != string(d) {
				t.Errorf("unexpected msg: %s", string(d))
				return errors.New("unexpected msg to decode")
			}
			*v = expectedResponse
			return nil
		}
	}); err != nil {
		t.Error(err)
		return
	}

	decoder := decoders.Get("custom")(false)
	input := strings.NewReader(expectedMsg)
	var result map[string]interface{}
	if err := decoder(input, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if v, ok := result["a"]; !ok || v.(int) != 42 {
		t.Error("Unexpected value:", result)
	}
}

func TestRegister_complete_ko(t *testing.T) {
	decoders = initDecoderRegister()
	defer func() { decoders = initDecoderRegister() }()

	expectedMsg := "a custom message to decode"
	expectedErr := errors.New("expect me")

	if err := decoders.Register("custom", func(_ bool) func(io.Reader, *map[string]interface{}) error {
		return func(r io.Reader, v *map[string]interface{}) error {
			d, err := ioutil.ReadAll(r)
			if err != nil {
				t.Error(err)
				return err
			}
			if expectedMsg != string(d) {
				t.Errorf("unexpected msg: %s", string(d))
				return errors.New("unexpected msg to decode")
			}
			// v = nil
			return expectedErr
		}
	}); err != nil {
		t.Error(err)
		return
	}

	decoder := decoders.Get("custom")(false)
	input := strings.NewReader(expectedMsg)
	var result map[string]interface{}
	if err := decoder(input, &result); err != expectedErr {
		t.Error("Unexpected error:", err)
	}
	if result != nil {
		t.Error("Unexpected value:", result)
	}
}

func checkDecoder(t *testing.T, name string) {
	d := decoders.Get(name)(false)

	input := strings.NewReader(`{"foo": "bar"}`)
	var result map[string]interface{}
	if err := d(input, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if result["foo"] != "bar" {
		t.Error("Unexpected value:", result["foo"])
	}
}

type erroredReader string

func (e erroredReader) Error() string {
	return string(e)
}

func (e erroredReader) Read(_ []byte) (n int, err error) {
	return 0, e
}
