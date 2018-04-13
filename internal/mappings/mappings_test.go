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
	"testing"
)

var (
	mockScriptParams1 = map[string]interface{}{
		"param11": "param1_value1",
		"param21": "param2_value1",
	}
	mockScriptParams2 = map[string]interface{}{
		"param12": "param1_value2",
		"param22": "param2_value2",
	}
	mockScript1 = Script{Name: "mock_script1", Params: mockScriptParams1}
	mockScript2 = Script{Name: "mock_script2", Params: mockScriptParams2}

	mockRegex1, _ = regexp.Compile("mock_host1")
	mockRegex2, _ = regexp.Compile("mock_host2")

	mockHostNameMap1 = HostnameMap{
		Hostname: mockRegex1,
		Script:   &mockScript1,
	}

	mockHostNameMap2 = HostnameMap{
		Hostname: mockRegex2,
		Script:   &mockScript2,
	}

	_, mockNetwork1, _ = net.ParseCIDR("10.0.0.0/8")
	_, mockNetwork2, _ = net.ParseCIDR("192.168.0.0/16")

	mockNetworkMap1 = NetworkMap{
		Network: mockNetwork1,
		Script:  &mockScript1,
	}
	mockNetworkMap2 = NetworkMap{
		Network: mockNetwork2,
		Script:  &mockScript2,
	}
)

func TestScript(t *testing.T) {
	expected1 := "mock_script1 : { param11: param1_value1, param21: param2_value1 }"
	expected2 := "mock_script1 : { param21: param2_value1, param11: param1_value1 }"
	mockScriptString := mockScript1.String()
	if mockScriptString != expected1 && mockScriptString != expected2 {
		t.Errorf("Expected: %s or %s\nGot: %s\n", expected1, expected2, mockScriptString)
	}
}

func TestFindScriptForHostname(t *testing.T) {
	maps := []HostnameMap{mockHostNameMap1, mockHostNameMap2}
	script, success := FindScriptForHostname(maps, "mock_host1")
	if !(script.Name == "mock_script1" && success) {
		t.Error("Hostname should have matched")
	}
	script, success = FindScriptForHostname(maps, "mock_host2")
	if !(script.Name == "mock_script2" && success) {
		t.Error("Hostname should have matched")
	}
	script, success = FindScriptForHostname(maps, "mock_host_bad")
	if !(script == nil && !success) {
		t.Error("Hostname should have not matched")
	}
}

func TestScriptForNetwork(t *testing.T) {
	maps := []NetworkMap{mockNetworkMap1, mockNetworkMap2}
	script, success := FindScriptForNetwork(maps, "10.0.0.1")
	if !(script.Name == "mock_script1" && success) {
		t.Error("IP should have matched the network map")
	}
	script, success = FindScriptForNetwork(maps, "192.168.0.1")
	if !(script.Name == "mock_script2" && success) {
		t.Error("IP should have matched the network map")
	}
	script, success = FindScriptForNetwork(maps, "8.8.8.8")
	if !(script == nil && !success) {
		t.Error("IP shouildn't have matched the network map")
	}
}
