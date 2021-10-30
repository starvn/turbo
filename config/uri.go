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

package config

import (
	"regexp"
	"strings"
)

var (
	endpointURLKeysPattern = regexp.MustCompile(`/{([a-zA-Z\-_0-9]+)}`)
	hostPattern            = regexp.MustCompile(`(https?://)?([a-zA-Z0-9._\-]+)(:[0-9]{2,6})?/?`)
)

type URIParser interface {
	CleanHosts([]string) []string
	CleanHost(string) string
	CleanPath(string) string
	GetEndpointPath(string, []string) string
}

func NewURIParser() URIParser {
	return URI(RoutingPattern)
}

type URI int

func (u URI) CleanHosts(hosts []string) []string {
	cleaned := make([]string, 0, len(hosts))
	for i := range hosts {
		cleaned = append(cleaned, u.CleanHost(hosts[i]))
	}
	return cleaned
}

func (URI) CleanHost(host string) string {
	matches := hostPattern.FindAllStringSubmatch(host, -1)
	if len(matches) != 1 {
		panic(errInvalidHost)
	}
	keys := matches[0][1:]
	if keys[0] == "" {
		keys[0] = "http://"
	}
	return strings.Join(keys, "")
}

func (URI) CleanPath(path string) string {
	return "/" + strings.TrimPrefix(path, "/")
}

func (u URI) GetEndpointPath(path string, params []string) string {
	result := path
	if u == ColonRouterPatternBuilder {
		for p := range params {
			parts := strings.Split(result, "?")
			parts[0] = strings.Replace(parts[0], "{"+params[p]+"}", ":"+params[p], -1)
			result = strings.Join(parts, "?")
		}
	}
	return result
}
