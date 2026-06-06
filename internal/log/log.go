// Copyright 2018-2026 ThousandEyes Inc.
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
	"log/slog"
)

// Logger wraps slog with the level switch Shoelaces uses for debug mode.
type Logger struct {
	*slog.Logger
	level *slog.LevelVar
}

// MakeLogger receives a io.Writer and return a Logger struct.
func MakeLogger(w io.Writer) Logger {
	level := &slog.LevelVar{}
	level.Set(slog.LevelInfo)

	return Logger{
		Logger: slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
			AddSource: true,
			Level:     level,
		})),
		level: level,
	}
}

// AllowDebug receives a Logger and enables the debug logging level.
func AllowDebug(l Logger) Logger {
	l.level.Set(slog.LevelDebug)
	return l
}
