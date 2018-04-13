// Copyright 2018 ThousandEyes Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"io"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Logger struct holds a log.Logger plus functions required for logging
// with different levels. They functions are syntactic sugar to avoid
// having to import "github.com/go-kit/kit/log/level" in every package that
// has to cast a log.
type Logger struct {
	Raw   log.Logger
	Info  func(...interface{}) error
	Debug func(...interface{}) error
	Error func(...interface{}) error
}

const callerLevel int = 6

// MakeLogger receives a io.Writer and return a Logger struct.
func MakeLogger(w io.Writer) Logger {
	raw := log.NewLogfmtLogger(log.NewSyncWriter(w))
	raw = log.With(raw, "ts", log.DefaultTimestampUTC, "caller", log.Caller(callerLevel))
	filtered := level.NewFilter(raw, level.AllowInfo())

	return Logger{
		Raw:   raw,
		Info:  level.Info(filtered).Log,
		Debug: level.Debug(filtered).Log,
		Error: level.Error(filtered).Log,
	}
}

// AllowDebug receives a Logger and enables the debug logging level.
func AllowDebug(l Logger) Logger {
	filtered := level.NewFilter(l.Raw, level.AllowDebug())

	return Logger{
		Raw:   l.Raw,
		Info:  level.Info(filtered).Log,
		Debug: level.Debug(filtered).Log,
		Error: level.Error(filtered).Log,
	}
}
