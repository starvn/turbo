/*
 * Copyright (c) 2021 Huy Duc Dao
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package graphql

import (
	"encoding/json"
	"errors"
	"github.com/starvn/turbo/config"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const Namespace = "github.com/starvn/turbo/transport/http/client/graphql"

type OperationType string

type OperationMethod string

const (
	OperationMutation OperationType = "mutation"
	OperationQuery    OperationType = "query"

	MethodPost OperationMethod = http.MethodPost
	MethodGet  OperationMethod = http.MethodGet
)

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

type Options struct {
	GraphQLRequest
	QueryPath string          `json:"query_path,omitempty"`
	Type      OperationType   `json:"type"`
	Method    OperationMethod `json:"method"`
}

var errNoConfigFound = errors.New("graphql: no configuration found")

func GetOptions(cfg config.ExtraConfig) (*Options, error) {
	tmp, ok := cfg[Namespace]
	if !ok {
		return nil, errNoConfigFound
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		return nil, err
	}

	var opt Options
	if err := json.Unmarshal(b, &opt); err != nil {
		return nil, err
	}

	opt.Method = OperationMethod(strings.ToUpper(string(opt.Method)))
	opt.Type = OperationType(strings.ToLower(string(opt.Type)))

	if opt.Method != MethodGet && opt.Method != MethodPost {
		opt.Method = MethodPost
	}

	if opt.QueryPath != "" {
		q, err := ioutil.ReadFile(opt.QueryPath)
		if err != nil {
			return nil, err
		}
		opt.Query = string(q)
	}

	return &opt, nil
}

func New(opt Options) Extractor {
	var replacements [][2]string
	for k, v := range opt.Variables {
		val, ok := v.(string)
		if !ok {
			continue
		}
		if val[0] == '{' && val[len(val)-1] == '}' {
			replacements = append(replacements, [2]string{k, strings.Title(val[1:2]) + val[2:len(val)-1]})
		}
	}

	if len(replacements) == 0 {
		b, _ := json.Marshal(opt.GraphQLRequest)

		return Extractor{
			cfg: opt,
			paramExtractor: func(map[string]string) (*GraphQLRequest, error) {
				return &opt.GraphQLRequest, nil
			},
			newBody: func(_ map[string]string) ([]byte, error) {
				return b, nil
			},
		}
	}

	paramExtractor := func(params map[string]string) (*GraphQLRequest, error) {
		val := GraphQLRequest{
			Query:         opt.Query,
			OperationName: opt.OperationName,
			Variables:     map[string]interface{}{},
		}
		for k, v := range opt.Variables {
			val.Variables[k] = v
		}
		for _, vs := range replacements {
			val.Variables[vs[0]] = params[vs[1]]
		}
		return &val, nil
	}

	return Extractor{
		cfg:            opt,
		paramExtractor: paramExtractor,
		newBody: func(params map[string]string) ([]byte, error) {
			val, err := paramExtractor(params)
			if err != nil {
				return []byte{}, err
			}
			return json.Marshal(val)
		},
	}
}

type Extractor struct {
	cfg            Options
	paramExtractor func(map[string]string) (*GraphQLRequest, error)
	newBody        func(map[string]string) ([]byte, error)
}

func (e Extractor) QueryFromBody(r io.Reader) (url.Values, error) {
	gr, err := e.fromBody(r)
	if err != nil {
		return nil, err
	}
	vars := url.Values{}

	vars.Add("query", gr.Query)
	if gr.OperationName != "" {
		vars.Add("operationName", gr.OperationName)
	}
	if len(gr.Variables) != 0 {
		encodedVars, _ := json.Marshal(gr.Variables)
		vars.Add("variables", string(encodedVars))
	}

	return vars, nil
}

func (e Extractor) BodyFromBody(r io.Reader) ([]byte, error) {
	v, err := e.fromBody(r)
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(v)
}

func (e Extractor) fromBody(r io.Reader) (*GraphQLRequest, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	vars := map[string]interface{}{}

	if err := json.Unmarshal(b, &vars); err != nil {
		return nil, err
	}

	for k, v := range e.cfg.Variables {
		if _, ok := vars[k]; ok {
			continue
		}
		vars[k] = v
	}

	return &GraphQLRequest{
		Query:         e.cfg.Query,
		OperationName: e.cfg.OperationName,
		Variables:     vars,
	}, nil
}

func (e Extractor) QueryFromParams(params map[string]string) (url.Values, error) {
	gr, err := e.paramExtractor(params)
	if err != nil {
		return nil, err
	}
	vars := url.Values{}

	vars.Add("query", gr.Query)
	if gr.OperationName != "" {
		vars.Add("operationName", gr.OperationName)
	}
	if len(gr.Variables) != 0 {
		encodedVars, _ := json.Marshal(gr.Variables)
		vars.Add("variables", string(encodedVars))
	}

	return vars, nil
}

func (e Extractor) BodyFromParams(params map[string]string) ([]byte, error) {
	return e.newBody(params)
}
