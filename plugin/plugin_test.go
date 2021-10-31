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

package plugin

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestScan_ok(t *testing.T) {
	tmpDir, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)
	f, err := ioutil.TempFile(tmpDir, "test.so")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	f.Close()
	defer os.RemoveAll(tmpDir)

	tot, err := Scan(tmpDir, ".so")
	if len(tot) != 1 {
		t.Error("unexpected number of plugins found:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
}

func TestScan_noFolder(t *testing.T) {
	expectedErr := "open unknown: no such file or directory"
	tot, err := Scan("unknown", "")
	if len(tot) != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err == nil {
		t.Error("expecting error!")
		return
	}
	if err.Error() != expectedErr {
		t.Error("unexpected error:", err.Error())
	}
}

func TestScan_emptyFolder(t *testing.T) {
	name, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	tot, err := Scan(name, "")
	if len(tot) != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
	os.RemoveAll(name)
}

func TestScan_noMatches(t *testing.T) {
	tmpDir, err := ioutil.TempDir(".", "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)
	f, err := ioutil.TempFile(tmpDir, "test")
	if err != nil {
		t.Error("unexpected error:", err.Error())
		return
	}
	f.Close()
	defer os.RemoveAll(tmpDir)
	tot, err := Scan(tmpDir, ".so")
	if len(tot) != 0 {
		t.Error("unexpected number of plugins loaded:", tot)
	}
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
}
