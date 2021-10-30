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

// Package log provides a simple logger interface
package log

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}

const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelCritical
)

var (
	ErrInvalidLogLevel = errors.New("invalid log level")
	defaultLogger      = BasicLogger{Level: LevelCritical, Logger: log.New(os.Stderr, "", log.LstdFlags)}
	logLevels          = map[string]int{
		"DEBUG":    LevelDebug,
		"INFO":     LevelInfo,
		"WARNING":  LevelWarning,
		"ERROR":    LevelError,
		"CRITICAL": LevelCritical,
	}
	NoOp, _ = NewLogger("CRITICAL", ioutil.Discard, "")
)

func NewLogger(level string, out io.Writer, prefix string) (BasicLogger, error) {
	l, ok := logLevels[strings.ToUpper(level)]
	if !ok {
		return defaultLogger, ErrInvalidLogLevel
	}
	return BasicLogger{Level: l, Prefix: prefix, Logger: log.New(out, "", log.LstdFlags)}, nil
}

type BasicLogger struct {
	Level  int
	Prefix string
	Logger *log.Logger
}

func (l BasicLogger) Debug(v ...interface{}) {
	if l.Level > LevelDebug {
		return
	}
	l.prependLog("DEBUG:", v...)
}

func (l BasicLogger) Info(v ...interface{}) {
	if l.Level > LevelInfo {
		return
	}
	l.prependLog("INFO:", v...)
}

func (l BasicLogger) Warning(v ...interface{}) {
	if l.Level > LevelWarning {
		return
	}
	l.prependLog("WARNING:", v...)
}

func (l BasicLogger) Error(v ...interface{}) {
	if l.Level > LevelError {
		return
	}
	l.prependLog("ERROR:", v...)
}

func (l BasicLogger) Critical(v ...interface{}) {
	l.prependLog("CRITICAL:", v...)
}

func (l BasicLogger) Fatal(v ...interface{}) {
	l.prependLog("FATAL:", v...)
	os.Exit(1)
}

func (l BasicLogger) prependLog(level string, v ...interface{}) {
	msg := make([]interface{}, len(v)+2)
	msg[0] = l.Prefix
	msg[1] = level
	copy(msg[2:], v[:])
	l.Logger.Println(msg...)
}
