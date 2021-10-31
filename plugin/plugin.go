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

// Package plugin provides tools for loading and registering plugins
package plugin

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

func Scan(folder, pattern string) ([]string, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return []string{}, err
	}

	var plugins []string
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), pattern) {
			plugins = append(plugins, filepath.Join(folder, file.Name()))
		}
	}

	return plugins, nil
}
