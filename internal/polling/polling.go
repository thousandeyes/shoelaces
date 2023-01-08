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

package polling

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"text/template"
	"time"

	"github.com/thousandeyes/shoelaces/internal/event"
	"github.com/thousandeyes/shoelaces/internal/log"
	"github.com/thousandeyes/shoelaces/internal/mappings"
	"github.com/thousandeyes/shoelaces/internal/server"
	"github.com/thousandeyes/shoelaces/internal/templates"
	"github.com/thousandeyes/shoelaces/internal/utils"
)

// ManualAction represent an action taken when no automatic boot is available.
type ManualAction int

const (
	startScript = "#!ipxe\n" +
		"echo Shoelaces starts polling\n" +
		"chain --autofree --replace \\\n" +
		"    http://{{.baseURL}}/poll/1/${netX/mac:hexhyp}\n" +
		"#\n" +
		"#\n" +
		"# Do\n" +
		"#    curl http://{{.baseURL}}/poll/1/06-66-de-ad-be-ef\n" +
		"# to get an idea about what the iPXE client will receive.\n"

	maxRetry = 10

	retryScript = "#!ipxe\n" +
		"prompt --key 0x02 --timeout 7000 shoelaces: Press Ctrl-B for manual override... \\\n" +
		"  && chain -ar http://{{.baseURL}}/ipxemenu \\\n" +
		"  || chain -ar http://{{.baseURL}}/poll/1/{{.macAddress}}\n" +
		"#\n" +
		"# FYI: For the iPXE client is the above an endless loop,\n" +
		"#      but it is the shoelaces server that decides if it loops.\n"

	timeoutScript = "#!ipxe\n" +
		"echo\n" +
		"echo Shoelaces is at maxRetry\n" +
		"echo\n" +
		"exit\n"

	// BootAction is used when a user selects a script for the polling
	// server. The server polls once again, so it gets the selected script
	// as answer.
	BootAction ManualAction = 0
	// RetryAction is used when a server polling does not yet have a script
	// selected by the user, hence it has to retry.
	RetryAction ManualAction = 1
	// TimeoutAction is used when a server polling is timing out.
	TimeoutAction ManualAction = 2
)

// ListServers provides a list of the servers that tried to boot
// but did not match the hostname regex or network mappings.
func ListServers(serverStates *server.States) server.Servers {
	ret := make([]server.Server, 0)

	serverStates.RLock()
	for _, s := range serverStates.Servers {
		if s.Target == server.InitTarget {
			ret = append(ret, s.Server)
		}
	}
	defer serverStates.RUnlock()
	sort.Sort(server.Servers(ret))

	return ret
}

// UpdateTarget receives parameters for booting manually. When a host
// didn't match any of the automatic methods for booting, it's going to be
// put on hold. This method is called when something is finally chosen for
// that host.
func UpdateTarget(logger log.Logger, serverStates *server.States,
	templateRenderer *templates.ShoelacesTemplates, eventLog *event.Log, baseURL string, srv server.Server,
	scriptName string, envName string, params map[string]interface{}) (inputErr bool, err error) {

	if !utils.IsValidMAC(srv.Mac) {
		return true, errors.New("Invalid MAC")
	}
	// Test the template with user inputs
	setHostName(params, srv.Mac)

	params["baseURL"] = utils.BaseURLforEnvName(baseURL, envName)
	_, err = templateRenderer.RenderTemplate(logger, scriptName, params, envName)
	if err != nil {
		inputErr = true
		return
	}

	serverStates.Lock()
	defer serverStates.Unlock()
	servers := serverStates.Servers
	if servers[srv.Mac] == nil {
		return true, errors.New("MAC is not in the booting state")
	}

	hostname := servers[srv.Mac].Server.Hostname
	logger.Debug("component", "polling", "msg", "Setting server override", "server", srv.Mac, "target", scriptName, "environment", envName, "hostname", hostname, "params", params)
	eventLog.AddEvent(event.UserSelection, srv, "", scriptName, nil)
	servers[srv.Mac].Target = scriptName
	servers[srv.Mac].Environment = envName
	servers[srv.Mac].Params = params
	return false, nil
}

// Poll contains the main logic of Shoelaces. It uses several heuristics to find
// the right script to return, as network maps, hostname maps and manual
// selection.
func Poll(logger log.Logger, serverStates *server.States,
	hostnameMaps []mappings.HostnameMap, networkMaps []mappings.NetworkMap,
	eventLog *event.Log, templateRenderer *templates.ShoelacesTemplates,
	baseURL string, srv server.Server) (scriptText string, err error) {

	script, found := attemptAutomaticBoot(logger, hostnameMaps, networkMaps, templateRenderer, eventLog, baseURL, srv)
	if found {
		return script, nil
	}

	return manualAction(logger, serverStates, templateRenderer, eventLog, baseURL, srv)
}

