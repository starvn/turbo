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

import "testing"

func TestURIParser_cleanHosts(t *testing.T) {
	samples := []string{
		"sonic",
		"127.0.0.1",
		"https://sonic.local/",
		"http://127.0.0.1",
		"sonic_42.local:8080/",
		"http://127.0.0.1:8080",
	}

	expected := []string{
		"http://sonic",
		"http://127.0.0.1",
		"https://sonic.local",
		"http://127.0.0.1",
		"http://sonic_42.local:8080",
		"http://127.0.0.1:8080",
	}

	result := NewURIParser().CleanHosts(samples)
	for i := range result {
		if expected[i] != result[i] {
			t.Errorf("want: %s, have: %s\n", expected[i], result[i])
		}
	}
}

func TestURIParser_cleanPath(t *testing.T) {
	samples := []string{
		"sonic/{turbo}",
		"sonic/{turbo}{sonic}",
		"/sonic/{turbo}",
		"/sonic.local/",
		"sonic_sonic.txt",
		"sonic_42.local?a=8080",
		"sonic/sonic/sonic?a=1&b=2",
		"debug/sonic/sonic?a=1&b=2",
	}

	expected := []string{
		"/sonic/{turbo}",
		"/sonic/{turbo}{sonic}",
		"/sonic/{turbo}",
		"/sonic.local/",
		"/sonic_sonic.txt",
		"/sonic_42.local?a=8080",
		"/sonic/sonic/sonic?a=1&b=2",
		"/debug/sonic/sonic?a=1&b=2",
	}

	subject := URI(BracketsRouterPatternBuilder)

	for i := range samples {
		if have := subject.CleanPath(samples[i]); expected[i] != have {
			t.Errorf("want: %s, have: %s\n", expected[i], have)
		}
	}
}

func TestURIParser_getEndpointPath(t *testing.T) {
	samples := []string{
		"sonic/{turbo}",
		"/sonic/{turbo}{sonic}",
		"/sonic/{turbo}",
		"/sonic.local/",
		"sonic/{turbo}/{sonic}?a={s}&b=2",
	}

	expected := []string{
		"sonic/:turbo",
		"/sonic/:turbo{sonic}",
		"/sonic/:turbo",
		"/sonic.local/",
		"sonic/:turbo/:sonic?a={s}&b=2",
	}

	sc := ServiceConfig{}
	subject := NewURIParser()

	for i := range samples {
		params := sc.extractPlaceHoldersFromURLTemplate(samples[i], sc.paramExtractionPattern())
		if have := subject.GetEndpointPath(samples[i], params); expected[i] != have {
			t.Errorf("want: %s, have: %s\n", expected[i], have)
		}
	}
}
func TestURIParser_getEndpointPath_notStrictREST(t *testing.T) {
	samples := []string{
		"sonic/{turbo}",
		"/sonic/{turbo}{sonic}",
		"/sonic/{turbo}",
		"/sonic.local/",
		"sonic/{turbo}/{sonic}?a={s}&b=2",
	}

	expected := []string{
		"sonic/:turbo",
		"/sonic/:turbo:sonic",
		"/sonic/:turbo",
		"/sonic.local/",
		"sonic/:turbo/:sonic?a={s}&b=2",
	}

	sc := ServiceConfig{DisableStrictREST: true}
	subject := NewURIParser()

	for i := range samples {
		params := sc.extractPlaceHoldersFromURLTemplate(samples[i], sc.paramExtractionPattern())
		if have := subject.GetEndpointPath(samples[i], params); expected[i] != have {
			t.Errorf("want: %s, have: %s\n", expected[i], have)
		}
	}
}
