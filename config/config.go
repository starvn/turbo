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

// Package config defines the config structs and some config parser interfaces and implementations
package config

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/starvn/turbo/encoding"
	"net/textproto"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	BracketsRouterPatternBuilder = iota
	ColonRouterPatternBuilder
	DefaultMaxIdleConnectionsPerHost = 250
	DefaultTimeout                   = 2 * time.Second
	TurboConfigVersion               = 1
)

var RoutingPattern = ColonRouterPatternBuilder

type ServiceConfig struct {
	Name                      string            `mapstructure:"name"`
	Endpoints                 []*EndpointConfig `mapstructure:"endpoints"`
	Timeout                   time.Duration     `mapstructure:"timeout"`
	CacheTTL                  time.Duration     `mapstructure:"cache_ttl"`
	Host                      []string          `mapstructure:"host"`
	Port                      int               `mapstructure:"port"`
	Version                   int               `mapstructure:"version"`
	OutputEncoding            string            `mapstructure:"output_encoding"`
	ExtraConfig               ExtraConfig       `mapstructure:"extra_config"`
	ReadTimeout               time.Duration     `mapstructure:"read_timeout"`
	WriteTimeout              time.Duration     `mapstructure:"write_timeout"`
	IdleTimeout               time.Duration     `mapstructure:"idle_timeout"`
	ReadHeaderTimeout         time.Duration     `mapstructure:"read_header_timeout"`
	DisableKeepAlives         bool              `mapstructure:"disable_keep_alives"`
	DisableCompression        bool              `mapstructure:"disable_compression"`
	MaxIdleConnections        int               `mapstructure:"max_idle_connections"`
	MaxIdleConnectionsPerHost int               `mapstructure:"max_idle_connections_per_host"`
	IdleConnectionTimeout     time.Duration     `mapstructure:"idle_connection_timeout"`
	ResponseHeaderTimeout     time.Duration     `mapstructure:"response_header_timeout"`
	ExpectContinueTimeout     time.Duration     `mapstructure:"expect_continue_timeout"`
	DialerTimeout             time.Duration     `mapstructure:"dialer_timeout"`
	DialerFallbackDelay       time.Duration     `mapstructure:"dialer_fallback_delay"`
	DialerKeepAlive           time.Duration     `mapstructure:"dialer_keep_alive"`
	DisableStrictREST         bool              `mapstructure:"disable_rest"`
	Plugin                    *Plugin           `mapstructure:"plugin"`
	TLS                       *TLS              `mapstructure:"tls"`
	Debug                     bool
	uriParser                 URIParser
}

type EndpointConfig struct {
	Endpoint        string        `mapstructure:"endpoint"`
	Method          string        `mapstructure:"method"`
	Backend         []*Backend    `mapstructure:"backend"`
	ConcurrentCalls int           `mapstructure:"concurrent_calls"`
	Timeout         time.Duration `mapstructure:"timeout"`
	CacheTTL        time.Duration `mapstructure:"cache_ttl"`
	QueryString     []string      `mapstructure:"querystring_params"`
	ExtraConfig     ExtraConfig   `mapstructure:"extra_config"`
	HeadersToPass   []string      `mapstructure:"headers_to_pass"`
	OutputEncoding  string        `mapstructure:"output_encoding"`
}

type Backend struct {
	Group                    string            `mapstructure:"group"`
	Method                   string            `mapstructure:"method"`
	Host                     []string          `mapstructure:"host"`
	HostSanitizationDisabled bool              `mapstructure:"disable_host_sanitize"`
	URLPattern               string            `mapstructure:"url_pattern"`
	AllowList                []string          `mapstructure:"allow"`
	DenyList                 []string          `mapstructure:"deny"`
	Mapping                  map[string]string `mapstructure:"mapping"`
	Encoding                 string            `mapstructure:"encoding"`
	IsCollection             bool              `mapstructure:"is_collection"`
	Target                   string            `mapstructure:"target"`
	SD                       string            `mapstructure:"sd"`
	URLKeys                  []string
	ConcurrentCalls          int
	Timeout                  time.Duration
	Decoder                  encoding.Decoder `json:"-"`
	ExtraConfig              ExtraConfig      `mapstructure:"extra_config"`
}

type Plugin struct {
	Folder  string `mapstructure:"folder"`
	Pattern string `mapstructure:"pattern"`
}

type TLS struct {
	IsDisabled               bool     `mapstructure:"disabled"`
	PublicKey                string   `mapstructure:"public_key"`
	PrivateKey               string   `mapstructure:"private_key"`
	MinVersion               string   `mapstructure:"min_version"`
	MaxVersion               string   `mapstructure:"max_version"`
	CurvePreferences         []uint16 `mapstructure:"curve_preferences"`
	PreferServerCipherSuites bool     `mapstructure:"prefer_server_cipher_suites"`
	CipherSuites             []uint16 `mapstructure:"cipher_suites"`
	EnableMTLS               bool     `mapstructure:"enable_mtls"`
}

