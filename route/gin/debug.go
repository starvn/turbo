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
	"github.com/starvn/turbo/log"
	"io/ioutil"
)

func DebugHandler(logger log.Logger) gin.HandlerFunc {
	logPrefixSecondary := "[ENDPOINT: /__debug/*]"
	return func(c *gin.Context) {
		logger.Debug(logPrefixSecondary, "Method:", c.Request.Method)
		logger.Debug(logPrefixSecondary, "URL:", c.Request.RequestURI)
		logger.Debug(logPrefixSecondary, "Query:", c.Request.URL.Query())
		logger.Debug(logPrefixSecondary, "Params:", c.Params)
		logger.Debug(logPrefixSecondary, "Headers:", c.Request.Header)
		body, _ := ioutil.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		logger.Debug(logPrefixSecondary, "Body:", string(body))
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}
