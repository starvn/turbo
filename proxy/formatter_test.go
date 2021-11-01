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
	"context"
	"github.com/starvn/turbo/config"
	"reflect"
	"testing"
)

func TestEntityFormatterFunc(t *testing.T) {
	expected := Response{Data: map[string]interface{}{"one": 1}, IsComplete: true}
	f := func(_ Response) Response { return expected }
	formatter := EntityFormatterFunc(f)
	result := formatter.Format(Response{})
	if result.Data["one"].(int) != 1 {
		t.Error("unexpected result:", result.Data)
	}
	if !result.IsComplete {
		t.Error("unexpected result:", result)
	}
}

func TestEntityFormatter_newAllowFilter(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "turbo",
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{AllowList: []string{"sonic", "a.b", "a.c", "foo.unknown"}})
	result := f.Format(sample)
	if v, ok := result.Data["sonic"]; !ok || v != expected.Data["sonic"] {
		t.Errorf("The formatter returned an unexpected result for the field sonic: %v\n", result)
	}
	v, ok := result.Data["a"]
	if !ok {
		t.Errorf("The formatter returned an unexpected result for the fields a.b & a.c: %v\n", result)
	}
	tmp := v.(map[string]interface{})
	if b, okk := tmp["b"]; !okk || !b.(bool) {
		t.Errorf("The formatter returned an unexpected result for the field a.b: %v\n", result)
	}
	if c, okk := tmp["c"]; !okk || c.(int) != 42 {
		t.Errorf("The formatter returned an unexpected result for the field a.c: %v\n", result)
	}
	if len(tmp) != 2 {
		t.Errorf("The formatter returned an unexpected result size for the field a: %v\n", result)
	}
	if len(result.Data) != 2 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_newAllowDeepFields(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"id": 42,
			"turbo": map[string]interface{}{
				"muku": map[string]interface{}{
					"sonic": 1,
					"muku":  2,
					"gutu": map[string]interface{}{
						"kugu": 42,
					},
				},
				"sonic": map[string]interface{}{
					"sonic": 3,
					"muku":  4,
				},
			},
		},
		IsComplete: true,
	}
	expectedSonicChild := 1

	var ok bool
	f := NewEntityFormatter(&config.Backend{AllowList: []string{"turbo.muku.sonic", "turbo.muku.gutu.kugu"}})
	res := f.Format(sample)
	var turbo map[string]interface{}
	var muku map[string]interface{}
	var gutu map[string]interface{}
	var kugu int
	var sonicChild int
	if turbo, ok = res.Data["turbo"].(map[string]interface{}); !ok {
		t.Errorf("The formatter does not have field turbo\n")
	}
	if muku, ok = turbo["muku"].(map[string]interface{}); !ok {
		t.Errorf("The formatter does not have field turbo.muku\n")
	}
	if sonicChild, ok = muku["sonic"].(int); !ok || sonicChild != expectedSonicChild {
		t.Errorf("The formatter does not have field turbo.muku.sonic or wrong value\n")
	}
	if _, ok = turbo["sonic"].(map[string]interface{}); ok {
		t.Errorf("The formatter should have removed turbo.sonic\n")
	}
	if _, ok = muku["muku"]; ok {
		t.Errorf("The formatter should have removed turbo.muku.muku\n")
	}
	if gutu, ok = muku["gutu"].(map[string]interface{}); !ok {
		t.Errorf("The formatter does not have field turbo.muku.gutu\n")
	}
	if kugu, ok = gutu["kugu"].(int); !ok || kugu != 42 {
		t.Errorf("The formatter does not have field turbo.muku.gutu.kugu\n")
	}
}

