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

package router

import (
	"net/http"

	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/handlers"
)

// ShoelacesRouter sets up all routes and handlers for shoelaces
func ShoelacesRouter(env *environment.Environment) http.Handler {
	mux := http.NewServeMux()
	staticFiles := http.StripPrefix("/static/", http.FileServer(http.Dir(env.StaticDir)))
	staticConfigs := http.StripPrefix("/configs/static/", handlers.StaticConfigFileServer())
	dynamicConfigs := http.StripPrefix("/configs/", handlers.TemplateServer())

	// UI pages and assets.
	mux.Handle("GET /{$}", handlers.RenderDefaultTemplate("index"))
	mux.Handle("GET /events", handlers.RenderDefaultTemplate("events"))
	mux.Handle("GET /mappings", handlers.RenderDefaultTemplate("mappings"))
	mux.Handle("GET /static/", staticFiles)

	// UI JSON endpoints and manual boot selection.
	mux.HandleFunc("POST /update/target", handlers.UpdateTargetHandler)
	mux.HandleFunc("GET /ajax/servers", handlers.ServerListHandler)
	mux.HandleFunc("GET /ajax/events", handlers.ListEvents)
	mux.HandleFunc("GET /ajax/script/params", handlers.GetTemplateParams)

	// Static and templated configuration files served to booting hosts.
	mux.Handle("GET /configs/static/", staticConfigs)
	mux.Handle("GET /configs/", dynamicConfigs)

	// iPXE boot endpoints.
	mux.HandleFunc("GET /start", handlers.StartPollingHandler)
	mux.HandleFunc("GET /poll/1/{mac}", handlers.PollHandler)
	mux.HandleFunc("GET /ipxemenu", handlers.IPXEMenu)

	return mux
}
