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

package gin

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/log"
	"io"
	"net/textproto"
)

const Namespace = "github.com/starvn/turbo/route/gin"

func NewEngine(cfg config.ServiceConfig, logger log.Logger, w io.Writer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		logger.Debug(logPrefix, "Debug enabled")
	}
	engine := gin.New()

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	var paths []string

	if v, ok := cfg.ExtraConfig[Namespace]; ok {
		if b, err := json.Marshal(v); err == nil {
			ginOptions := engineConfiguration{}
			if err := json.Unmarshal(b, &ginOptions); err == nil {
				engine.RedirectTrailingSlash = !ginOptions.DisableRedirectTrailingSlash
				engine.RedirectFixedPath = !ginOptions.DisableRedirectFixedPath
				engine.HandleMethodNotAllowed = !ginOptions.DisableHandleMethodNotAllowed
				engine.ForwardedByClientIP = ginOptions.ForwardedByClientIP
				engine.RemoteIPHeaders = ginOptions.RemoteIPHeaders
				for k, h := range engine.RemoteIPHeaders {
					engine.RemoteIPHeaders[k] = textproto.CanonicalMIMEHeaderKey(h)
				}
				engine.TrustedProxies = ginOptions.TrustedProxies
				engine.AppEngine = ginOptions.AppEngine
				engine.MaxMultipartMemory = ginOptions.MaxMultipartMemory
				engine.RemoveExtraSlash = ginOptions.RemoveExtraSlash
				paths = ginOptions.LoggerSkipPaths

				returnErrorMsg = ginOptions.ReturnErrorMsg
			}
		}
	}

	engine.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{Output: w, SkipPaths: paths}),
		gin.Recovery(),
	)

	return engine
}

type engineConfiguration struct {
	DisableRedirectTrailingSlash  bool     `json:"disable_redirect_trailing_slash"`
	DisableRedirectFixedPath      bool     `json:"disable_redirect_fixed_path"`
	DisableHandleMethodNotAllowed bool     `json:"disable_handle_method_not_allowed"`
	ForwardedByClientIP           bool     `json:"forwarded_by_client_ip"`
	RemoteIPHeaders               []string `json:"remote_ip_headers"`
	TrustedProxies                []string `json:"trusted_proxies"`
	AppEngine                     bool     `json:"app_engine"`
	MaxMultipartMemory            int64    `json:"max_multipart_memory"`
	RemoveExtraSlash              bool     `json:"remove_extra_slash"`
	LoggerSkipPaths               []string `json:"logger_skip_paths"`
	AutoOptions                   bool     `json:"auto_options"`
	ReturnErrorMsg                bool     `json:"return_error_msg"`
}

var returnErrorMsg bool