func TestEntityFormatter_newDenyFilter(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "turbo",
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"turbo": false,
			"foo":   "bar",
			"a": map[string]interface{}{
				"d": "turbo",
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{DenyList: []string{"sonic", "a.b", "a.c", "foo.unknown"}})
	result := f.Format(sample)
	if v, ok := result.Data["turbo"]; !ok || v != expected.Data["turbo"] {
		t.Errorf("The formatter returned an unexpected result for the field turbo: %v\n", result)
	}
	if v, ok := result.Data["foo"]; !ok || v != expected.Data["foo"] {
		t.Errorf("The formatter returned an unexpected result for the field foo: %v\n", result)
	}
	v, ok := result.Data["a"]
	if !ok {
		t.Errorf("The formatter returned an unexpected result for the field a.d: %v\n", result)
	}
	tmp := v.(map[string]interface{})
	if d, okk := tmp["d"]; !okk || d != "turbo" {
		t.Errorf("The formatter returned an unexpected result for the field a.d: %v\n", result)
	}
	if len(tmp) != 1 {
		t.Errorf("The formatter returned an unexpected result size for the field a: %v\n", result)
	}
	if len(result.Data) != 3 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_grouping(t *testing.T) {
	prefix := "group1"
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			prefix: map[string]interface{}{
				"sonic": 42,
				"turbo": false,
				"foo":   "bar",
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{Group: prefix})
	result := f.Format(sample)
	if len(result.Data) != 1 || result.IsComplete != expected.IsComplete {
		t.Fail()
	}
	if _, ok := result.Data[prefix]; !ok {
		t.Fail()
	}
	group := result.Data[prefix].(map[string]interface{})
	for k, expectedValue := range expected.Data[prefix].(map[string]interface{}) {
		if v, ok := group[k]; !ok || v != expectedValue {
			t.Fail()
		}
	}
}

func TestEntityFormatter_mapping(t *testing.T) {
	mapping := map[string]string{"sonic": "SONICCCCC", "turbo": "TURBOOOOO", "a.b": "a.BOOOOO"}

	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "turbo",
	}
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
			"a":     sub,
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"SONICCCCC": 42,
			"TURBOOOOO": false,
			"foo":       "bar",
			"a":         sub,
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{Mapping: mapping})
	result := f.Format(sample)

	if len(result.Data) != 4 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result.Data)
	}
	for k, expectedValue := range expected.Data {
		if k == "a" {
			continue
		}
		if v, ok := result.Data[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}
	group := result.Data["a"].(map[string]interface{})
	for k, expectedValue := range expected.Data["a"].(map[string]interface{}) {
		if v, ok := group[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}

	if len(group) != 3 {
		t.Errorf("The formatter returned an unexpected result size for the subentity: %v\n", group)
	}
}

func TestEntityFormatter_targeting(t *testing.T) {
	target := "group1"
	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "turbo",
	}
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
			target:  sub,
		},
		IsComplete: true,
	}
	expected := Response{
		Data:       sub,
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{Target: target})
	result := f.Format(sample)
	if len(result.Data) != 3 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
	for k, expectedValue := range expected.Data {
		if v, ok := result.Data[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}
}

