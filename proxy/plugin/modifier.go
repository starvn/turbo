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

// Package plugin provides tools for loading and registering proxy plugins
package plugin

import (
	"fmt"
	"github.com/starvn/turbo/log"
	turboplugin "github.com/starvn/turbo/plugin"
	"github.com/starvn/turbo/register"
	"plugin"
	"strings"
)

const (
	Namespace         = "github.com/starvn/turbo/proxy/plugin"
	requestNamespace  = "github.com/starvn/turbo/proxy/plugin/request"
	responseNamespace = "github.com/starvn/turbo/proxy/plugin/response"
)

var modifierRegister = register.New()

type ModifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error)

func GetRequestModifier(name string) (ModifierFactory, bool) {
	return getModifier(requestNamespace, name)
}

func GetResponseModifier(name string) (ModifierFactory, bool) {
	return getModifier(responseNamespace, name)
}

func getModifier(namespace, name string) (ModifierFactory, bool) {
	r, ok := modifierRegister.Get(namespace)
	if !ok {
		return nil, ok
	}
	m, ok := r.Get(name)
	if !ok {
		return nil, ok
	}
	res, ok := m.(func(map[string]interface{}) func(interface{}) (interface{}, error))
	if !ok {
		return nil, ok
	}
	return ModifierFactory(res), ok
}

func RegisterModifier(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
) {
	if appliesToRequest {
		fmt.Println("registering request modifier:", name)
		modifierRegister.Register(requestNamespace, name, modifierFactory)
	}
	if appliesToResponse {
		fmt.Println("registering response modifier:", name)
		modifierRegister.Register(responseNamespace, name, modifierFactory)
	}
}

type Registerer interface {
	RegisterModifiers(func(
		name string,
		modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
		appliesToRequest bool,
		appliesToResponse bool,
	))
}

type LoggerRegisterer interface {
	RegisterLogger(interface{})
}

type RegisterModifierFunc func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)

func LoadModifiers(path, pattern string, rmf RegisterModifierFunc) (int, error) {
	return LoadModifiersWithLogger(path, pattern, rmf, nil)
}

func LoadModifiersWithLogger(path, pattern string, rmf RegisterModifierFunc, logger log.Logger) (int, error) {
	plugins, err := turboplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rmf, logger)
}

func load(plugins []string, rmf RegisterModifierFunc, logger log.Logger) (int, error) {
	var errors []error
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rmf, logger); err != nil {
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

func open(pluginName string, rmf RegisterModifierFunc, logger log.Logger) (err error) {
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
	r, err = p.Lookup("ModifierRegisterer")
	if err != nil {
		return
	}
	registerer, ok := r.(Registerer)
	if !ok {
		return fmt.Errorf("modifier plugin loader: unknown type")
	}

	if logger != nil {
		if lr, ok := r.(LoggerRegisterer); ok {
			lr.RegisterLogger(logger)
		}
	}

	registerer.RegisterModifiers(rmf)
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
