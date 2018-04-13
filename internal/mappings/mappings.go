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

package mappings

import (
	"net"
	"regexp"
	"strings"
)

// Script holds information related to a booting script.
type Script struct {
	Name        string
	Environment string
	Params      map[string]interface{}
}

// NetworkMap struct contains an association between a CIDR network and a
// Script.
type NetworkMap struct {
	Network *net.IPNet
	Script  *Script
}

// HostnameMap struct contains an association between a hostname regular
// expression and a Script.
type HostnameMap struct {
	Hostname *regexp.Regexp
	Script   *Script
}

// FindScriptForHostname receives a HostnameMap and a string (that can be a
// regular expression), and tries to find a match in that map. If it finds
// a match, it returns the associated script.
func FindScriptForHostname(maps []HostnameMap, hostname string) (script *Script, ok bool) {
	for _, m := range maps {
		if m.Hostname.MatchString(hostname) {
			return m.Script, true
		}
	}
	return nil, false
}

// FindScriptForNetwork receives a NetworkMap and an IP and tries to see if
// that IP belongs to any of the configured networks. If it finds a match,
// it returns the associated script.
func FindScriptForNetwork(maps []NetworkMap, ip string) (script *Script, ok bool) {
	for _, m := range maps {
		if m.Network.Contains(net.ParseIP(ip)) {
			return m.Script, true
		}
	}
	return nil, false
}

func (s Script) String() string {
	var result = s.Name + " : { "
	elems := []string{}
	if s.Environment != "" {
		elems = append(elems, "environment: "+s.Environment)
	}
	for key, value := range s.Params {
		elems = append(elems, key+": "+value.(string))
	}
	result += strings.Join(elems, ", ") + " }"

	return result
}