func TestEntityFormatter_targetingNested(t *testing.T) {
	target := "group1"
	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "turbo",
	}
	sample := Response{
		Data: map[string]interface{}{
			target: map[string]interface{}{
				"sonic": 42,
				"turbo": false,
				"foo":   "bar",
				target:  sub,
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data:       sub,
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{Target: target + "." + target})
	result := f.Format(sample)
	if len(result.Data) != 3 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
	for k, expectedValue := range expected.Data {
		if v, ok := result.Data[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}
}

func TestEntityFormatter_targetingUnknownFields(t *testing.T) {
	target := "group1"
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{Target: target})
	result := f.Format(sample)
	if len(result.Data) != 0 || result.IsComplete != sample.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_targetingNonObjects(t *testing.T) {
	target := "group1"
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
			target:  false,
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{Target: target})
	result := f.Format(sample)
	if len(result.Data) != 0 || result.IsComplete != sample.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_altogether(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"sonic": 42,
			"turbo": false,
			"foo":   "bar",
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "turbo",
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"group": map[string]interface{}{
				"D": "turbo",
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{
		Target:    "a",
		AllowList: []string{"d"},
		Group:     "group",
		Mapping:   map[string]string{"d": "D"},
	})
	result := f.Format(sample)
	v, ok := result.Data["group"]
	if !ok {
		t.Errorf("The formatter returned an unexpected result for the field group.D: %v\n", result)
	}
	tmp := v.(map[string]interface{})
	if d, okk := tmp["D"]; !okk || d != "turbo" {
		t.Errorf("The formatter returned an unexpected result for the field group.D: %v\n", result)
	}
	if len(tmp) != 1 {
		t.Errorf("The formatter returned an unexpected result size for the field group: %v\n", result)
	}
	if len(result.Data) != 1 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_flatmap(t *testing.T) {
	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "turbo",
		"e": []interface{}{1, 2, 3, 4},
	}
	sample := Response{
		Data: map[string]interface{}{
			"content": map[string]interface{}{
				"sonic":      42,
				"turbo":      false,
				"foo":        "bar",
				"a":          sub,
				"collection": []interface{}{sub, sub, sub, sub},
				"y":          []interface{}{0, 1, 2, 3, 4, 5, 6},
				"z":          []interface{}{10, 11, 12, 13, 14, 15, 16},
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"group": map[string]interface{}{
				"SONICCCCC": 42,
				"turbo":     false,
				"foo":       "bar",
				"a": map[string]interface{}{
					"BOOOOO": true,
					"c":      42,
					"d":      "turbo",
					"e":      []interface{}{1, 2, 3, 4},
				},
				"collection": []interface{}{
					map[string]interface{}{"x": 42},
					map[string]interface{}{"x": 42},
					map[string]interface{}{"x": 42},
					map[string]interface{}{"x": 42},
				},
				"z": []interface{}{10, 11, 12, 13, 14, 15, 16, 0, 1, 2, 3, 4, 5, 6},
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(&config.Backend{
		Target: "content",
		Group:  "group",
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				flatmapKey: []interface{}{
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"c"},
					},
					map[string]interface{}{
						"type": "append",
						"args": []interface{}{"y", "z"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"sonic", "SONICCCCC"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"a.b", "a.BOOOOO"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{
							"collection.*.b",
							"collection.*.d",
							"collection.*.e",
						},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"collection.*.c", "collection.*.x"},
					},
				},
			},
		},
	})
	result := f.Format(sample)

	if len(result.Data) != len(expected.Data) || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result.Data)
	}

	if !reflect.DeepEqual(expected.Data, result.Data) {
		t.Errorf("unexpected result: %v", result.Data)
	}
}

func TestNewFlatmapMiddleware(t *testing.T) {
	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "turbo",
		"e": []interface{}{1, 2, 3, 4},
	}
	sample := Response{
		Data: map[string]interface{}{
			"sonic":      42,
			"turbo":      false,
			"foo":        "bar",
			"a":          sub,
			"collection": []interface{}{sub, sub, sub, sub},
			"y":          []interface{}{0, 1, 2, 3, 4, 5, 6},
			"z":          []interface{}{10, 11, 12, 13, 14, 15, 16},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"SONICCCCC": 42,
			"turbo":     false,
			"foo":       "bar",
			"a": map[string]interface{}{
				"BOOOOO": true,
				"c":      42,
				"d":      "turbo",
				"e":      []interface{}{1, 2, 3, 4},
			},
			"collection": []interface{}{
				map[string]interface{}{"x": 42},
				map[string]interface{}{"x": 42},
				map[string]interface{}{"x": 42},
				map[string]interface{}{"x": 42},
			},
			"z": []interface{}{10, 11, 12, 13, 14, 15, 16, 0, 1, 2, 3, 4, 5, 6},
		},
		IsComplete: true,
	}
	p := NewFlatmapMiddleware(&config.EndpointConfig{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				flatmapKey: []interface{}{
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"c"},
					},
					map[string]interface{}{
						"type": "append",
						"args": []interface{}{"y", "z"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"sonic", "SONICCCCC"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"a.b", "a.BOOOOO"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"collection.*.b"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"collection.*.d"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"collection.*.e"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"collection.*.c", "collection.*.x"},
					},
				},
			},
		},
	})(func(_ context.Context, _ *Request) (*Response, error) {
		return &sample, nil
	})

	result, err := p(context.TODO(), nil)

	if err != nil {
		t.Error(err)
	}

	if len(result.Data) != len(expected.Data) {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result.Data)
	}

	if result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected completion flag: %v\n", result.IsComplete)
	}

	if !reflect.DeepEqual(expected.Data, result.Data) {
		t.Errorf("unexpected result: %v", result.Data)
	}
}
