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
	"context"
	"github.com/starvn/flatex/tree"
	"github.com/starvn/turbo/config"
	"strings"
)

type EntityFormatter interface {
	Format(Response) Response
}

type EntityFormatterFunc func(Response) Response

func (e EntityFormatterFunc) Format(entity Response) Response { return e(entity) }

type propertyFilter func(*Response)

type entityFormatter struct {
	Target         string
	Prefix         string
	PropertyFilter propertyFilter
	Mapping        map[string]string
}

func NewEntityFormatter(remote *config.Backend) EntityFormatter {
	if ef := newFlatmapFormatter(remote.ExtraConfig, remote.Target, remote.Group); ef != nil {
		return ef
	}

	var propertyFilter propertyFilter
	if len(remote.AllowList) > 0 {
		propertyFilter = newAllowlistingFilter(remote.AllowList)
	} else {
		propertyFilter = newDenylistingFilter(remote.DenyList)
	}
	sanitizedMappings := make(map[string]string, len(remote.Mapping))
	for i, m := range remote.Mapping {
		v := strings.Split(m, ".")
		sanitizedMappings[i] = v[0]
	}
	return entityFormatter{
		Target:         remote.Target,
		Prefix:         remote.Group,
		PropertyFilter: propertyFilter,
		Mapping:        sanitizedMappings,
	}
}

func (e entityFormatter) Format(entity Response) Response {
	if e.Target != "" {
		extractTarget(e.Target, &entity)
	}
	if len(entity.Data) > 0 {
		e.PropertyFilter(&entity)
	}
	if len(entity.Data) > 0 {
		for formerKey, newKey := range e.Mapping {
			if v, ok := entity.Data[formerKey]; ok {
				entity.Data[newKey] = v
				delete(entity.Data, formerKey)
			}
		}
	}
	if e.Prefix != "" {
		entity.Data = map[string]interface{}{e.Prefix: entity.Data}
	}
	return entity
}

func extractTarget(target string, entity *Response) {
	for _, part := range strings.Split(target, ".") {
		if tmp, ok := entity.Data[part]; ok {
			entity.Data, ok = tmp.(map[string]interface{})
			if !ok {
				entity.Data = map[string]interface{}{}
				return
			}
		} else {
			entity.Data = map[string]interface{}{}
			return
		}
	}
}

func AllowlistPrune(wlDict map[string]interface{}, inDict map[string]interface{}) bool {
	canDelete := true
	var deleteSibling bool
	for k, v := range inDict {
		deleteSibling = true
		if subWl, ok := wlDict[k]; ok {
			if subWlDict, okk := subWl.(map[string]interface{}); okk {
				if subInDict, isDict := v.(map[string]interface{}); isDict && !AllowlistPrune(subWlDict, subInDict) {
					deleteSibling = false
				}
			} else {
				deleteSibling = false
			}
		}
		if deleteSibling {
			delete(inDict, k)
		} else {
			canDelete = false
		}
	}
	return canDelete
}

func newAllowlistingFilter(Allowlist []string) propertyFilter {
	wlDict := make(map[string]interface{})
	for _, k := range Allowlist {
		wlFields := strings.Split(k, ".")
		d := buildDictPath(wlDict, wlFields[:len(wlFields)-1])
		d[wlFields[len(wlFields)-1]] = true
	}

	return func(entity *Response) {
		if AllowlistPrune(wlDict, entity.Data) {
			for k := range entity.Data {
				delete(entity.Data, k)
			}
		}
	}
}

func buildDictPath(accumulator map[string]interface{}, fields []string) map[string]interface{} {
	var ok bool
	var c map[string]interface{}
	var fIdx int
	fEnd := len(fields)
	p := accumulator
	for fIdx = 0; fIdx < fEnd; fIdx++ {
		if c, ok = p[fields[fIdx]].(map[string]interface{}); !ok {
			break
		}
		p = c
	}
	for ; fIdx < fEnd; fIdx++ {
		c = make(map[string]interface{})
		p[fields[fIdx]] = c
		p = c
	}
	return p
}

func newDenylistingFilter(blacklist []string) propertyFilter {
	bl := make(map[string][]string, len(blacklist))
	for _, key := range blacklist {
		keys := strings.Split(key, ".")
		if len(keys) > 1 {
			if sub, ok := bl[keys[0]]; ok {
				bl[keys[0]] = append(sub, keys[1])
			} else {
				bl[keys[0]] = []string{keys[1]}
			}
		} else {
			bl[keys[0]] = []string{}
		}
	}

	return func(entity *Response) {
		for k, sub := range bl {
			if len(sub) == 0 {
				delete(entity.Data, k)
			} else {
				if tmp := blacklistFilterSub(entity.Data[k], sub); len(tmp) > 0 {
					entity.Data[k] = tmp
				}
			}
		}
	}
}

func blacklistFilterSub(v interface{}, blacklist []string) map[string]interface{} {
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	for _, key := range blacklist {
		delete(tmp, key)
	}
	return tmp
}

const flatmapKey = "flatmap_filter"

type flatmapFormatter struct {
	Target string
	Prefix string
	Ops    []flatmapOp
}

type flatmapOp struct {
	Type string
	Args [][]string
}

func (e flatmapFormatter) Format(entity Response) Response {
	if e.Target != "" {
		extractTarget(e.Target, &entity)
	}

	e.processOps(&entity)

	if e.Prefix != "" {
		entity.Data = map[string]interface{}{e.Prefix: entity.Data}
	}
	return entity
}

func (e flatmapFormatter) processOps(entity *Response) {
	flatten, err := tree.New(entity.Data)
	if err != nil {
		return
	}
	for _, op := range e.Ops {
		switch op.Type {
		case "move":
			flatten.Move(op.Args[0], op.Args[1])
		case "append":
			flatten.Append(op.Args[0], op.Args[1])
		case "del":
			for _, k := range op.Args {
				flatten.Del(k)
			}
		default:
		}
	}

	entity.Data, _ = flatten.Get([]string{}).(map[string]interface{})
}

func newFlatmapFormatter(cfg config.ExtraConfig, target, group string) EntityFormatter {
	if v, ok := cfg[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if vs, ok := e[flatmapKey].([]interface{}); ok {
				if len(vs) == 0 {
					return nil
				}
				var ops []flatmapOp
				for _, v := range vs {
					m, ok := v.(map[string]interface{})
					if !ok {
						continue
					}
					op := flatmapOp{}
					if t, ok := m["type"].(string); ok {
						op.Type = t
					} else {
						continue
					}
					if args, ok := m["args"].([]interface{}); ok {
						op.Args = make([][]string, len(args))
						for k, arg := range args {
							if t, ok := arg.(string); ok {
								op.Args[k] = strings.Split(t, ".")
							}
						}
					}
					ops = append(ops, op)
				}
				if len(ops) == 0 {
					return nil
				}
				return &flatmapFormatter{
					Target: target,
					Prefix: group,
					Ops:    ops,
				}
			}
		}
	}
	return nil
}

func NewFlatmapMiddleware(cfg *config.EndpointConfig) Middleware {
	formatter := newFlatmapFormatter(cfg.ExtraConfig, "", "")
	return func(next ...Proxy) Proxy {
		if len(next) != 1 {
			panic(ErrTooManyProxies)
		}

		if formatter == nil {
			return next[0]
		}

		return func(ctx context.Context, request *Request) (*Response, error) {
			resp, err := next[0](ctx, request)
			if err != nil {
				return resp, err
			}
			r := formatter.Format(*resp)
			return &r, nil
		}
	}
}
