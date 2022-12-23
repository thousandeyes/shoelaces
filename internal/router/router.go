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

package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thousandeyes/shoelaces/internal/environment"
	"github.com/thousandeyes/shoelaces/internal/handlers"
)

// ShoelacesRouter sets up all routes and handlers for shoelaces
func ShoelacesRouter(env *environment.Environment) http.Handler {
	r := mux.NewRouter()

	// Main UI page
	r.Handle("/", handlers.RenderDefaultTemplate("index")).Methods("GET")
	// Event Log History page
	r.Handle("/events", handlers.RenderDefaultTemplate("events")).Methods("GET")
	// Currently configured mappings page
	r.Handle("/mappings", handlers.RenderDefaultTemplate("mappings")).Methods("GET")
	// Static files used by the UI
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir(env.StaticDir))))
	// Manual boot parameters POST endpoint
	r.HandleFunc("/update/target", handlers.UpdateTargetHandler).Methods("POST")
	// Provides a list of the servers that tried to boot but did not match
	// the hostname regex or network mappings
	r.HandleFunc("/ajax/servers", handlers.ServerListHandler).Methods("GET")
	// Event Log History JSON endpoint
	r.HandleFunc("/ajax/events", handlers.ListEvents).Methods("GET")
	// Provides the list of possible parameters for a given template
	r.HandleFunc("/ajax/script/params", handlers.GetTemplateParams)

	// Static configuration files endpoint
	r.PathPrefix("/configs/static/").Handler(http.StripPrefix("/configs/static/",
		handlers.StaticConfigFileServer()))

	// Dynamic configuration endpoint
	r.PathPrefix("/configs/").Handler(http.StripPrefix("/configs/",
		handlers.TemplateServer()))

	// Starting point for iPXE boot agents, usualy defined by DHCP server.
	// Gets the iPXE boot agents into the polling loop.
	r.HandleFunc("/start", handlers.StartPollingHandler).Methods("GET")
	// Called by iPXE boot agents, returns boot script specified on the configuration
	// or if the host is unknown makes it retry for a while until the user specifies
	// alternative ipxe boot script
	r.HandleFunc("/poll/1/{mac}", handlers.PollHandler).Methods("GET")
	// Serves a generated iPXE boot script providing a selection
	// of all of the boot scripts available on the filesystem for that environment.
	r.HandleFunc("/ipxemenu", handlers.IPXEMenu).Methods("GET")

	return r
}