func attemptAutomaticBoot(logger log.Logger, hostnameMaps []mappings.HostnameMap, networkMaps []mappings.NetworkMap,
	templateRenderer *templates.ShoelacesTemplates, eventLog *event.Log,
	baseURL string, srv server.Server) (scriptText string, found bool) {

	// Find with reverse hostname matched with the hostname regexps
	if script, found := mappings.FindScriptForHostname(hostnameMaps, srv.Hostname); found {
		logger.Debug("component", "polling", "msg", "Host found", "where", "hostname-mapping", "host", srv.Hostname)
		eventLog.AddEvent(event.HostBoot, srv, event.PtrMatchBoot, script.Name, script.Params)
		script.Params["hostname"] = srv.Hostname

		return genBootScript(logger, templateRenderer, baseURL, script), found
	}
	logger.Debug("component", "polling", "msg", "Host not found", "where", "hostname-mapping", "host", srv.Hostname)

	// Find with IP belonging to a configured subnet
	if script, found := mappings.FindScriptForNetwork(networkMaps, srv.IP); found {
		logger.Debug("component", "polling", "msg", "Host found", "where", "network-mapping", "ip", srv.IP)
		setHostName(script.Params, srv.Mac)
		srv.Hostname = script.Params["hostname"].(string)
		eventLog.AddEvent(event.HostBoot, srv, event.SubnetMatchBoot, script.Name, script.Params)

		return genBootScript(logger, templateRenderer, baseURL, script), found
	}
	logger.Debug("component", "polling", "msg", "Host not found", "where", "network-mapping", "ip", srv.IP)

	return "", false
}

func manualAction(logger log.Logger, serverStates *server.States, templateRenderer *templates.ShoelacesTemplates,
	eventLog *event.Log, baseURL string, srv server.Server) (scriptText string, err error) {

	script, action := chooseManualAction(logger, serverStates, eventLog, srv)
	logger.Debug("component", "polling", "target-script-name", script, "action", action)

	switch action {
	case BootAction:
		setHostName(script.Params, srv.Mac)
		srv.Hostname = script.Params["hostname"].(string)
		eventLog.AddEvent(event.HostBoot, srv, event.ManualBoot, script.Name, script.Params)
		return genBootScript(logger, templateRenderer, baseURL, script), nil

	case RetryAction:
		return genRetryScript(logger, baseURL, srv.Mac), nil

	case TimeoutAction:
		return timeoutScript, nil

	default:
		logger.Info("component", "polling", "msg", "Unknown action")
		return "", fmt.Errorf("%s", "Unknown action")
	}
}

func chooseManualAction(logger log.Logger, serverStates *server.States,
	eventLog *event.Log, srv server.Server) (*mappings.Script, ManualAction) {

	serverStates.Lock()
	defer serverStates.Unlock()

	if m := serverStates.Servers[srv.Mac]; m != nil {
		if m.Target != server.InitTarget {
			serverStates.DeleteServer(srv.Mac)
			logger.Debug("component", "polling", "msg", "Server boot", "mac", srv.Mac)
			return &mappings.Script{
				Name:        m.Target,
				Environment: m.Environment,
				Params:      m.Params}, BootAction
		} else if m.Retry <= maxRetry {
			m.Retry++
			m.LastAccess = int(time.Now().UTC().Unix())
			logger.Debug("component", "polling", "msg", "Retrying reboot", "mac", srv.Mac)
			return nil, RetryAction
		} else {
			serverStates.DeleteServer(srv.Mac)
			logger.Debug("component", "polling", "msg", "Timing out server", "mac", srv.Mac)
			return nil, TimeoutAction
		}
	}

	serverStates.AddServer(srv)
	logger.Debug("component", "polling", "msg", "New server", "mac", srv.Mac)
	eventLog.AddEvent(event.HostPoll, srv, "", "", nil)

	return nil, RetryAction
}

func setHostName(params map[string]interface{}, mac string) {
	if _, ok := params["hostname"]; !ok {
		hostname := utils.MacColonToDash(mac)
		if hnPrefix, ok := params["hostnamePrefix"]; ok {
			hnPrefixStr, isString := hnPrefix.(string)
			if !isString {
				hnPrefixStr = ""
			}
			params["hostname"] = hnPrefixStr + hostname
		} else {
			params["hostname"] = hostname
		}
	}
}

func GenStartScript(logger log.Logger, baseURL string) string {
	variablesMap := map[string]interface{}{}
	parsedTemplate := &bytes.Buffer{}

	tmpl, err := template.New("retry").Parse(startScript)
	if err != nil {
		logger.Info("component", "polling", "msg", "Error parsing start template")
		panic(err)
	}

	variablesMap["baseURL"] = baseURL
	err = tmpl.Execute(parsedTemplate, variablesMap)
	if err != nil {
		logger.Info("component", "polling", "msg", "Error executing start template")
		panic(err)
	}

	return parsedTemplate.String()
}

func genBootScript(logger log.Logger, templateRenderer *templates.ShoelacesTemplates, baseURL string, script *mappings.Script) string {
	script.Params["baseURL"] = utils.BaseURLforEnvName(baseURL, script.Environment)
	text, err := templateRenderer.RenderTemplate(logger, script.Name, script.Params, script.Environment)
	if err != nil {
		panic(err)
	}
	return text
}

func genRetryScript(logger log.Logger, baseURL string, mac string) string {
	variablesMap := map[string]interface{}{}
	parsedTemplate := &bytes.Buffer{}

	tmpl, err := template.New("retry").Parse(retryScript)
	if err != nil {
		logger.Info("component", "polling", "msg", "Error parsing retry template", "mac", mac)
		panic(err)
	}

	variablesMap["baseURL"] = baseURL
	variablesMap["macAddress"] = utils.MacColonToDash(mac)
	err = tmpl.Execute(parsedTemplate, variablesMap)
	if err != nil {
		logger.Info("component", "polling", "msg", "Error executing retry template", "mac", mac)
		panic(err)
	}

	return parsedTemplate.String()
}
