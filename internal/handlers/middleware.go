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

package handlers

import (
	"context"
	"github.com/justinas/alice"
	"net/http"
	"regexp"

	"github.com/thousandeyes/shoelaces/internal/environment"
)

// ShoelacesCtxID Shoelaces Specific Request Context ID.
type ShoelacesCtxID int

// ShoelacesEnvCtxID is the context id key for the shoelaces.Environment.
const ShoelacesEnvCtxID ShoelacesCtxID = 0

// ShoelacesEnvNameCtxID is the context ID key for the chosen environment.
const ShoelacesEnvNameCtxID ShoelacesCtxID = 1

var envRe = regexp.MustCompile(`^(:?/env\/([a-zA-Z0-9_-]+))?(\/.*)`)

// environmentMiddleware Rewrites the URL in case it was an environment
// specific and sets the environment in the context.
func environmentMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqEnv string
		m := envRe.FindStringSubmatch(r.URL.Path)
		if len(m) > 0 && m[2] != "" {
			r.URL.Path = m[3]
			reqEnv = m[2]
		}
		ctx := context.WithValue(r.Context(), ShoelacesEnvNameCtxID, reqEnv)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware adds an entry to the logger each time the HTTP service
// receives a request.
func loggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := envFromRequest(r).Logger

		logger.Info("component", "http", "type", "request", "src", r.RemoteAddr, "method", r.Method, "url", r.URL)
		h.ServeHTTP(w, r)
	})
}

// SecureHeaders adds secure headers to the responses
func secureHeadersMiddleware(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add X-XSS-Protection header
		w.Header().Add("X-XSS-Protection", "1; mode=block")

		// Add X-Content-Type-Options header
		w.Header().Add("X-Content-Type-Options", "nosniff")

		// Prevent page from being displayed in an iframe
		w.Header().Add("X-Frame-Options", "DENY")

		// Prevent page from being displayed in an iframe
		w.Header().Add("Content-Security-Policy", "script-src 'self'")

		h.ServeHTTP(w, r)
	})
}

// disableCacheMiddleware sets a header for disabling HTTP caching
func disableCacheMiddleware(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")

		h.ServeHTTP(w, r)
	})
}

// MiddlewareChain receives a Shoelaces environment and returns a chains of
// middlewares to apply to every request.
func MiddlewareChain(env *environment.Environment) alice.Chain {
	// contextMiddleware sets the environment key in the request Context.
	contextMiddleware := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ShoelacesEnvCtxID, env)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	return alice.New(
		secureHeadersMiddleware,
		disableCacheMiddleware,
		environmentMiddleware,
		contextMiddleware,
		loggingMiddleware)
}
