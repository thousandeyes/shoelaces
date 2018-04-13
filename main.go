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

package main

import (
	"net/http"
	"os"

	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/handlers"
	"github.com/thousandeyes/shoelaces/internal/router"
)

func main() {
	env := environment.New()
	app := handlers.MiddlewareChain(env).Then(router.ShoelacesRouter(env))

	env.Logger.Info("component", "main", "transport", "http", "addr", env.BaseURL, "msg", "listening")
	env.Logger.Error("component", "main", "err", http.ListenAndServe(env.BaseURL, app))

	os.Exit(1)
}
