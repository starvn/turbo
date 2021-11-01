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
	"testing"
)

func TestNewRegister_responseCombiner_ok(t *testing.T) {
	r := NewRegister()
	r.SetResponseCombiner("name1", func(total int, parts []*Response) *Response {
		if total < 0 || total >= len(parts) {
			return nil
		}
		return parts[total]
	})

	rc, ok := r.GetResponseCombiner("name1")
	if !ok {
		t.Error("expecting response combiner")
		return
	}

	result := rc(0, []*Response{{IsComplete: true, Data: map[string]interface{}{"a": 42}}})

	if result == nil {
		t.Error("expecting result")
		return
	}

	if !result.IsComplete {
		t.Error("expecting a complete result")
		return
	}

	if len(result.Data) != 1 {
		t.Error("unexpected result size:", len(result.Data))
		return
	}
}

func TestNewRegister_responseCombiner_fallbackIfErrored(t *testing.T) {
	r := NewRegister()

	r.data.Register("errored", true)

	rc, ok := r.GetResponseCombiner("errored")
	if !ok {
		t.Error("expecting response combiner")
		return
	}

	original := &Response{IsComplete: true, Data: map[string]interface{}{"a": 42}}

	result := rc(0, []*Response{original})

	if result != original {
		t.Error("unexpected result:", result)
		return
	}
}

func TestNewRegister_responseCombiner_fallbackIfUnknown(t *testing.T) {
	r := NewRegister()

	rc, ok := r.GetResponseCombiner("unknown")
	if ok {
		t.Error("the response combiner should not be found")
		return
	}

	original := &Response{IsComplete: true, Data: map[string]interface{}{"a": 42}}

	result := rc(0, []*Response{original})

	if result != original {
		t.Error("unexpected result:", result)
		return
	}
}
