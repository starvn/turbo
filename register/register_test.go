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

package register

import "testing"

func TestNamespaced(t *testing.T) {
	r := New()
	r.Register("namespace1", "name1", 42)
	r.AddNamespace("namespace1")
	r.AddNamespace("namespace2")
	r.Register("namespace2", "name2", true)

	nr, ok := r.Get("namespace1")
	if !ok {
		t.Error("namespace1 not found")
		return
	}
	if _, ok := nr.Get("name2"); ok {
		t.Error("name2 found into namespace1")
		return
	}
	v1, ok := nr.Get("name1")
	if !ok {
		t.Error("name1 not found")
		return
	}
	if i, ok := v1.(int); !ok || i != 42 {
		t.Error("unexpected value:", v1)
	}

	nr, ok = r.Get("namespace2")
	if !ok {
		t.Error("namespace2 not found")
		return
	}
	if _, ok := nr.Get("name1"); ok {
		t.Error("name1 found into namespace2")
		return
	}
	v2, ok := nr.Get("name2")
	if !ok {
		t.Error("name2 not found")
		return
	}
	if b, ok := v2.(bool); !ok || !b {
		t.Error("unexpected value:", v2)
	}
}
