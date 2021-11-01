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

package main

import (
	"errors"
	"fmt"
	"github.com/starvn/turbo/log"
	"io"
	"net/url"
	"path"
)

func main() {}

func init() {
	fmt.Println(string(ModifierRegisterer), "loaded!!!")
}

var ModifierRegisterer = registerer("turbo-request-modifier-example")

var logger log.Logger = nil

type registerer string

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), r.modifierFactory, true, false)
	fmt.Println(string(ModifierRegisterer), "registered!!!")
}

func (r registerer) RegisterLogger(in interface{}) {
	l, ok := in.(log.Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(string(ModifierRegisterer), "logger registered!!!")

}

func (r registerer) modifierFactory(
	map[string]interface{},
) func(interface{}) (interface{}, error) {
	// check the config
	// return the modifier
	fmt.Println(string(ModifierRegisterer), "injected!!!")

	if logger == nil {
		return func(input interface{}) (interface{}, error) {
			req, ok := input.(RequestWrapper)
			if !ok {
				return nil, unkownTypeErr
			}

			return modifier(req), nil
		}
	}

	return func(input interface{}) (interface{}, error) {
		req, ok := input.(RequestWrapper)
		if !ok {
			return nil, unkownTypeErr
		}

		r := modifier(req)

		logger.Debug("params:", r.params)
		logger.Debug("headers:", r.headers)
		logger.Debug("method:", r.method)
		logger.Debug("url:", r.url)
		logger.Debug("query:", r.query)
		logger.Debug("path:", r.path)

		return r, nil
	}
}

func modifier(req RequestWrapper) requestWrapper {
	return requestWrapper{
		params:  req.Params(),
		headers: req.Headers(),
		body:    req.Body(),
		method:  req.Method(),
		url:     req.URL(),
		query:   req.Query(),
		path:    path.Join(req.Path(), "/fooo"),
	}
}

var unkownTypeErr = errors.New("unknow request type")

type RequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
}

type requestWrapper struct {
	method  string
	url     *url.URL
	query   url.Values
	path    string
	body    io.ReadCloser
	params  map[string]string
	headers map[string][]string
}

func (r requestWrapper) Method() string               { return r.method }
func (r requestWrapper) URL() *url.URL                { return r.url }
func (r requestWrapper) Query() url.Values            { return r.query }
func (r requestWrapper) Path() string                 { return r.path }
func (r requestWrapper) Body() io.ReadCloser          { return r.body }
func (r requestWrapper) Params() map[string]string    { return r.params }
func (r requestWrapper) Headers() map[string][]string { return r.headers }