type ExtraConfig map[string]interface{}

func (e *ExtraConfig) sanitize() {
	for module, extra := range *e {
		switch extra := extra.(type) {
		case map[interface{}]interface{}:
			sanitized := map[string]interface{}{}
			for k, v := range extra {
				sanitized[fmt.Sprintf("%v", k)] = v
			}
			(*e)[module] = sanitized
		}

		if alias, ok := ExtraConfigAlias[module]; ok {
			(*e)[alias] = (*e)[module]
			delete(*e, module)
		}
	}
}

var ExtraConfigAlias = map[string]string{}

const defaultNamespace = "github.com/starvn/turbo/config"

var (
	simpleURLKeysPattern    = regexp.MustCompile(`{([a-zA-Z\-_0-9.]+)}`)
	sequentialParamsPattern = regexp.MustCompile(`^(resp[\d]+_.*)?(JWT\.([\w\-.]*))?$`)
	debugPattern            = "^[^/]|/__debug(/.*)?$"
	errInvalidHost          = errors.New("invalid host")
	errInvalidNoOpEncoding  = errors.New("can not use NoOp encoding with more than one backends connected to the same endpoint")
	defaultPort             = 8080
)

func (s *ServiceConfig) Hash() (string, error) {
	var name string
	name, s.Name = s.Name, ""
	defer func() { s.Name = name }()

	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}

func (s *ServiceConfig) Init() error {
	s.uriParser = NewURIParser()

	if s.Version != TurboConfigVersion {
		return &UnsupportedVersionError{
			Have: s.Version,
			Want: TurboConfigVersion,
		}
	}

	s.initGlobalParams()

	return s.initEndpoints()
}

func (s *ServiceConfig) initGlobalParams() {
	if s.Port == 0 {
		s.Port = defaultPort
	}
	if s.MaxIdleConnectionsPerHost == 0 {
		s.MaxIdleConnectionsPerHost = DefaultMaxIdleConnectionsPerHost
	}
	if s.Timeout == 0 {
		s.Timeout = DefaultTimeout
	}

	s.Host = s.uriParser.CleanHosts(s.Host)

	s.ExtraConfig.sanitize()
}

func (s *ServiceConfig) initEndpoints() error {
	for i, e := range s.Endpoints {
		e.Endpoint = s.uriParser.CleanPath(e.Endpoint)

		if err := e.validate(); err != nil {
			return err
		}

		for i := range e.HeadersToPass {
			e.HeadersToPass[i] = textproto.CanonicalMIMEHeaderKey(e.HeadersToPass[i])
		}

		inputParams := s.extractPlaceHoldersFromURLTemplate(e.Endpoint, s.paramExtractionPattern())
		inputSet := map[string]interface{}{}
		for ip := range inputParams {
			inputSet[inputParams[ip]] = nil
		}

		e.Endpoint = s.uriParser.GetEndpointPath(e.Endpoint, inputParams)

		s.initEndpointDefaults(i)

		if e.OutputEncoding == encoding.NOOP && len(e.Backend) > 1 {
			return errInvalidNoOpEncoding
		}

		e.ExtraConfig.sanitize()

		for j, b := range e.Backend {
			s.initBackendDefaults(i, j)

			if err := s.initBackendURLMappings(i, j, inputSet); err != nil {
				return err
			}

			b.ExtraConfig.sanitize()
		}
	}
	return nil
}

func (s *ServiceConfig) paramExtractionPattern() *regexp.Regexp {
	if s.DisableStrictREST {
		return simpleURLKeysPattern
	}
	return endpointURLKeysPattern
}

func (s *ServiceConfig) extractPlaceHoldersFromURLTemplate(subject string, pattern *regexp.Regexp) []string {
	matches := pattern.FindAllStringSubmatch(subject, -1)
	keys := make([]string, len(matches))
	for k, v := range matches {
		keys[k] = v[1]
	}
	return keys
}

func (s *ServiceConfig) initEndpointDefaults(e int) {
	endpoint := s.Endpoints[e]
	if endpoint.Method == "" {
		endpoint.Method = "GET"
	}
	if s.CacheTTL != 0 && endpoint.CacheTTL == 0 {
		endpoint.CacheTTL = s.CacheTTL
	}
	if s.Timeout != 0 && endpoint.Timeout == 0 {
		endpoint.Timeout = s.Timeout
	}
	if endpoint.ConcurrentCalls == 0 {
		endpoint.ConcurrentCalls = 1
	}
	if endpoint.OutputEncoding == "" {
		if s.OutputEncoding != "" {
			endpoint.OutputEncoding = s.OutputEncoding
		} else {
			endpoint.OutputEncoding = encoding.JSON
		}
	}
}

