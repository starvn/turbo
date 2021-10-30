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
	"encoding/json"
	"strings"
	"testing"
)

func TestNewJSONDecoder_map(t *testing.T) {
	decoder := NewJSONDecoder(false)
	original := strings.NewReader(`{"foo": "bar", "sonic": false, "turbo": 4.20}`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if len(result) != 3 {
		t.Error("Unexpected result:", result)
	}
	if v, ok := result["foo"]; !ok || v.(string) != "bar" {
		t.Error("wrong result:", result)
	}
	if v, ok := result["sonic"]; !ok || v.(bool) {
		t.Error("wrong result:", result)
	}
	if v, ok := result["turbo"]; !ok || v.(json.Number).String() != "4.20" {
		t.Error("wrong result:", result)
	}
}

func TestNewJSONDecoder_collection(t *testing.T) {
	decoder := NewJSONDecoder(true)
	original := strings.NewReader(`["foo", "bar", "sonic"]`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if len(result) != 1 {
		t.Error("Unexpected result:", result)
	}
	v, ok := result["collection"]
	if !ok {
		t.Error("wrong result:", result)
	}
	embedded := v.([]interface{})
	if embedded[0].(string) != "foo" {
		t.Error("wrong result:", result)
	}
	if embedded[1].(string) != "bar" {
		t.Error("wrong result:", result)
	}
	if embedded[2].(string) != "sonic" {
		t.Error("wrong result:", result)
	}
}

func TestNewJSONDecoder_ko(t *testing.T) {
	decoder := NewJSONDecoder(true)
	original := strings.NewReader(`3`)
	var result map[string]interface{}
	if err := decoder(original, &result); err == nil {
		t.Error("Expecting error!")
	}
}

func TestNewSafeJSONDecoder_map(t *testing.T) {
	decoder := NewSafeJSONDecoder(false)
	original := strings.NewReader(`{"foo": "bar", "sonic": false, "turbo": 4.20}`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if len(result) != 3 {
		t.Error("Unexpected result:", result)
	}
	if v, ok := result["foo"]; !ok || v.(string) != "bar" {
		t.Error("wrong result:", result)
	}
	if v, ok := result["sonic"]; !ok || v.(bool) {
		t.Error("wrong result:", result)
	}
	if v, ok := result["turbo"]; !ok || v.(json.Number).String() != "4.20" {
		t.Error("wrong result:", result)
	}
}

func TestNewSafeJSONDecoder_collection(t *testing.T) {
	decoder := NewSafeJSONDecoder(true)
	original := strings.NewReader(`["foo", "bar", "sonic"]`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if len(result) != 1 {
		t.Error("Unexpected result:", result)
	}
	v, ok := result["collection"]
	if !ok {
		t.Error("wrong result:", result)
	}
	embedded := v.([]interface{})
	if embedded[0].(string) != "foo" {
		t.Error("wrong result:", result)
	}
	if embedded[1].(string) != "bar" {
		t.Error("wrong result:", result)
	}
	if embedded[2].(string) != "sonic" {
		t.Error("wrong result:", result)
	}
}

func TestNewSafeJSONDecoder_other(t *testing.T) {
	decoder := NewSafeJSONDecoder(true)
	original := strings.NewReader(`3`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if v, ok := result["result"]; !ok || v.(json.Number).String() != "3" {
		t.Error("wrong result:", result)
	}
}
