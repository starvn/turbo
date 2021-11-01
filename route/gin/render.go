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
	"github.com/gin-gonic/gin"
	"github.com/starvn/turbo/config"
	"github.com/starvn/turbo/encoding"
	"github.com/starvn/turbo/proxy"
	"io"
	"net/http"
	"sync"
)

type Render func(*gin.Context, *proxy.Response)

const NEGOTIATE = "negotiate"

var (
	mutex          = &sync.RWMutex{}
	renderRegister = map[string]Render{
		NEGOTIATE:         negotiatedRender,
		encoding.STRING:   stringRender,
		encoding.JSON:     jsonRender,
		encoding.NOOP:     noopRender,
		"json-collection": jsonCollectionRender,
	}
)

func RegisterRender(name string, r Render) {
	mutex.Lock()
	renderRegister[name] = r
	mutex.Unlock()
}

func getRender(cfg *config.EndpointConfig) Render {
	fallback := jsonRender
	if len(cfg.Backend) == 1 {
		fallback = getWithFallback(cfg.Backend[0].Encoding, fallback)
	}

	if cfg.OutputEncoding == "" {
		return fallback
	}

	return getWithFallback(cfg.OutputEncoding, fallback)
}

func getWithFallback(key string, fallback Render) Render {
	mutex.RLock()
	r, ok := renderRegister[key]
	mutex.RUnlock()
	if !ok {
		return fallback
	}
	return r
}

func negotiatedRender(c *gin.Context, response *proxy.Response) {
	switch c.NegotiateFormat(gin.MIMEJSON, gin.MIMEPlain, gin.MIMEXML) {
	case gin.MIMEXML:
		xmlRender(c, response)
	case gin.MIMEPlain:
		yamlRender(c, response)
	default:
		jsonRender(c, response)
	}
}

func stringRender(c *gin.Context, response *proxy.Response) {
	status := c.Writer.Status()

	if response == nil {
		c.String(status, "")
		return
	}
	d, ok := response.Data["content"]
	if !ok {
		c.String(status, "")
		return
	}
	msg, ok := d.(string)
	if !ok {
		c.String(status, "")
		return
	}
	c.String(status, msg)
}

func jsonRender(c *gin.Context, response *proxy.Response) {
	status := c.Writer.Status()
	if response == nil {
		c.JSON(status, emptyResponse)
		return
	}
	c.JSON(status, response.Data)
}

func jsonCollectionRender(c *gin.Context, response *proxy.Response) {
	status := c.Writer.Status()
	if response == nil {
		c.JSON(status, []struct{}{})
		return
	}
	col, ok := response.Data["collection"]
	if !ok {
		c.JSON(status, []struct{}{})
		return
	}
	c.JSON(status, col)
}

func xmlRender(c *gin.Context, response *proxy.Response) {
	status := c.Writer.Status()
	if response == nil {
		c.XML(status, nil)
		return
	}
	d, ok := response.Data["content"]
	if !ok {
		c.XML(status, nil)
		return
	}
	c.XML(status, d)
}

func yamlRender(c *gin.Context, response *proxy.Response) {
	status := c.Writer.Status()
	if response == nil {
		c.YAML(status, emptyResponse)
		return
	}
	c.YAML(status, response.Data)
}

func noopRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	for k, vs := range response.Metadata.Headers {
		for _, v := range vs {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Status(response.Metadata.StatusCode)
	if response.Io == nil {
		return
	}
	_, _ = io.Copy(c.Writer, response.Io)
}

var emptyResponse = gin.H{}