func (s *ServiceConfig) initBackendDefaults(e, b int) {
	endpoint := s.Endpoints[e]
	backend := endpoint.Backend[b]
	if len(backend.Host) == 0 {
		backend.Host = s.Host
	} else if !backend.HostSanitizationDisabled {
		backend.Host = s.uriParser.CleanHosts(backend.Host)
	}
	if backend.Method == "" {
		backend.Method = endpoint.Method
	}
	backend.Timeout = endpoint.Timeout
	backend.ConcurrentCalls = endpoint.ConcurrentCalls
	backend.Decoder = encoding.GetRegister().Get(strings.ToLower(backend.Encoding))(backend.IsCollection)
}

func (s *ServiceConfig) initBackendURLMappings(e, b int, inputParams map[string]interface{}) error {
	backend := s.Endpoints[e].Backend[b]

	backend.URLPattern = s.uriParser.CleanPath(backend.URLPattern)

	outputParams, outputSetSize := uniqueOutput(s.extractPlaceHoldersFromURLTemplate(backend.URLPattern, simpleURLKeysPattern))

	ip := fromSetToSortedSlice(inputParams)

	if outputSetSize > len(ip) {
		return &WrongNumberOfParamsError{
			Endpoint:     s.Endpoints[e].Endpoint,
			Method:       s.Endpoints[e].Method,
			Backend:      b,
			InputParams:  ip,
			OutputParams: outputParams,
		}
	}

	backend.URLKeys = []string{}
	for _, output := range outputParams {
		if !sequentialParamsPattern.MatchString(output) {
			if _, ok := inputParams[output]; !ok {
				return &UndefinedOutputParamError{
					Param:        output,
					Endpoint:     s.Endpoints[e].Endpoint,
					Method:       s.Endpoints[e].Method,
					Backend:      b,
					InputParams:  ip,
					OutputParams: outputParams,
				}
			}
		}
		key := strings.Title(output[:1]) + output[1:]
		backend.URLPattern = strings.Replace(backend.URLPattern, "{"+output+"}", "{{."+key+"}}", -1)
		backend.URLKeys = append(backend.URLKeys, key)
	}
	return nil
}

func fromSetToSortedSlice(set map[string]interface{}) []string {
	res := make([]string, 0, len(set))
	for element := range set {
		res = append(res, element)
	}
	sort.Strings(res)
	return res
}

func uniqueOutput(output []string) ([]string, int) {
	sort.Strings(output)
	j := 0
	outputSetSize := 0
	for i := 1; i < len(output); i++ {
		if output[j] == output[i] {
			continue
		}
		if !sequentialParamsPattern.MatchString(output[j]) {
			outputSetSize++
		}
		j++
		output[j] = output[i]
	}
	if j == len(output) {
		return output, outputSetSize
	}
	return output[:j+1], outputSetSize
}

func (e *EndpointConfig) validate() error {
	matched, err := regexp.MatchString(debugPattern, e.Endpoint)
	if err != nil {
		return &EndpointMatchError{
			Err:    err,
			Path:   e.Endpoint,
			Method: e.Method,
		}
	}
	if matched {
		return &EndpointPathError{Path: e.Endpoint, Method: e.Method}
	}

	if len(e.Backend) == 0 {
		return &NoBackendsError{Path: e.Endpoint, Method: e.Method}
	}
	return nil
}

type EndpointMatchError struct {
	Path   string
	Method string
	Err    error
}

func (e *EndpointMatchError) Error() string {
	return fmt.Sprintf("ignoring the '%s %s' endpoint due to a parsing error: %s", e.Method, e.Path, e.Err.Error())
}

type NoBackendsError struct {
	Path   string
	Method string
}

func (n *NoBackendsError) Error() string {
	return "ignoring the '" + n.Method + " " + n.Path + "' endpoint, since it has 0 backends defined!"
}

type UnsupportedVersionError struct {
	Have int
	Want int
}

func (u *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("unsupported version: %d (want: %d)", u.Have, u.Want)
}

type EndpointPathError struct {
	Path   string
	Method string
}

func (e *EndpointPathError) Error() string {
	return "ignoring the '" + e.Method + " " + e.Path + "' endpoint, since it is invalid!!!"
}

type UndefinedOutputParamError struct {
	Endpoint     string
	Method       string
	Backend      int
	InputParams  []string
	OutputParams []string
	Param        string
}

func (u *UndefinedOutputParamError) Error() string {
	return fmt.Sprintf(
		"undefined output param '%s'! endpoint: %s %s, backend: %d. input: %v, output: %v",
		u.Param,
		u.Method,
		u.Endpoint,
		u.Backend,
		u.InputParams,
		u.OutputParams,
	)
}

type WrongNumberOfParamsError struct {
	Endpoint     string
	Method       string
	Backend      int
	InputParams  []string
	OutputParams []string
}

func (w *WrongNumberOfParamsError) Error() string {
	return fmt.Sprintf(
		"input and output params do not match. endpoint: %s %s, backend: %d. input: %v, output: %v",
		w.Method,
		w.Endpoint,
		w.Backend,
		w.InputParams,
		w.OutputParams,
	)
}
