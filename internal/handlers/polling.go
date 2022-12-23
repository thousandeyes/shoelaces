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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/thousandeyes/shoelaces/internal/log"
	"github.com/thousandeyes/shoelaces/internal/polling"
	"github.com/thousandeyes/shoelaces/internal/server"
	"github.com/thousandeyes/shoelaces/internal/utils"
)

// StartPollingHandler is called by iPXE boot agents. It returns the poll script.
func StartPollingHandler(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)

	script := polling.GenStartScript(env.Logger,  env.BaseURL)

	w.Write([]byte(script))
}


// PollHandler is called by iPXE boot agents. It returns the boot script
// specified on the configuration or, if the host is unknown, it makes it
// retry for a while until the user specifies alternative IPXE boot script.
func PollHandler(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	// iPXE MAC addresses come with dashes instead of colons
	mac := utils.MacDashToColon(vars["mac"])
	host := r.FormValue("host")

	err = validateMACAndIP(env.Logger, mac, ip)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if host == "" {
		host = resolveHostname(env.Logger, ip)
	}

	server := server.New(mac, ip, host)
	script, err := polling.Poll(
		env.Logger, env.ServerStates, env.HostnameMaps, env.NetworkMaps,
		env.EventLog, env.Templates, env.BaseURL, server)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(script))
}

// ServerListHandler provides a list of the servers that tried to boot
// but did not match the hostname regex or network mappings.
func ServerListHandler(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)

	servers, err := json.Marshal(polling.ListServers(env.ServerStates))
	if err != nil {
		env.Logger.Error("component", "handler", "err", err)
		os.Exit(1)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(servers)
}

// UpdateTargetHandler is a POST endpoint that receives parameters for
// booting manually.
func UpdateTargetHandler(w http.ResponseWriter, r *http.Request) {
	env := envFromRequest(r)

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mac, scriptName, environment, params := parsePostForm(r.PostForm)
	if mac == "" || scriptName == "" {
		http.Error(w, "MAC address and target must not be empty", http.StatusBadRequest)
		return
	}

	server := server.New(mac, ip, "")
	inputErr, err := polling.UpdateTarget(
		env.Logger, env.ServerStates, env.Templates, env.EventLog, env.BaseURL, server,
		scriptName, environment, params)

	if err != nil {
		if inputErr {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func parsePostForm(form map[string][]string) (mac, scriptName, environment string, params map[string]interface{}) {
	params = make(map[string]interface{})
	for k, v := range form {
		if k == "mac" {
			mac = utils.MacDashToColon(v[0])
		} else if k == "target" {
			scriptName = v[0]
		} else if k == "environment" {
			environment = v[0]
		} else {
			params[k] = v[0]
		}
	}
	return
}

func validateMACAndIP(logger log.Logger, mac string, ip string) (err error) {
	if !utils.IsValidMAC(mac) {
		logger.Error("component", "polling", "msg", "Invalid MAC", "mac", mac)
		return fmt.Errorf("%s", "Invalid MAC")
	}

	if !utils.IsValidIP(ip) {
		logger.Error("component", "polling", "msg", "Invalid IP", "ip", ip)
		return fmt.Errorf("%s", "Invalid IP")
	}

	logger.Debug("component", "polling", "msg", "MAC and IP validated", "mac", mac, "ip", ip)

	return nil
}

func resolveHostname(logger log.Logger, ip string) string {
	host := utils.ResolveHostname(ip)
	if host == "" {
		logger.Info("component", "polling", "msg", "Can't resolve IP", "ip", ip)
	}

	return host
}
