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
	"context"
	"fmt"
	"github.com/starvn/turbo/log"
	turboplugin "github.com/starvn/turbo/plugin"
	"github.com/starvn/turbo/register"
	"net/http"
	"plugin"
	"strings"
)

var serverRegister = register.New()

func RegisterHandler(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
) {
	serverRegister.Register(Namespace, name, handler)
}

type Registerer interface {
	RegisterHandlers(func(
		name string,
		handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
	))
}

type LoggerRegisterer interface {
	RegisterLogger(interface{})
}

type RegisterHandlerFunc func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)

func Load(path, pattern string, rcf RegisterHandlerFunc) (int, error) {
	return LoadWithLogger(path, pattern, rcf, nil)
}

func LoadWithLogger(path, pattern string, rcf RegisterHandlerFunc, logger log.Logger) (int, error) {
	plugins, err := turboplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rcf, logger)
}

func load(plugins []string, rcf RegisterHandlerFunc, logger log.Logger) (int, error) {
	var errors []error
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rcf, logger); err != nil {
			errors = append(errors, fmt.Errorf("opening plugin %d (%s): %s", k, pluginName, err.Error()))
			continue
		}
		loadedPlugins++
	}

	if len(errors) > 0 {
		return loadedPlugins, loaderError{errors}
	}
	return loadedPlugins, nil
}

func open(pluginName string, rcf RegisterHandlerFunc, logger log.Logger) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	var p Plugin
	p, err = pluginOpener(pluginName)
	if err != nil {
		return
	}
	var r interface{}
	r, err = p.Lookup("HandlerRegisterer")
	if err != nil {
		return
	}
	registerer, ok := r.(Registerer)
	if !ok {
		return fmt.Errorf("http-server-handler plugin loader: unknown type")
	}

	if logger != nil {
		if lr, ok := r.(LoggerRegisterer); ok {
			lr.RegisterLogger(logger)
		}
	}

	registerer.RegisterHandlers(rcf)
	return
}

type Plugin interface {
	Lookup(name string) (plugin.Symbol, error)
}

var pluginOpener = defaultPluginOpener

func defaultPluginOpener(name string) (Plugin, error) {
	return plugin.Open(name)
}

type loaderError struct {
	errors []error
}

func (l loaderError) Error() string {
	msgs := make([]string, len(l.errors))
	for i, err := range l.errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("plugin loader found %d error(s): \n%s", len(msgs), strings.Join(msgs, "\n"))
}
